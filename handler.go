package main

import (
	"fmt"
	"net"

	"go.uber.org/zap"
)

type handler struct {
	logger     *zap.Logger
	registry   *registry
	connection net.Conn
}

func newHandler(lg *zap.Logger, reg *registry, conn net.Conn) handler {
	return handler{
		logger:     lg,
		registry:   reg,
		connection: conn,
	}
}

func (h handler) addPackage(args ...string) error {
	if len(args) == 0 {
		_, err := h.connection.Write([]byte("\nERROR: no package name\n"))
		return err
	}
	name, deps := args[0], args[1:]
	if err := h.registry.add(name, deps); err != nil {
		_, err = h.connection.Write([]byte(fmt.Sprintf("\nERROR: failed adding package: %s\n", err)))
		return err
	}
	_, err := h.connection.Write([]byte("\nPackage added\n"))
	return err
}

func (h handler) removePackage(args ...string) error {
	if len(args) == 0 {
		_, err := h.connection.Write([]byte("\nERROR: no package name\n"))
		return err
	}
	if err := h.registry.remove(args[0]); err != nil {
		_, err = h.connection.Write([]byte(fmt.Sprintf("\nERROR: failed removing package: %s\n", err)))
		return err
	}
	_, err := h.connection.Write([]byte("\nPackage removed\n"))
	return err
}

func (h handler) listPackage() error {
	_, err := h.connection.Write([]byte("\n" + h.registry.list() + "\n"))
	return err
}
