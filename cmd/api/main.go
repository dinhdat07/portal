package main

import (
	"log"
	"portal-system/internal/app"
	"portal-system/internal/util"
)

// Run with: go run github.com/air-verse/air@latest
func main() {
	util.InitLogger()
	application, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
