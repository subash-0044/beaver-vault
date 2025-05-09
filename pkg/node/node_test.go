package node

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/subash-0044/beaver-vault/pkg/storage"
)

func TestNode(t *testing.T) {
	// Create temporary directory for node data
	tmpDir, err := os.MkdirTemp("", "beaver-vault-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	t.Run("should handle node lifecycle properly", func(t *testing.T) {
		cfg := Config{
			ID:      "node1",
			Address: "localhost:8001",
			DataDir: filepath.Join(tmpDir, "node1"),
		}

		node, err := New(cfg)
		require.NoError(t, err)
		require.NoError(t, node.Start())

		testKey := []byte("test-key")
		testValue := []byte("test-value")

		err = node.Put(testKey, testValue)
		require.NoError(t, err)

		value, err := node.Get(testKey)
		require.NoError(t, err)
		assert.Equal(t, testValue, value)

		err = node.Delete(testKey)
		require.NoError(t, err)

		_, err = node.Get(testKey)
		assert.ErrorIs(t, err, storage.ErrKeyNotFound)

		require.NoError(t, node.Stop())
	})

	t.Run("should manage the node state properly", func(t *testing.T) {
		cfg := Config{
			ID:      "node1",
			Address: "localhost:8002",
			DataDir: filepath.Join(tmpDir, "node2"),
		}

		node, err := New(cfg)
		require.NoError(t, err)

		// Starting twice should fail
		require.NoError(t, node.Start())
		assert.Error(t, node.Start())

		// Stopping twice should be fine
		require.NoError(t, node.Stop())
		assert.NoError(t, node.Stop())

		// Can start again after stop
		assert.NoError(t, node.Start())
		assert.NoError(t, node.Stop())
	})

	t.Run("should handle concurrent operations properly", func(t *testing.T) {
		cfg := Config{
			ID:      "node1",
			Address: "localhost:8003",
			DataDir: filepath.Join(tmpDir, "node3"),
		}

		// Create and start node
		node, err := New(cfg)
		require.NoError(t, err)
		require.NoError(t, node.Start())
		defer node.Stop()

		// Test concurrent operations
		const numOperations = 100
		var wg sync.WaitGroup
		wg.Add(2)

		// Writer goroutine
		go func() {
			defer wg.Done()
			for i := 0; i < numOperations; i++ {
				key := []byte(fmt.Sprintf("key-%d", i))
				value := []byte(fmt.Sprintf("value-%d", i))
				err := node.Put(key, value)
				require.NoError(t, err)
			}
		}()

		// Reader goroutine
		go func() {
			defer wg.Done()
			for i := 0; i < numOperations; i++ {
				key := []byte(fmt.Sprintf("key-%d", i))
				_, _ = node.Get(key) // Errors are expected as some keys might not exist yet
			}
		}()

		// Wait for both goroutines to finish
		wg.Wait()

		// Verify final state
		for i := 0; i < numOperations; i++ {
			key := []byte(fmt.Sprintf("key-%d", i))
			value, err := node.Get(key)
			require.NoError(t, err)
			assert.Equal(t, []byte(fmt.Sprintf("value-%d", i)), value)
		}
	})
}
