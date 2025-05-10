package main

import (
	"flag"
	"log"

	"github.com/subash-0044/beaver-vault/pkg/bootstrap"
	"github.com/subash-0044/beaver-vault/pkg/config"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config/config.yaml", "path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize server components
	components, err := bootstrap.InitializeServer(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}
	defer components.Cleanup()

	// Start server
	log.Printf("Starting server on %s", cfg.Server.GetHTTPAddress())
	if err := components.Server.Run(cfg.Server.GetHTTPAddress()); err != nil {
		defer components.Cleanup()
		log.Fatalf("Server failed: %v", err) //nolint:gocritic
	}
}
