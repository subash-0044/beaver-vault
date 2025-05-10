package parser

import (
	"encoding/json"
	"fmt"
)

// JSONValue represents a JSON-serializable value
type JSONValue struct {
	Data interface{}
}

// Store interface defines the required storage operations
type Store interface {
	Get(key []byte) ([]byte, error)
	Put(key, value []byte) error
	Delete(key []byte) error
}

// Parser handles JSON serialization and storage operations
type Parser struct {
	store Store
}

// NewParser creates a new Parser instance
func NewParser(store Store) *Parser {
	return &Parser{store: store}
}

// UnmarshalTo converts JSON bytes to a specific type
func (p *Parser) UnmarshalTo(data []byte, v any) error {
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, v)
}

// Get retrieves and unmarshals a JSON value for a given key
func (p *Parser) Get(key string) (*JSONValue, error) {
	if len(key) == 0 {
		return nil, fmt.Errorf("key cannot be empty")
	}

	value, err := p.store.Get([]byte(key))
	if err != nil {
		return nil, err
	}

	if len(value) == 0 {
		return &JSONValue{Data: make(map[string]any)}, nil
	}

	var data any
	if err := json.Unmarshal(value, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return &JSONValue{Data: data}, nil
}

// Put marshals and stores a JSON value for a given key
func (p *Parser) Put(key string, value any) error {
	if len(key) == 0 {
		return fmt.Errorf("key cannot be empty")
	}

	if value == nil {
		return nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return p.store.Put([]byte(key), data)
}

// Delete removes a key-value pair
func (p *Parser) Delete(key string) error {
	if len(key) == 0 {
		return fmt.Errorf("key cannot be empty")
	}
	return p.store.Delete([]byte(key))
}
