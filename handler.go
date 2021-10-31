package main

import (
	"fmt"
	"net"

	"go.uber.org/zap"
)

type handler interface {
	addPackage(connection net.Conn, args ...string) error
	removePackage(connection net.Conn, args ...string) error
	listPackages(connection net.Conn) error
}

type action struct {
	logger   *zap.Logger
	registry registry
}

func newAction(lg *zap.Logger, reg registry) action {
	return action{
		logger:   lg,
		registry: reg,
	}
}

func (a action) addPackage(connection net.Conn, args ...string) error {
	if len(args) == 0 {
		_, err := connection.Write([]byte("\nERROR: no package name\n"))
		return err
	}
	name, deps := args[0], args[1:]
	if err := a.registry.add(name, deps); err != nil {
		_, err = connection.Write([]byte(fmt.Sprintf("\nERROR: failed adding package: %s\n", err)))
		return err
	}
	_, err := connection.Write([]byte("\nPackage added\n"))
	return err
}

func (a action) removePackage(connection net.Conn, args ...string) error {
	if len(args) == 0 {
		_, err := connection.Write([]byte("\nERROR: no package name\n"))
		return err
	}
	if err := a.registry.remove(args[0]); err != nil {
		_, err = connection.Write([]byte(fmt.Sprintf("\nERROR: failed removing package: %s\n", err)))
		return err
	}
	_, err := connection.Write([]byte("\nPackage removed\n"))
	return err
}

func (a action) listPackages(connection net.Conn) error {
	_, err := connection.Write([]byte("\n" + a.registry.list() + "\n"))
	return err
}
