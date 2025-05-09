package storage

import "errors"

// Common errors
var (
	ErrKeyNotFound = errors.New("key not found")
)

type Value struct {
	Data []byte
}

type Storage interface {
	Get(key []byte) (*Value, error)
	Put(key, value []byte) error
	Delete(key []byte) error
	Close() error
}

// Options configures the storage engine
type Options struct {
	// Directory where the database files will be stored
	Dir string
	// Whether to create the directory if it doesn't exist
	CreateIfMissing bool
}
