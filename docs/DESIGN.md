# Beaver-Vault Design Document

## What is Beaver-Vault?
Beaver-Vault is a distributed key-value store written in Go. It's designed to store data across multiple computers, making the data safe and reliable.

## Main Components

### 1. Storage (BadgerDB)
- Stores data on local computer
- Optimized for SSD performance
- Stores data in JSON format

### 2. Consensus (Raft)
- Keeps multiple servers in sync
- Uses leader-follower system
- Maintains data consistency
- Handles automatic recovery during node failures

### 3. API Server (Gin)
- Provides HTTP API
- Handles client requests
- Provides simple UI for monitoring

### 4. Node Management
- Ability to add new nodes to cluster
- Ability to remove existing nodes
- Node status monitoring

## How it Works

1. Client Request:
   - Client can send request to any node
   - Request reaches the leader node

2. Data Storage:
   - Leader node stores the data
   - Data automatically syncs to other nodes
   - Data consistency is maintained

3. Node Failure:
   - If a node fails, system recovers automatically
   - New leader is selected
   - No data loss occurs

## Features

1. High Availability:
   - Data is stored on multiple nodes
   - System keeps running even if nodes fail

2. Strong Consistency:
   - Data remains same on all nodes
   - Write operations go through leader

3. Easy to Use:
   - Simple HTTP API
   - Web UI for monitoring
   - Easy node management

4. Scalable:
   - Can easily add new nodes
   - Performance improves automatically

## Technical Details

1. Storage:
   - BadgerDB for local storage
   - JSON format for data
   - Fast read/write operations

2. Network:
   - TCP for node communication
   - HTTP for client requests
   - Raft protocol for consensus

3. Configuration:
   - YAML config file
   - Command line options
   - Environment variables

## Usage

1. Start Node:
   ```bash
   ./server -node-id node1 -http-port 8000 -raft-port 7000
   ```

2. Add Node:
   ```bash
   curl -X POST http://localhost:8000/api/v1/raft/join -d '{"NodeID": "node2", "RaftAddress": "localhost:7001"}'
   ```

3. Store Data:
   ```bash
   curl -X PUT http://localhost:8000/api/v1/kv/mykey -d '"myvalue"'
   ```

4. Get Data:
   ```bash
   curl http://localhost:8000/api/v1/kv/mykey
   ```