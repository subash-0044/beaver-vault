package bootstrap

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger/v4"
	"github.com/hashicorp/raft"

	"github.com/subash-0044/beaver-vault/pkg/config"
	"github.com/subash-0044/beaver-vault/pkg/consensus"
	"github.com/subash-0044/beaver-vault/pkg/handler"
	"github.com/subash-0044/beaver-vault/pkg/server"
	"github.com/subash-0044/beaver-vault/pkg/storage"
)

// ServerComponents holds all the components needed to run the server
type ServerComponents struct {
	Server    *server.Server
	Consensus *consensus.Raft
	Transport *raft.NetworkTransport
	DB        *badger.DB
	Cleanup   func()
}

// InitializeServer sets up all the components needed to run the server
func InitializeServer(cfg *config.Config) (*ServerComponents, error) {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(cfg.Data.Directory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %v", err)
	}

	// Initialize BadgerDB
	badgerDir := filepath.Join(cfg.Data.Directory, cfg.Raft.NodeID, "badger")
	badgerStore, err := storage.NewBadgerStore(storage.Options{
		Dir:             badgerDir,
		CreateIfMissing: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create BadgerStore: %v", err)
	}

	// Use consensus package to create Raft node
	raftNode, transport, err := consensus.NewRaftNode(consensus.RaftNodeOptions{
		NodeID:           cfg.Raft.NodeID,
		Host:             cfg.Raft.Host,
		Port:             cfg.Raft.Port,
		DataDir:          cfg.Data.Directory,
		MaxSnapshots:     cfg.Raft.MaxSnapshots,
		HeartbeatTimeout: cfg.Raft.HeartbeatTimeout,
		ElectionTimeout:  cfg.Raft.ElectionTimeout,
		CommitTimeout:    cfg.Raft.CommitTimeout,
		DB:               badgerStore.DB,
		Bootstrap:        cfg.Raft.Bootstrap,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Raft node: %v", err)
	}

	// Create handler and server
	h := handler.NewActionHandler(raftNode.GetRaft(), badgerStore.DB)
	s := server.NewGinServer(h, raftNode)

	cleanup := func() {
		if err := raftNode.GetRaft().Shutdown().Error(); err != nil {
			log.Printf("Error shutting down Raft: %v", err)
		}
		if err := transport.Close(); err != nil {
			log.Printf("Error closing transport: %v", err)
		}
		if err := badgerStore.Close(); err != nil {
			log.Printf("Error closing BadgerDB: %v", err)
		}
	}

	return &ServerComponents{
		Server:    s,
		Consensus: raftNode,
		Transport: transport,
		DB:        badgerStore.DB,
		Cleanup:   cleanup,
	}, nil
}
