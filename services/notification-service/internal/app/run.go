package app

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
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

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return a.EmailWorker.Run(gCtx)
	})

	g.Go(func() error {
		return a.RetryWorker.Run(gCtx)
	})

	err := g.Wait()
	if err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	log.Println("notification service stopped cleanly")
	return nil
}
