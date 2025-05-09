# Beaver-Vault: Distributed Key-Value Store Design Document

## 1. System Overview
Beaver-Vault is a distributed key-value store designed for high scalability, strong consistency, and fault tolerance. The system employs a peer-to-peer architecture where each node can serve both read and write requests while participating in cluster consensus.

## 2. Architecture Components

### 2.1 Node Architecture
Each node in the system consists of the following components:
- **Storage Engine**: Local key-value storage implementation
- **Consensus Module**: Raft-based distributed consensus
- **Network Layer**: gRPC-based communication
- **Membership Manager**: Serf-based cluster membership
- **Request Handler**: Client request processing

### 2.2 Data Flow
```
Client Request → Request Handler → Consensus Module → Storage Engine
                                ↓
                          Network Layer
                                ↓
                    Other Cluster Nodes
```

## 3. Key Design Decisions

### 3.1 Consensus Protocol (Raft)
- **Why Raft?**
  - Simple to understand and implement
  - Strong leader-based approach
  - Built-in leader election and log replication
  - Proven in production systems
- **Implementation**: Using HashiCorp's Raft library

### 3.2 Node Discovery and Membership (Serf)
- **Why Serf?**
  - Efficient gossip protocol
  - Automatic node failure detection
  - Low bandwidth overhead
  - Scalable membership management
- **Implementation**: Using HashiCorp's Serf library

### 3.3 Communication Protocol (gRPC)
- **Why gRPC?**
  - High performance bi-directional streaming
  - Strong typing with Protocol Buffers
  - Built-in load balancing and service discovery
  - Excellent support for Go

### 3.4 Data Partitioning
- Using Consistent Hashing
- Partition rebalancing on node changes
- Handles node addition and removal seamlessly

### 3.5 Storage Engine
- **Why BadgerDB?**: 
    - Pure Go implementation
    - Optimized for SSDs
    - Excellent performance for both point lookups and range scans
    - Built-in support for transactions and versioning
    - LSM(Log-Structured Merge-tree) tree-based architecture for better write performance
    - Compression support for reduced storage footprint
  
- **Storage Layout**:
  ```
  /data
    /raft              # Raft consensus logs and snapshots
      /logs
      /snapshots
    /keyvalue         # Actual key-value data store
    /metadata         # Cluster metadata and configuration
  ```

## 4. Consistency Model
- Strong consistency through Raft consensus
- Write operations require majority agreement
- Read operations can be configured for strong or eventual consistency

## 5. Fault Tolerance
### 5.1 Node Failures
- Automatic leader election via Raft
- Data replication across multiple nodes
- Configurable replication factor

### 5.2 Network Partitions
- Partition tolerance through Raft
- Minority partitions become read-only
- Automatic recovery when partition heals

## 6. Scalability Features
- Horizontal scaling through node addition
- Dynamic partition rebalancing
- Read scalability through replicas
- Write scalability through partitioning

## 7. Performance Considerations
### 7.1 Write Path
1. Client sends write request
2. Leader node receives request
3. Raft consensus process
4. Commit to storage engine
5. Acknowledge to client

### 7.2 Read Path
1. Client sends read request
2. Node checks if it has the data
3. If consistent read: verify leadership
4. Return data to client

## 8. Monitoring and Operations
- Health checks via Serf
- Metrics collection points
- Operational commands for cluster management
