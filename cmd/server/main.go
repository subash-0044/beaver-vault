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
	nodeID := flag.String("node-id", "", "node ID for this instance")
	httpPort := flag.Int("http-port", 0, "HTTP port for this instance")
	raftPort := flag.Int("raft-port", 0, "Raft port for this instance")
	raftHost := flag.String("raft-host", "", "Raft host for this instance")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Override config values if provided via command line
	if *nodeID != "" {
		cfg.Raft.NodeID = *nodeID
	}
	if *httpPort != 0 {
		cfg.Server.Port = *httpPort
	}
	if *raftPort != 0 {
		cfg.Raft.Port = *raftPort
	}
	if *raftHost != "" {
		cfg.Raft.Host = *raftHost
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
