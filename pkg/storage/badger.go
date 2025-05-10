package storage

import (
	"fmt"
	"os"

	"github.com/dgraph-io/badger/v4"
)

type BadgerStore struct {
	DB *badger.DB
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

	return &BadgerStore{DB: db}, nil
}

// Get retrieves a value for a given key
func (b *BadgerStore) Get(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, ErrKeyCannotBeEmpty
	}

	var value []byte
	err := b.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			return nil
		}
		if err != nil {
			return err
		}

		value, err = item.ValueCopy(nil)
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get value: %w", err)
	}
	return value, nil
}

// Put stores a value for a given key
func (b *BadgerStore) Put(key, data []byte) error {
	if len(key) == 0 {
		return ErrKeyCannotBeEmpty
	}

	err := b.DB.Update(func(txn *badger.Txn) error {
		return txn.Set(key, data)
	})

	if err != nil {
		return fmt.Errorf("failed to put value: %w", err)
	}
	return nil
}

// Delete removes a key-value pair
func (b *BadgerStore) Delete(key []byte) error {
	if len(key) == 0 {
		return ErrKeyCannotBeEmpty
	}

	err := b.DB.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})

	if err != nil {
		return fmt.Errorf("failed to delete key: %w", err)
	}
	return nil
}

// Close closes the database connection
func (b *BadgerStore) Close() error {
	return b.DB.Close()
}
