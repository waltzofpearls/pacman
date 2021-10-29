package main

import (
	"log"

	"go.uber.org/zap"
)

func main() {
	logger, err := newLogger()
	if err != nil {
		log.Fatal("cannot create logger")
	}
	defer logger.Sync()

	config, err := newConfig()
	if err != nil {
		logger.Fatal("cannot read env configs", zap.Error(err))
	}

	registry := newRegistry()

	pacman := newPacman(logger, config, registry)
	if err := pacman.start(); err != nil {
		logger.Fatal("cannot start TCP service", zap.Error(err))
	}
}
