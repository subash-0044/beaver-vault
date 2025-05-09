package storage

import (
	"context"
)

type Value struct {
	Data []byte 
}

type Storage interface {
	Get(ctx context.Context, key []byte) (*Value, error)
	Put(ctx context.Context, key, value []byte) error
	Delete(ctx context.Context, key []byte) error
	Close() error
}

// Options configures the storage engine
type Options struct {
	// Directory where the database files will be stored
	Dir string
	// Whether to create the directory if it doesn't exist
	CreateIfMissing bool
}
