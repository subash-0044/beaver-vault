package handler

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v4"
)

// Get fetches data from BadgerDB where the Raft uses to store data.
// This method can be called on any Raft server, offering eventual consistency on read.
func (h Handler) Get(key string) (any, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, fmt.Errorf("key is empty")
	}

	txn := h.db.NewTransaction(false)
	defer func() {
		_ = txn.Commit()
	}()

	item, err := txn.Get([]byte(key))
	if err == badger.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting key %s from storage: %s", key, err.Error())
	}

	var value []byte
	err = item.Value(func(val []byte) error {
		value = append(value, val...)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error retrieving value for key %s: %s", key, err.Error())
	}

	var data any
	if len(value) > 0 {
		if err = json.Unmarshal(value, &data); err != nil {
			return nil, fmt.Errorf("error unmarshaling data for key %s: %s", key, err.Error())
		}
	}

	return data, nil
}
