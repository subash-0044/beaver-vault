package consensus

import (
	"fmt"

	"github.com/hashicorp/raft"
)

// RequestJoin represents the payload for joining a Raft cluster.
type RequestJoin struct {
	NodeID      string
	RaftAddress string
}

// JoinRaftHandler handles the join raft request.
func (r *Raft) JoinRaftHandler(req RequestJoin) (bool, error) {
	nodeID := req.NodeID
	raftAddr := req.RaftAddress

	if r.GetRaft().State() != raft.Leader {
		return false, fmt.Errorf("not the leader")
	}

	configFuture := r.GetRaft().GetConfiguration()
	if err := configFuture.Error(); err != nil {
		return false, fmt.Errorf("failed to get raft configuration: %w", err)
	}

	f := r.GetRaft().AddVoter(raft.ServerID(nodeID), raft.ServerAddress(raftAddr), 0, 0)
	if f.Error() != nil {
		return false, fmt.Errorf("error adding voter: %w", f.Error())
	}

	return true, nil
}
