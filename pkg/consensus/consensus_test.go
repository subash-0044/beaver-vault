package consensus

import (
	"encoding/json"
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
func setupTestRaft(t *testing.T, nodeID string) (*Raft, string, raft.ServerAddress) {
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

	// Create our real FSM
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

	// Create our consensus.Raft wrapper
	r := NewRaftObj(ra)

	return r, tmpDir, addr
}

func TestRaftWithRealNodes(t *testing.T) {
	// Create leader node
	leader, leaderDir, _ := setupTestRaft(t, "node1")
	defer os.RemoveAll(leaderDir)

	// Wait for leader to be elected
	timeout := time.Now().Add(3 * time.Second)
	for time.Now().Before(timeout) {
		if leader.GetRaft().State() == raft.Leader {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	assert.Equal(t, raft.Leader, leader.GetRaft().State(), "Node1 should become leader")

	// Test stats
	stats, err := leader.StatsRaftHandler()
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "Leader", stats["state"])

	// Create follower node
	follower, followerDir, followerAddr := setupTestRaft(t, "node2")
	defer os.RemoveAll(followerDir)

	// Join follower to leader
	req := RequestJoin{
		NodeID:      "node2",
		RaftAddress: string(followerAddr),
	}
	success, err := leader.JoinRaftHandler(req)
	assert.NoError(t, err)
	assert.True(t, success)

	// Wait for follower to be added
	timeout = time.Now().Add(3 * time.Second)
	for time.Now().Before(timeout) {
		if follower.GetRaft().State() == raft.Follower {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	assert.Equal(t, raft.Follower, follower.GetRaft().State(), "Node2 should become follower")

	// Test FSM operations using our CommandPayload format
	cmd := fsm.CommandPayload{
		Operation: "SET",
		Key:       "test-key",
		Value:     "test-value",
	}
	data, err := json.Marshal(cmd)
	assert.NoError(t, err)

	// Apply the command through leader
	future := leader.GetRaft().Apply(data, 5*time.Second)
	assert.NoError(t, future.Error())

	// Wait for replication
	time.Sleep(1 * time.Second)

	// Read the value through leader
	cmd = fsm.CommandPayload{
		Operation: "GET",
		Key:       "test-key",
	}
	data, err = json.Marshal(cmd)
	assert.NoError(t, err)

	future = leader.GetRaft().Apply(data, 5*time.Second)
	assert.NoError(t, future.Error())
	response := future.Response().(*fsm.ApplyResponse)
	assert.NoError(t, response.Error)
	assert.Equal(t, "test-value", response.Data)

	// Delete the value
	cmd = fsm.CommandPayload{
		Operation: "DELETE",
		Key:       "test-key",
	}
	data, err = json.Marshal(cmd)
	assert.NoError(t, err)

	future = leader.GetRaft().Apply(data, 5*time.Second)
	assert.NoError(t, future.Error())

	// Wait for replication
	time.Sleep(1 * time.Second)

	// Verify deletion
	cmd = fsm.CommandPayload{
		Operation: "GET",
		Key:       "test-key",
	}
	data, err = json.Marshal(cmd)
	assert.NoError(t, err)

	future = leader.GetRaft().Apply(data, 5*time.Second)
	assert.NoError(t, future.Error())
	response = future.Response().(*fsm.ApplyResponse)
	assert.NoError(t, response.Error)
	assert.Equal(t, make(map[string]interface{}), response.Data)

	// Test dropping the follower
	dropReq := RequestDrop{
		NodeID: "node2",
	}
	success, err = leader.DropRaftHandler(dropReq)
	assert.NoError(t, err)
	assert.True(t, success)

	// Verify the node was dropped
	timeout = time.Now().Add(3 * time.Second)
	for time.Now().Before(timeout) {
		cfg := leader.GetRaft().GetConfiguration()
		if len(cfg.Configuration().Servers) == 1 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	cfg := leader.GetRaft().GetConfiguration()
	assert.Equal(t, 1, len(cfg.Configuration().Servers), "Cluster should have only one node after drop")
}

func TestRaftLeaderDrop(t *testing.T) {
	// Create initial leader node
	leader, leaderDir, _ := setupTestRaft(t, "node1")
	defer os.RemoveAll(leaderDir)

	// Wait for leader to be elected
	timeout := time.Now().Add(3 * time.Second)
	for time.Now().Before(timeout) {
		if leader.GetRaft().State() == raft.Leader {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	assert.Equal(t, raft.Leader, leader.GetRaft().State(), "Node1 should become leader")

	// Create two follower nodes
	follower1, follower1Dir, follower1Addr := setupTestRaft(t, "node2")
	defer os.RemoveAll(follower1Dir)
	follower2, follower2Dir, follower2Addr := setupTestRaft(t, "node3")
	defer os.RemoveAll(follower2Dir)

	// Join both followers to the cluster
	req1 := RequestJoin{
		NodeID:      "node2",
		RaftAddress: string(follower1Addr),
	}
	success, err := leader.JoinRaftHandler(req1)
	assert.NoError(t, err)
	assert.True(t, success)

	req2 := RequestJoin{
		NodeID:      "node3",
		RaftAddress: string(follower2Addr),
	}
	success, err = leader.JoinRaftHandler(req2)
	assert.NoError(t, err)
	assert.True(t, success)

	// Wait for followers to be added and synced
	time.Sleep(2 * time.Second)

	// Verify cluster size
	cfg := leader.GetRaft().GetConfiguration()
	assert.Equal(t, 3, len(cfg.Configuration().Servers), "Cluster should have three nodes")

	// Drop the leader node
	dropReq := RequestDrop{
		NodeID: "node1",
	}
	success, err = leader.DropRaftHandler(dropReq)
	assert.NoError(t, err)
	assert.True(t, success)

	// Wait for new leader election and node removal
	time.Sleep(2 * time.Second)

	// Verify that one of the followers became leader
	assert.True(t, follower1.GetRaft().State() == raft.Leader || follower2.GetRaft().State() == raft.Leader,
		"One of the followers should become leader")

	// Get the new leader
	var newLeader *Raft
	if follower1.GetRaft().State() == raft.Leader {
		newLeader = follower1
	} else {
		newLeader = follower2
	}

	// Verify cluster size is now 2
	cfg = newLeader.GetRaft().GetConfiguration()
	assert.Equal(t, 2, len(cfg.Configuration().Servers), "Cluster should have two nodes after leader removal")

	// Verify the old leader was actually removed
	servers := cfg.Configuration().Servers
	for _, server := range servers {
		assert.NotEqual(t, raft.ServerID("node1"), server.ID, "Old leader should not be in the configuration")
	}
}
