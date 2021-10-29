package main

import (
	"bufio"
	"crypto/tls"
	"net"
	"strings"

	"go.uber.org/zap"
)

type pacman struct {
	logger   *zap.Logger
	config   *config
	registry *registry
}

func newPacman(lg *zap.Logger, cfg *config, reg *registry) pacman {
	return pacman{
		logger:   lg,
		config:   cfg,
		registry: reg,
	}
}

func (p pacman) start() error {
	var (
		listener net.Listener
		err      error
	)
	if p.config.UseMTLS {
		tlsConfig, err := p.config.tls()
		if err != nil {
			p.logger.Error("cannot load TLS config", zap.Error(err))
			return err
		}
		listener, err = tls.Listen("tcp", p.config.Listen, tlsConfig)
	} else {
		listener, err = net.Listen("tcp", p.config.Listen)
	}
	if err != nil {
		p.logger.Error("cannot listen to TCP address", zap.String("listen", p.config.Listen), zap.Error(err))
		return err
	}
	defer listener.Close()
	p.logger.Info("TCP service started", zap.String("listen", p.config.Listen))

	for {
		connection, err := listener.Accept()
		if err != nil {
			p.logger.Error("cannot open a new connection", zap.Error(err))
			continue
		}
		go p.handle(connection)
	}
}

const (
	AddPackage    = "AddPackage"
	RemovePackage = "RemovePackage"
	ListPackages  = "ListPackages"
)

func (p pacman) handle(connection net.Conn) {
	defer connection.Close()

	handler := newHandler(p.logger, p.registry, connection)
	scanner := bufio.NewScanner(connection)
	for scanner.Scan() {
		var err error
		input := scanner.Text()
		segments := strings.Split(input, " ")
		if len(segments) == 0 {
			_, err = connection.Write([]byte("ERROR: input is empty"))
		} else {
			action, args := segments[0], segments[1:]
			switch action {
			case AddPackage:
				err = handler.addPackage(args...)
			case RemovePackage:
				err = handler.removePackage(args...)
			case ListPackages:
				err = handler.listPackage()
			default:
				_, err = connection.Write([]byte("ERROR: unknown action"))
			}
		}
		if err != nil {
			p.logger.Error("cannot write TCP response", zap.Error(err))
		}
	}
}
