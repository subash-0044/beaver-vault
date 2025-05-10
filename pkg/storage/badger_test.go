package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestStore(t *testing.T) (*BadgerStore, string) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "badger-test-*")
	require.NoError(t, err, "Failed to create temp directory")

	// Create a new store instance
	opts := Options{
		Dir:             tmpDir,
		CreateIfMissing: true,
	}

	store, err := NewBadgerStore(opts)
	require.NoError(t, err, "Failed to create BadgerStore")

	return store, tmpDir
}

func cleanupTestStore(t *testing.T, store *BadgerStore, tmpDir string) {
	err := store.Close()
	assert.NoError(t, err, "Failed to close store")

	err = os.RemoveAll(tmpDir)
	assert.NoError(t, err, "Failed to remove temp directory")
}

func TestNewBadgerStore(t *testing.T) {
	t.Run("should create new store with valid options", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "badger-test-*")
		require.NoError(t, err)
		defer func() {
			if err := os.RemoveAll(tmpDir); err != nil {
				t.Errorf("Failed to remove temp directory: %v", err)
			}
		}()

		opts := Options{
			Dir:             tmpDir,
			CreateIfMissing: true,
		}

		store, err := NewBadgerStore(opts)
		require.NoError(t, err)
		assert.NotNil(t, store)
		assert.NoError(t, store.Close())
	})

	t.Run("should fail with invalid directory", func(t *testing.T) {
		opts := Options{
			Dir:             "/nonexistent/directory",
			CreateIfMissing: false,
		}

		store, err := NewBadgerStore(opts)
		assert.Error(t, err)
		assert.Nil(t, store)
	})
}

func TestPut(t *testing.T) {
	store, tmpDir := setupTestStore(t)
	defer cleanupTestStore(t, store, tmpDir)

	t.Run("should successfully put new value", func(t *testing.T) {
		err := store.Put([]byte("key1"), []byte("value1"))
		assert.NoError(t, err)
	})

	t.Run("should successfully update existing value", func(t *testing.T) {
		key := []byte("key2")

		err := store.Put(key, []byte("initial-value"))
		require.NoError(t, err)

		err = store.Put(key, []byte("updated-value"))
		assert.NoError(t, err)
	})

	t.Run("should fail with empty key", func(t *testing.T) {
		err := store.Put([]byte{}, []byte("value"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key cannot be empty")
	})

	t.Run("should handle empty value", func(t *testing.T) {
		err := store.Put([]byte("empty-value-key"), []byte{})
		assert.NoError(t, err)
	})
}

func TestGet(t *testing.T) {
	store, tmpDir := setupTestStore(t)
	defer cleanupTestStore(t, store, tmpDir)

	t.Run("should get existing value", func(t *testing.T) {
		key := []byte("test-key")
		value := []byte("test-value")

		err := store.Put(key, value)
		require.NoError(t, err)

		result, err := store.Get(key)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, value, result)
	})

	t.Run("should return nil for non-existent key", func(t *testing.T) {
		result, err := store.Get([]byte("non-existent-key"))
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("should fail with empty key", func(t *testing.T) {
		result, err := store.Get([]byte{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key cannot be empty")
		assert.Nil(t, result)
	})
}

func TestDelete(t *testing.T) {
	store, tmpDir := setupTestStore(t)
	defer cleanupTestStore(t, store, tmpDir)

	t.Run("should delete existing value", func(t *testing.T) {
		key := []byte("delete-test-key")
		value := []byte("delete-test-value")

		err := store.Put(key, value)
		require.NoError(t, err)

		err = store.Delete(key)
		assert.NoError(t, err)

		result, err := store.Get(key)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("should handle non-existent key", func(t *testing.T) {
		err := store.Delete([]byte("non-existent-key"))
		assert.NoError(t, err)
	})

	t.Run("should fail with empty key", func(t *testing.T) {
		err := store.Delete([]byte{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key cannot be empty")
	})
}
