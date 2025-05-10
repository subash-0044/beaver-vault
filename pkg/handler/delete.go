package handler

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/raft"

	"github.com/subash-0044/beaver-vault/pkg/fsm"
)

// Delete removes data from the Raft cluster.
// The operation is applied to the Raft cluster and acknowledged by a quorum.
// This method must be executed on the Raft leader; otherwise, it returns an error.
func (h Handler) Delete(key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("key is empty")
	}

	if h.raft.State() != raft.Leader {
		return fmt.Errorf("not the leader")
	}

	payload := fsm.CommandPayload{
		Operation: "DELETE",
		Key:       key,
		Value:     nil,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error preparing remove data payload: %s", err.Error())
	}

	applyFuture := h.raft.Apply(data, 500*time.Millisecond)
	if err := applyFuture.Error(); err != nil {
		return fmt.Errorf("error removing data in raft cluster: %s", err.Error())
	}

	_, ok := applyFuture.Response().(*fsm.ApplyResponse)
	if !ok {
		return fmt.Errorf("error response is not match apply response")
	}

	return nil
}
