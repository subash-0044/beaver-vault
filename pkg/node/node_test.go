package node

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/subash-0044/beaver-vault/pkg/storage"
	pb "github.com/subash-0044/beaver-vault/proto"
	"google.golang.org/grpc/status"
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

		putResp, err := node.Put(context.Background(), &pb.PutRequest{
			Key:   testKey,
			Value: testValue,
		})
		require.NoError(t, err)
		assert.True(t, putResp.Success)

		getResp, err := node.Get(context.Background(), &pb.GetRequest{
			Key: testKey,
		})
		require.NoError(t, err)
		assert.True(t, getResp.Found)
		assert.Equal(t, testValue, getResp.Value)

		delResp, err := node.Delete(context.Background(), &pb.DeleteRequest{
			Key: testKey,
		})
		require.NoError(t, err)
		assert.True(t, delResp.Success)

		getResp, err = node.Get(context.Background(), &pb.GetRequest{
			Key: testKey,
		})
		require.NoError(t, err)
		assert.False(t, getResp.Found)
		assert.Equal(t, "key not found", getResp.Error)

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
				resp, err := node.Put(context.Background(), &pb.PutRequest{
					Key:   key,
					Value: value,
				})
				require.NoError(t, err)
				require.True(t, resp.Success)
			}
		}()

		// Reader goroutine
		go func() {
			defer wg.Done()
			for i := 0; i < numOperations; i++ {
				key := []byte(fmt.Sprintf("key-%d", i))
				_, _ = node.Get(context.Background(), &pb.GetRequest{
					Key: key,
				}) // Errors are expected as some keys might not exist yet
			}
		}()

		// Wait for both goroutines to finish
		wg.Wait()

		// Verify final state
		for i := 0; i < numOperations; i++ {
			key := []byte(fmt.Sprintf("key-%d", i))
			resp, err := node.Get(context.Background(), &pb.GetRequest{
				Key: key,
			})
			require.NoError(t, err)
			require.True(t, resp.Found)
			assert.Equal(t, []byte(fmt.Sprintf("value-%d", i)), resp.Value)
		}
	})

	t.Run("should handle empty key properly", func(t *testing.T) {
		cfg := Config{
			ID:      "node1",
			Address: "localhost:8004",
			DataDir: filepath.Join(tmpDir, "node4"),
		}

		node, err := New(cfg)
		require.NoError(t, err)
		require.NoError(t, node.Start())
		defer node.Stop()

		// Test Get with empty key
		_, err = node.Get(context.Background(), &pb.GetRequest{
			Key: []byte{},
		})
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Contains(t, st.Message(), storage.ErrKeyCannotBeEmpty.Error())

		// Test Put with empty key
		_, err = node.Put(context.Background(), &pb.PutRequest{
			Key:   []byte{},
			Value: []byte("value"),
		})
		assert.Error(t, err)
		st, ok = status.FromError(err)
		assert.True(t, ok)
		assert.Contains(t, st.Message(), storage.ErrKeyCannotBeEmpty.Error())

		// Test Delete with empty key
		_, err = node.Delete(context.Background(), &pb.DeleteRequest{
			Key: []byte{},
		})
		assert.Error(t, err)
		st, ok = status.FromError(err)
		assert.True(t, ok)
		assert.Contains(t, st.Message(), storage.ErrKeyCannotBeEmpty.Error())
	})
}
