package main

import (
	"log"

	"github.com/alexuserid/go-codevis/internal/backend"
)

func main() {
	if err := backend.Run(); err != nil {
		log.Fatalf("run app failed: %s", err)
	}
}
