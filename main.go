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

	store := newInMemoryStore()
	action := newAction(logger, store)
	pacman := newPacman(logger, config, store, action)
	listener, err := pacman.listen()
	if err != nil {
		logger.Fatal("cannot listen to TCP address", zap.String("listen", config.Listen), zap.Error(err))
	}
	pacman.serve(listener)
}
