package handler

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/hashicorp/raft"
	"github.com/stretchr/testify/assert"
	"github.com/subash-0044/beaver-vault/pkg/fsm"
)

// setupTestRaft creates a new Raft node for testing
func setupTestRaft(t *testing.T, nodeID string) (*raft.Raft, *badger.DB, string, raft.ServerAddress) {
	// Create a temporary directory for Raft data
	tmpDir, err := ioutil.TempDir("", "raft-test-"+nodeID)
	assert.NoError(t, err)

	// Create BadgerDB for FSM
	badgerOpts := badger.DefaultOptions(filepath.Join(tmpDir, "badger"))
	badgerOpts.Logger = nil
	db, err := badger.Open(badgerOpts)
	assert.NoError(t, err)

	// Create Raft configuration
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(nodeID)
	config.HeartbeatTimeout = 50 * time.Millisecond
	config.ElectionTimeout = 50 * time.Millisecond
	config.LeaderLeaseTimeout = 50 * time.Millisecond
	config.CommitTimeout = 5 * time.Millisecond

	// Create the log store and stable store
	logStore := raft.NewInmemStore()
	stableStore := raft.NewInmemStore()

	// Create the snapshot store
	snapshotStore, err := raft.NewFileSnapshotStore(tmpDir, 1, nil)
	assert.NoError(t, err)

	// Create the transport
	transport, err := raft.NewTCPTransport("localhost:0", nil, 3, 10*time.Second, nil)
	assert.NoError(t, err)
	addr := transport.LocalAddr()

	// Create our FSM
	fsmStore := fsm.New(db)

	// Create and bootstrap the Raft node
	ra, err := raft.NewRaft(config, fsmStore, logStore, stableStore, snapshotStore, transport)
	assert.NoError(t, err)

	if nodeID == "node1" {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: addr,
				},
			},
		}
		ra.BootstrapCluster(configuration)
	}

	return ra, db, tmpDir, addr
}

func TestHandlerWithRealRaft(t *testing.T) {
	// Create leader node
	raftNode, db, tmpDir, _ := setupTestRaft(t, "node1")
	defer os.RemoveAll(tmpDir)
	defer db.Close()

	// Wait for leader to be elected
	timeout := time.Now().Add(3 * time.Second)
	for time.Now().Before(timeout) {
		if raftNode.State() == raft.Leader {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	assert.Equal(t, raft.Leader, raftNode.State(), "Node1 should become leader")

	// Create handler
	h := NewActionHandler(raftNode, db)

	// Test Store operation
	t.Run("Store", func(t *testing.T) {
		err := h.Store(context.Background(), RequestStore{
			Key:   "test-key",
			Value: "test-value",
		})
		assert.NoError(t, err)

		// Wait for replication
		time.Sleep(100 * time.Millisecond)
	})

	// Test Get operation
	t.Run("Get", func(t *testing.T) {
		value, err := h.Get("test-key")
		assert.NoError(t, err)
		assert.Equal(t, "test-value", value)

		// Test non-existent key
		value, err = h.Get("non-existent")
		assert.NoError(t, err)
		assert.Nil(t, value)
	})

	// Test Delete operation
	t.Run("Delete", func(t *testing.T) {
		err := h.Delete("test-key")
		assert.NoError(t, err)

		// Wait for deletion to be applied
		time.Sleep(100 * time.Millisecond)

		// Verify deletion
		value, err := h.Get("test-key")
		assert.NoError(t, err)
		assert.Nil(t, value)
	})

	// Test operations with follower
	t.Run("Follower Operations", func(t *testing.T) {
		followerRaft, followerDB, followerDir, followerAddr := setupTestRaft(t, "node2")
		defer os.RemoveAll(followerDir)
		defer followerDB.Close()

		// Join follower to cluster
		future := raftNode.AddVoter(
			raft.ServerID("node2"),
			followerAddr,
			0,
			0,
		)
		assert.NoError(t, future.Error())

		// Wait for follower to be added
		time.Sleep(1 * time.Second)

		followerHandler := NewActionHandler(followerRaft, followerDB)

		// Store should fail on follower
		err := followerHandler.Store(context.Background(), RequestStore{
			Key:   "follower-key",
			Value: "follower-value",
		})
		assert.EqualError(t, err, "not the leader")

		// Get should work on follower (after leader writes)
		err = h.Store(context.Background(), RequestStore{
			Key:   "replicated-key",
			Value: "replicated-value",
		})
		assert.NoError(t, err)

		// Wait for replication
		time.Sleep(1 * time.Second)

		value, err := followerHandler.Get("replicated-key")
		assert.NoError(t, err)
		assert.Equal(t, "replicated-value", value)

		// Delete should fail on follower
		err = followerHandler.Delete("replicated-key")
		assert.EqualError(t, err, "not the leader")
	})

	// Test error cases
	t.Run("Error Cases", func(t *testing.T) {
		// Empty key
		err := h.Store(context.Background(), RequestStore{
			Key:   "",
			Value: "test",
		})
		assert.EqualError(t, err, "key is empty")

		_, err = h.Get("")
		assert.EqualError(t, err, "key is empty")

		err = h.Delete("")
		assert.EqualError(t, err, "key is empty")

		// Non-existent key
		value, err := h.Get("non-existent")
		assert.NoError(t, err)
		assert.Nil(t, value)
	})
}
