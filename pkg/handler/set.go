package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/raft"

	"github.com/subash-0044/beaver-vault/pkg/fsm"
)

// RequestStore represents the payload for storing new data in the Raft cluster.
type RequestStore struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// Store handles saving data to the Raft cluster.
// It invokes raft.Apply to store the data across the cluster with acknowledgment from a quorum.
// This operation must be performed on the Raft leader.
func (h Handler) Store(_ context.Context, form RequestStore) error {
	form.Key = strings.TrimSpace(form.Key)
	if form.Key == "" {
		return fmt.Errorf("key is empty")
	}

	if h.raft.State() != raft.Leader {
		return fmt.Errorf("not the leader")
	}

	payload := fsm.CommandPayload{
		Operation: "SET",
		Key:       form.Key,
		Value:     form.Value,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error preparing saving data payload: %s", err.Error())
	}

	applyFuture := h.raft.Apply(data, 500*time.Millisecond)
	if err := applyFuture.Error(); err != nil {
		return fmt.Errorf("error persisting data in raft cluster: %s", err.Error())
	}

	_, ok := applyFuture.Response().(*fsm.ApplyResponse)
	if !ok {
		return fmt.Errorf("response does not match apply response")
	}

	return nil
}
