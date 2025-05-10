package fsm

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/subash-0044/beaver-vault/pkg/parser"
	"github.com/subash-0044/beaver-vault/pkg/storage"

	"github.com/dgraph-io/badger/v4"
	"github.com/hashicorp/raft"
)

// CommandPayload is payload sent by system when calling raft.Apply(cmd []byte, timeout time.Duration)
type CommandPayload struct {
	Operation string
	Key       string
	Value     interface{}
}

// ApplyResponse response from Apply raft
type ApplyResponse struct {
	Error error
	Data  interface{}
}

// FSM implements raft.FSM using badgerDB
type FSM struct {
	parser *parser.Parser
}

// Apply log is invoked once a log entry is committed.
// It returns a value which will be made available in the
// ApplyFuture returned by Raft.Apply method if that
// method was called on the same Raft node as the FSM.
func (f FSM) Apply(log *raft.Log) interface{} {
	switch log.Type {
	case raft.LogCommand:
		var payload = CommandPayload{}
		if err := f.parser.UnmarshalTo(log.Data, &payload); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error unmarshalling store payload %s\n", err.Error())
			return nil
		}

		op := strings.ToUpper(strings.TrimSpace(payload.Operation))
		switch op {
		case "SET":
			err := f.parser.Put(payload.Key, payload.Value)
			return &ApplyResponse{
				Error: err,
				Data:  payload.Value,
			}
		case "GET":
			value, err := f.parser.Get(payload.Key)
			var data interface{}
			if err == nil && value != nil {
				data = value.Data
			} else {
				data = make(map[string]interface{})
			}
			return &ApplyResponse{
				Error: err,
				Data:  data,
			}
		case "DELETE":
			return &ApplyResponse{
				Error: f.parser.Delete(payload.Key),
				Data:  nil,
			}
		}
	}

	_, _ = fmt.Fprintf(os.Stderr, "not raft log command type\n")
	return nil
}

// Snapshot will be called during make snapshot.
// Snapshot is used to support log compaction.
// No need to call snapshot since it already persisted in disk (using BadgerDB) when raft calling Apply function.
func (f FSM) Snapshot() (raft.FSMSnapshot, error) {
	return newSnapshotNoop()
}

// Restore is used to restore an FSM from a Snapshot. It is not called
// concurrently with any other command. The FSM must discard all previous
// state.
// Restore will update all data in BadgerDB
func (f FSM) Restore(rClose io.ReadCloser) error {
	defer func() {
		if err := rClose.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stdout, "[FINALLY RESTORE] close error %s\n", err.Error())
		}
	}()

	_, _ = fmt.Fprintf(os.Stdout, "[START RESTORE] read all message from snapshot\n")
	var totalRestored int

	decoder := json.NewDecoder(rClose)
	for decoder.More() {
		var data = &CommandPayload{}
		err := decoder.Decode(data)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stdout, "[END RESTORE] error decode data %s\n", err.Error())
			return err
		}

		if err := f.parser.Put(data.Key, data.Value); err != nil {
			_, _ = fmt.Fprintf(os.Stdout, "[END RESTORE] error persist data %s\n", err.Error())
			return err
		}

		totalRestored++
	}

	// read closing bracket
	_, err := decoder.Token()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stdout, "[END RESTORE] error %s\n", err.Error())
		return err
	}

	_, _ = fmt.Fprintf(os.Stdout, "[END RESTORE] success restore %d messages in snapshot\n", totalRestored)
	return nil
}

// New creates a new raft.FSM implementation using badgerDB
func New(badgerDB *badger.DB) raft.FSM {
	store := &storage.BadgerStore{DB: badgerDB}
	return &FSM{
		parser: parser.NewParser(store),
	}
}
