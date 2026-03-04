package main

import (
	"log"

	"github.com/balub/apisto/internal/config"
	"github.com/balub/apisto/internal/server"
)

func main() {
	cfg := config.Load()
	log.Printf("apisto: starting on port %s", cfg.Port)
	if err := server.Run(cfg); err != nil {
		log.Fatalf("apisto: fatal error: %v", err)
	}
}
