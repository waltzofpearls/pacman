package main

import (
	"bufio"
	"crypto/tls"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"
)

type pacman struct {
	logger   *zap.Logger
	config   *config
	registry registry
	handler  handler
	shutdown chan os.Signal
}

func newPacman(lg *zap.Logger, cfg *config, reg registry, hdl handler) pacman {
	return pacman{
		logger:   lg,
		config:   cfg,
		registry: reg,
		handler:  hdl,
		shutdown: make(chan os.Signal, 1),
	}
}

func (p pacman) listen() (net.Listener, error) {
	if p.config.UseMTLS {
		tlsConfig, err := p.config.tls()
		if err != nil {
			p.logger.Error("cannot load TLS config", zap.Error(err))
			return nil, err
		}
		return tls.Listen("tcp", p.config.Listen, tlsConfig)
	}
	return net.Listen("tcp", p.config.Listen)
}

func (p pacman) serve(listener net.Listener) {
	listenField := zap.String("listen", p.config.Listen)
	p.logger.Info("TCP service started", listenField)

	defer func() {
		_ = listener.Close()
		p.logger.Info("stopped listening TCP address", listenField)
	}()

	signal.Notify(p.shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		var tempDelay time.Duration // how long to sleep on accept failure
		for {
			select {
			case <-p.shutdown:
				return
			default:
			}
			connection, err := listener.Accept()
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Temporary() {
					if tempDelay == 0 {
						tempDelay = 5 * time.Millisecond
					} else {
						tempDelay *= 2
					}
					if max := 1 * time.Second; tempDelay > max {
						tempDelay = max
					}
					p.logger.Error("cannot accept connection", zap.Stringer("retry_in", tempDelay), zap.Error(err))
					time.Sleep(tempDelay)
					continue
				}
				p.shutdown <- os.Interrupt
				return
			}
			tempDelay = 0

			go p.handle(connection)
		}
	}()

	<-p.shutdown
}

const (
	AddPackage    = "AddPackage"
	RemovePackage = "RemovePackage"
	ListPackages  = "ListPackages"

	MaxLineLenBytes  = 1024
	ReadWriteTimeout = time.Minute
)

func (p pacman) handle(connection net.Conn) {
	addrField := zap.Stringer("remote_addr", connection.RemoteAddr())
	p.logger.Info("accepted TCP connection", addrField)

	defer func() {
		_ = connection.Close()
		p.logger.Info("closed TCP connection", addrField)
	}()

	done := make(chan bool)
	_ = connection.SetReadDeadline(time.Now().Add(ReadWriteTimeout))

	go func() {
		limited := &io.LimitedReader{
			R: connection,
			N: MaxLineLenBytes,
		}
		scanner := bufio.NewScanner(limited)
		for scanner.Scan() {
			var err error
			input := scanner.Text()
			segments := strings.Split(input, " ")
			if len(segments) == 0 {
				_, err = connection.Write([]byte("\nERROR: input is empty\n"))
			} else {
				action, args := segments[0], segments[1:]
				switch action {
				case AddPackage:
					err = p.handler.addPackage(connection, args...)
				case RemovePackage:
					err = p.handler.removePackage(connection, args...)
				case ListPackages:
					err = p.handler.listPackages(connection)
				default:
					_, err = connection.Write([]byte("\nERROR: unknown action\n"))
				}
			}
			if err != nil {
				p.logger.Error("cannot write TCP response", zap.Error(err))
				break
			}
			// reset remaining bytes left in the LimitReader
			limited.N = MaxLineLenBytes
			// reset read deadline
			_ = connection.SetReadDeadline(time.Now().Add(ReadWriteTimeout))
		}
		done <- true
	}()

	<-done
}
