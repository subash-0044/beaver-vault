package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
	"github.com/stretchr/testify/assert"
	"github.com/subash-0044/beaver-vault/pkg/fsm"
	"github.com/subash-0044/beaver-vault/pkg/handler"
)

func setupTestServer(t *testing.T) (*Server, string, func()) {
	tmpDir, err := os.MkdirTemp("", "raft-test-server")
	assert.NoError(t, err)

	badgerOpts := badger.DefaultOptions(filepath.Join(tmpDir, "badger"))
	badgerOpts.Logger = nil
	db, err := badger.Open(badgerOpts)
	assert.NoError(t, err)

	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID("node1")
	config.HeartbeatTimeout = 100 * time.Millisecond
	config.ElectionTimeout = 100 * time.Millisecond
	config.LeaderLeaseTimeout = 100 * time.Millisecond
	config.CommitTimeout = 10 * time.Millisecond
	config.ShutdownOnRemove = true
	config.SnapshotInterval = 10 * time.Second
	config.SnapshotThreshold = 100

	logStore := raft.NewInmemStore()
	stableStore := raft.NewInmemStore()

	snapshotStore, err := raft.NewFileSnapshotStore(tmpDir, 1, nil)
	assert.NoError(t, err)

	transport, err := raft.NewTCPTransport("localhost:0", nil, 3, 10*time.Second, nil)
	assert.NoError(t, err)

	fsmStore := fsm.New(db)

	ra, err := raft.NewRaft(config, fsmStore, logStore, stableStore, snapshotStore, transport)
	assert.NoError(t, err)

	configuration := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      config.LocalID,
				Address: transport.LocalAddr(),
			},
		},
	}
	ra.BootstrapCluster(configuration)

	// Wait for leader election with timeout
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("Timeout waiting for leader election")
		case <-ticker.C:
			if ra.State() == raft.Leader {
				goto leaderElected
			}
		}
	}
leaderElected:

	h := handler.NewActionHandler(ra, db)

	// Create a test-specific server without template loading
	s := &Server{
		handler:   h,
		consensus: nil,
		router:    gin.New(),
	}
	s.setupRoutes()

	cleanup := func() {
		if err := ra.Shutdown().Error(); err != nil {
			t.Logf("Error shutting down Raft: %v", err)
		}

		if err := transport.Close(); err != nil {
			t.Logf("Error closing transport: %v", err)
		}

		if err := db.Close(); err != nil {
			t.Logf("Error closing BadgerDB: %v", err)
		}

		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Error removing temp directory: %v", err)
		}
	}

	return s, tmpDir, cleanup
}

func TestHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s, _, cleanup := setupTestServer(t)
	defer cleanup()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	s.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

func TestKeyValueOperations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s, _, cleanup := setupTestServer(t)
	defer cleanup()

	// Wait for system to stabilize
	time.Sleep(200 * time.Millisecond)

	t.Run("Set Value", func(t *testing.T) {
		value := map[string]interface{}{
			"name": "test",
			"age":  30,
		}
		body, _ := json.Marshal(value)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/v1/kv/test-key", bytes.NewBuffer(body))
		s.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "ok", response["status"])
	})

	// Wait for write to be committed
	time.Sleep(200 * time.Millisecond)

	t.Run("Get Value", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/kv/test-key", nil)
		s.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "test-key", response["key"])
		value := response["value"].(map[string]interface{})
		assert.Equal(t, "test", value["name"])
		assert.Equal(t, float64(30), value["age"])
	})

	t.Run("Get Non-existent Key", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/kv/non-existent", nil)
		s.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "key not found", response["error"])
	})

	t.Run("Delete Value", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/v1/kv/test-key", nil)
		s.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "ok", response["status"])

		// Wait for deletion to be committed
		time.Sleep(200 * time.Millisecond)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/api/v1/kv/test-key", nil)
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
