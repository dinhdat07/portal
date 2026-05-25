package main

import (
	"log"

	cdcinfra "portal-system/internal/infrastructure/cdc"
)

func main() {
	consumer, err := cdcinfra.New()
	if err != nil {
		log.Fatalf("init cdc consumer: %v", err)
	}

	if err := consumer.Run(); err != nil {
		log.Fatalf("run cdc consumer: %v", err)
	}
}
