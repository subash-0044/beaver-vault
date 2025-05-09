package node

import (
	"fmt"
	"net"
	"sync"

	"github.com/subash-0044/beaver-vault/pkg/storage"
)

// Config holds the configuration for a node
type Config struct {
	ID      string
	Address string
	DataDir string
}

// Node represents a single node in the Beaver-Vault cluster
type Node struct {
	id       string
	address  string
	store    storage.Storage
	mu       sync.RWMutex
	listener net.Listener
	running  bool
}

// New creates a new node with the given configuration
func New(cfg Config) (*Node, error) {
	// Initialize storage
	opts := storage.Options{
		Dir:             cfg.DataDir,
		CreateIfMissing: true,
	}
	store, err := storage.NewBadgerStore(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	return &Node{
		id:      cfg.ID,
		address: cfg.Address,
		store:   store,
	}, nil
}

// Start starts the node's services
func (n *Node) Start() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.running {
		return fmt.Errorf("node already running")
	}

	// Start listening for connections
	listener, err := net.Listen("tcp", n.address)
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}
	n.listener = listener
	n.running = true

	// Start serving in a goroutine
	go n.serve()

	return nil
}

// Stop stops the node's services
func (n *Node) Stop() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.running {
		return nil
	}

	// Close the listener
	if n.listener != nil {
		if err := n.listener.Close(); err != nil {
			return fmt.Errorf("failed to close listener: %w", err)
		}
	}

	// Close the storage
	if err := n.store.Close(); err != nil {
		return fmt.Errorf("failed to close storage: %w", err)
	}

	n.running = false
	return nil
}

// serve handles incoming connections
func (n *Node) serve() {
	for {
		conn, err := n.listener.Accept()
		if err != nil {
			// Check if the listener was closed intentionally
			if opErr, ok := err.(*net.OpError); ok && opErr.Op == "accept" {
				return
			}
			// Log error and continue accepting connections
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		// Handle connection in a goroutine
		go n.handleConnection(conn)
	}
}

// handleConnection processes a single client connection
func (n *Node) handleConnection(conn net.Conn) {
	defer conn.Close()
	// TODO: Implement request handling protocol
}

// Get retrieves a value for the given key
func (n *Node) Get(key []byte) ([]byte, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	value, err := n.store.Get(key)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, storage.ErrKeyNotFound
	}
	return value.Data, nil
}

// Put stores a value for the given key
func (n *Node) Put(key, value []byte) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.store.Put(key, value)
}

// Delete removes a key-value pair
func (n *Node) Delete(key []byte) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.store.Delete(key)
}
