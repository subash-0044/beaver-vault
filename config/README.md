# Configuration Guide

This document describes the configuration options for the Beaver Vault distributed key-value store.

## Configuration File Structure

The configuration file is in YAML format and consists of three main sections:

### Server Configuration
```yaml
server:
  host: "localhost"  # HTTP server host
  port: 8000        # HTTP server port
```

### Raft Configuration
```yaml
raft:
  nodeId: "node1"           # Unique identifier for the Raft node
  host: "localhost"         # Raft transport host
  port: 7000               # Raft transport port
  bootstrap: true          # Whether to bootstrap the cluster with this node
  heartbeatTimeout: "1s"   # Raft heartbeat timeout
  electionTimeout: "1s"    # Raft election timeout
  commitTimeout: "50ms"    # Raft commit timeout
  maxSnapshots: 3         # Maximum number of snapshots to retain
```

### Data Configuration
```yaml
data:
  directory: "data"  # Directory for storing Raft and BadgerDB data
```

## Usage

To use a custom configuration file, use the `-config` flag when starting the server:

```bash
./server -config path/to/config.yaml
```

If no configuration file is specified, the server will look for `config/config.yaml` in the current directory.

## Configuration Options

### Server Options
- `host`: The hostname or IP address for the HTTP server
- `port`: The port number for the HTTP server

### Raft Options
- `nodeId`: A unique identifier for the Raft node in the cluster
- `host`: The hostname or IP address for Raft communication
- `port`: The port number for Raft communication
- `bootstrap`: Whether this node should bootstrap the cluster
- `heartbeatTimeout`: How often the leader sends heartbeats to followers
- `electionTimeout`: How long followers wait before starting an election
- `commitTimeout`: How long the leader waits for followers to commit
- `maxSnapshots`: Maximum number of Raft snapshots to keep

### Data Options
- `directory`: The directory where all persistent data will be stored

## Example Configuration

```yaml
server:
  host: "localhost"
  port: 8000

raft:
  nodeId: "node1"
  host: "localhost"
  port: 7000
  bootstrap: true
  heartbeatTimeout: "1s"
  electionTimeout: "1s"
  commitTimeout: "50ms"
  maxSnapshots: 3

data:
  directory: "data"
``` 