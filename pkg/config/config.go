package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	Server ServerConfig `yaml:"server"`
	Raft   RaftConfig   `yaml:"raft"`
	Data   DataConfig   `yaml:"data"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// RaftConfig holds Raft consensus configuration
type RaftConfig struct {
	NodeID           string `yaml:"nodeId"`
	Host             string `yaml:"host"`
	Port             int    `yaml:"port"`
	Bootstrap        bool   `yaml:"bootstrap"`
	HeartbeatTimeout string `yaml:"heartbeatTimeout"`
	ElectionTimeout  string `yaml:"electionTimeout"`
	CommitTimeout    string `yaml:"commitTimeout"`
	MaxSnapshots     int    `yaml:"maxSnapshots"`
}

// DataConfig holds data storage configuration
type DataConfig struct {
	Directory string `yaml:"directory"`
}

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return &config, nil
}

// GetHTTPAddress returns the formatted HTTP address
func (c *ServerConfig) GetHTTPAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetRaftAddress returns the formatted Raft address
func (c *RaftConfig) GetRaftAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
