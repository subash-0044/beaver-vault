package consensus

import (
	"fmt"

	"github.com/hashicorp/raft"
)

// RequestDrop represents the payload for removing a node from the Raft cluster.
type RequestDrop struct {
	NodeID string
}

func (r *Raft) DropRaftHandler(form RequestDrop) (bool, error) {
	nodeID := form.NodeID

	if r.raft.State() != raft.Leader {
		return false, fmt.Errorf("not the leader")
	}

	configFuture := r.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		return false, fmt.Errorf("failed to get raft configuration: %w", err)
	}

	future := r.raft.RemoveServer(raft.ServerID(nodeID), 0, 0)
	if err := future.Error(); err != nil {
		return false, fmt.Errorf("error removing existing node %s: %w", nodeID, err)
	}

	return true, nil
}
