package cdc

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/segmentio/kafka-go"
)

type MessageHandler interface {
	Handle(ctx context.Context, value []byte) error
}

type Consumer struct {
	reader  *kafka.Reader
	handler MessageHandler
}

func NewConsumer(reader *kafka.Reader, handler MessageHandler) *Consumer {
	return &Consumer{
		reader:  reader,
		handler: handler,
	}
}

func (c *Consumer) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("[CDC Consumer] Started")

	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Println("[CDC Consumer] Stopped")
				return nil
			}

			log.Printf("[CDC Consumer] Fetch message error: %v", err)
			continue
		}

		if err := c.handler.Handle(ctx, msg.Value); err != nil {
			log.Printf(
				"[CDC Consumer] Handle message failed: topic=%s partition=%d offset=%d err=%v",
				msg.Topic,
				msg.Partition,
				msg.Offset,
				err,
			)
			// no commit, retry
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			log.Printf(
				"[CDC Consumer] Commit message failed: topic=%s partition=%d offset=%d err=%v",
				msg.Topic,
				msg.Partition,
				msg.Offset,
				err,
			)
			continue
		}
	}
}
