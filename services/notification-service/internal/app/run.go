package app

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func (a *App) Run() error {
	log.Println("notification service runtime starting")

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	defer func() {
		if a.KafkaReader == nil {
			return
		}

		if err := a.KafkaReader.Close(); err != nil {
			log.Printf("failed to close kafka reader: %v", err)
			return
		}

		log.Println("kafka reader closed")
	}()

	err := a.EmailWorker.Run(ctx)
	if errors.Is(err, context.Canceled) {
		log.Println("notification service stopped by context cancellation")
		return nil
	}

	return err
}
