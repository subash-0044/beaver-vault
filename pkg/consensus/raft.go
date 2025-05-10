package consensus

import (
	"github.com/hashicorp/raft"
)

// handler struct handler
type Raft struct {
	raft *raft.Raft
}

func NewRaftObj(raft *raft.Raft) *Raft {
	return &Raft{
		raft: raft,
	}
}

// StatsRaftHandler get raft status
func (r Raft) StatsRaftHandler() (map[string]string, error) {
	return r.raft.Stats(), nil
}
