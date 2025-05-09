package storage

import (
	"context"
	"fmt"
	"os"

	"github.com/dgraph-io/badger/v4"
)

type BadgerStore struct {
	db *badger.DB
}

// NewBadgerStore creates a new BadgerDB storage instance
func NewBadgerStore(opts Options) (*BadgerStore, error) {
	// Ensure the directory exists
	if opts.CreateIfMissing {
		if err := os.MkdirAll(opts.Dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}
	}

	badgerOpts := badger.DefaultOptions(opts.Dir)

	db, err := badger.Open(badgerOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger database: %w", err)
	}

	return &BadgerStore{db: db}, nil
}

// Get retrieves a value for a given key
func (b *BadgerStore) Get(ctx context.Context, key []byte) (*Value, error) {
	var value *Value
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			return nil
		}
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			value = &Value{Data: val}
			return nil
		})
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get value: %w", err)
	}
	return value, nil
}

// Put stores a value for a given key
func (b *BadgerStore) Put(ctx context.Context, key, data []byte) error {
	err := b.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, data)
	})

	if err != nil {
		return fmt.Errorf("failed to put value: %w", err)
	}
	return nil
}

// Delete removes a key-value pair
func (b *BadgerStore) Delete(ctx context.Context, key []byte) error {
	err := b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})

	if err != nil {
		return fmt.Errorf("failed to delete key: %w", err)
	}
	return nil
}

// Close closes the database
func (b *BadgerStore) Close() error {
	return b.db.Close()
}
