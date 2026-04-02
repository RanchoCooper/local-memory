package main

import (
	"fmt"
	"log"
	"os"

	"localmemory/config"
	"localmemory/server"
)

var (
	cfgFile string
)

func main() {
	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Create and start server
	srv := server.NewServer(cfg)

	port := cfg.Server.Port
	if port <= 0 {
		port = 8080
	}

	log.Printf("LocalMemory server starting...")
	log.Printf("Listening on port: %d", port)
	log.Printf("API docs: http://localhost:%d/docs", port)

	if err := srv.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
