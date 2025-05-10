package fsm

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/dgraph-io/badger/v4"
	"github.com/hashicorp/raft"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSnapshotSink implements raft.SnapshotSink for testing
type mockSnapshotSink struct {
	*bytes.Buffer
	err error
}

func (m *mockSnapshotSink) ID() string    { return "mock-snapshot" }
func (m *mockSnapshotSink) Cancel() error { return nil }
func (m *mockSnapshotSink) Close() error  { return m.err }

func setupTestFSM(t *testing.T) (*FSM, *badger.DB, string) {
	// Create a temporary directory for the test
	tmpDir := t.TempDir()

	// Open a BadgerDB instance
	opts := badger.DefaultOptions(tmpDir)
	opts.Logger = nil // Disable logging for tests

	db, err := badger.Open(opts)
	require.NoError(t, err, "Failed to create BadgerDB")

	// Create FSM instance
	fsm := New(db).(*FSM)
	return fsm, db, tmpDir
}

func TestNew(t *testing.T) {
	fsm, db, _ := setupTestFSM(t)
	defer db.Close()

	assert.NotNil(t, fsm)
	assert.NotNil(t, fsm.parser)
}

func TestFSM_Apply(t *testing.T) {
	fsm, db, _ := setupTestFSM(t)
	defer db.Close()

	tests := []struct {
		name     string
		logType  raft.LogType
		payload  CommandPayload
		wantData interface{}
		wantErr  bool
	}{
		{
			name:    "set operation",
			logType: raft.LogCommand,
			payload: CommandPayload{
				Operation: "SET",
				Key:       "test-key",
				Value:     map[string]interface{}{"test": "value"},
			},
			wantData: map[string]interface{}{"test": "value"},
		},
		{
			name:    "get operation",
			logType: raft.LogCommand,
			payload: CommandPayload{
				Operation: "GET",
				Key:       "test-key",
			},
			wantData: map[string]interface{}{"test": "value"},
		},
		{
			name:    "delete operation",
			logType: raft.LogCommand,
			payload: CommandPayload{
				Operation: "DELETE",
				Key:       "test-key",
			},
			wantData: nil,
		},
		{
			name:    "invalid operation",
			logType: raft.LogCommand,
			payload: CommandPayload{
				Operation: "INVALID",
				Key:       "test-key",
			},
			wantData: nil,
		},
		{
			name:    "empty key",
			logType: raft.LogCommand,
			payload: CommandPayload{
				Operation: "SET",
				Key:       "",
				Value:     "test",
			},
			wantErr: true,
		},
		{
			name:     "non-command log type",
			logType:  raft.LogAddPeerDeprecated,
			payload:  CommandPayload{},
			wantData: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			log := &raft.Log{
				Type: tt.logType,
				Data: data,
			}

			result := fsm.Apply(log)
			if result == nil {
				assert.Nil(t, tt.wantData)
				return
			}

			response, ok := result.(*ApplyResponse)
			require.True(t, ok)

			if tt.wantErr {
				assert.Error(t, response.Error)
			} else {
				assert.NoError(t, response.Error)
				assert.Equal(t, tt.wantData, response.Data)
			}
		})
	}
}

func TestFSM_Snapshot(t *testing.T) {
	fsm, db, _ := setupTestFSM(t)
	defer db.Close()

	// Take snapshot
	snapshot, err := fsm.Snapshot()
	assert.NoError(t, err)
	assert.NotNil(t, snapshot)

	// Test snapshot persistence
	sink := &mockSnapshotSink{Buffer: new(bytes.Buffer)}
	err = snapshot.Persist(sink)
	assert.NoError(t, err)

	// Test snapshot release
	snapshot.Release()
}
