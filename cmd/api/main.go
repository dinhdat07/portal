package main

import (
	"log"
	"portal-system/internal/app"
	"portal-system/internal/platform/logger"
)

// Run with: go run github.com/air-verse/air@latest
func main() {
	logger.InitLogger()
	application, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
