package consensus

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/hashicorp/raft"

	"github.com/subash-0044/beaver-vault/pkg/fsm"
)

// RaftNodeOptions holds all options needed to create a Raft node
type RaftNodeOptions struct {
	NodeID           string
	Host             string
	Port             int
	DataDir          string
	MaxSnapshots     int
	HeartbeatTimeout string
	ElectionTimeout  string
	CommitTimeout    string
	DB               *badger.DB
	Bootstrap        bool
}

// NewRaftNode initializes and returns a consensus.Raft and the underlying transport
func NewRaftNode(opts RaftNodeOptions) (*Raft, *raft.NetworkTransport, error) {
	// Create Raft configuration
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(opts.NodeID)
	var err error
	raftConfig.HeartbeatTimeout, err = time.ParseDuration(opts.HeartbeatTimeout)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid HeartbeatTimeout: %v", err)
	}
	raftConfig.ElectionTimeout, err = time.ParseDuration(opts.ElectionTimeout)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid ElectionTimeout: %v", err)
	}
	raftConfig.CommitTimeout, err = time.ParseDuration(opts.CommitTimeout)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid CommitTimeout: %v", err)
	}

	// Create Raft storage
	raftDir := filepath.Join(opts.DataDir, "raft")
	if mkdirErr := os.MkdirAll(raftDir, 0755); mkdirErr != nil {
		return nil, nil, fmt.Errorf("failed to create Raft directory: %v", mkdirErr)
	}

	logStore := raft.NewInmemStore()
	stableStore := raft.NewInmemStore()
	snapshotStore, err := raft.NewFileSnapshotStore(raftDir, opts.MaxSnapshots, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create snapshot store: %v", err)
	}

	addr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	transport, err := raft.NewTCPTransport(addr, nil, 3, 10*time.Second, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Raft transport: %v", err)
	}

	fsmStore := fsm.New(opts.DB)

	r, err := raft.NewRaft(raftConfig, fsmStore, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Raft: %v", err)
	}

	// Bootstrap the cluster if configured
	if opts.Bootstrap {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raftConfig.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		r.BootstrapCluster(configuration)
	}

	return NewRaftObj(r), transport, nil
}
