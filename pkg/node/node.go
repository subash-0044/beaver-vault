package node

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/subash-0044/beaver-vault/pkg/storage"
	pb "github.com/subash-0044/beaver-vault/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Config holds the configuration for a node
type Config struct {
	ID      string
	Address string
	DataDir string
}

// Node represents a single node in the Beaver-Vault cluster
type Node struct {
	pb.UnimplementedKeyValueStoreServer
	id         string
	address    string
	store      storage.Storage
	mu         sync.RWMutex
	grpcServer *grpc.Server
	running    bool
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

	n.grpcServer = grpc.NewServer()
	pb.RegisterKeyValueStoreServer(n.grpcServer, n)

	lis, err := net.Listen("tcp", n.address)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	// Start serving in a goroutine
	go func() {
		if err := n.grpcServer.Serve(lis); err != nil {
			fmt.Printf("failed to serve: %v\n", err)
		}
	}()

	n.running = true
	return nil
}

// Stop stops the node's services
func (n *Node) Stop() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.running {
		return nil
	}

	if n.grpcServer != nil {
		n.grpcServer.GracefulStop()
	}

	if err := n.store.Close(); err != nil {
		return fmt.Errorf("failed to close storage: %w", err)
	}

	n.running = false
	return nil
}

// Get handles get requests
func (n *Node) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	if len(req.Key) == 0 {
		return nil, status.Error(codes.InvalidArgument, storage.ErrKeyCannotBeEmpty.Error())
	}

	n.mu.RLock()
	defer n.mu.RUnlock()

	value, err := n.store.Get(req.Key)
	if err != nil {
		return &pb.GetResponse{
			Found: false,
			Error: err.Error(),
		}, nil
	}

	if value == nil {
		return &pb.GetResponse{
			Found: false,
			Error: "key not found",
		}, nil
	}

	return &pb.GetResponse{
		Found: true,
		Value: value.Data,
	}, nil
}

// Put handles put requests
func (n *Node) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	if len(req.Key) == 0 {
		return nil, status.Error(codes.InvalidArgument, storage.ErrKeyCannotBeEmpty.Error())
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	if err := n.store.Put(req.Key, req.Value); err != nil {
		return &pb.PutResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.PutResponse{
		Success: true,
	}, nil
}

// Delete handles delete requests
func (n *Node) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	if len(req.Key) == 0 {
		return nil, status.Error(codes.InvalidArgument, storage.ErrKeyCannotBeEmpty.Error())
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	if err := n.store.Delete(req.Key); err != nil {
		return &pb.DeleteResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.DeleteResponse{
		Success: true,
	}, nil
}
