package main

import (
	"log"

	"portal-system/internal/composer"
)

// Run with: go run github.com/air-verse/air@latest
func main() {
	application, err := composer.Composer()
	if err != nil {
		log.Fatal(err)
	}

	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
