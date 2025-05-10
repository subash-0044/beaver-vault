package handler

import (
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/hashicorp/raft"
)

// RaftNode represents the minimal Raft interface needed by Handler
type RaftNode interface {
	Apply([]byte, time.Duration) raft.ApplyFuture
	State() raft.RaftState
}

// DB represents the minimal BadgerDB interface needed by Handler
type DB interface {
	NewTransaction(bool) *badger.Txn
}

type Handler struct {
	raft RaftNode
	db   DB
}

func NewActionHandler(raft RaftNode, db DB) *Handler {
	return &Handler{
		raft: raft,
		db:   db,
	}
}
