package parser

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

// mockStore implements the Store interface for testing
type mockStore struct {
	data   map[string][]byte
	getErr error
	putErr error
	delErr error
}

func newMockStore() *mockStore {
	return &mockStore{
		data: make(map[string][]byte),
	}
}

func (m *mockStore) Get(key []byte) ([]byte, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.data[string(key)], nil
}

func (m *mockStore) Put(key, value []byte) error {
	if m.putErr != nil {
		return m.putErr
	}
	m.data[string(key)] = value
	return nil
}

func (m *mockStore) Delete(key []byte) error {
	if m.delErr != nil {
		return m.delErr
	}
	delete(m.data, string(key))
	return nil
}

func TestNewParser(t *testing.T) {
	store := newMockStore()
	parser := NewParser(store)
	if parser.store == nil {
		t.Error("Expected store to be set, got nil")
	}
}

func TestUnmarshalTo(t *testing.T) {
	store := newMockStore()
	parser := NewParser(store)

	tests := []struct {
		name    string
		data    []byte
		want    interface{}
		wantErr bool
	}{
		{
			name: "valid json",
			data: []byte(`{"name":"test"}`),
			want: map[string]interface{}{"name": "test"},
		},
		{
			name:    "empty data",
			data:    []byte{},
			want:    nil,
			wantErr: false,
		},
		{
			name:    "invalid json",
			data:    []byte(`{"invalid`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got interface{}
			err := parser.UnmarshalTo(tt.data, &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalTo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UnmarshalTo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_Get(t *testing.T) {
	store := newMockStore()
	parser := NewParser(store)

	// Set up test data
	testData := map[string]interface{}{"test": "value"}
	jsonData, _ := json.Marshal(testData)
	store.data["testKey"] = jsonData

	tests := []struct {
		name    string
		key     string
		want    *JSONValue
		wantErr bool
		setErr  error
	}{
		{
			name: "existing key",
			key:  "testKey",
			want: &JSONValue{Data: testData},
		},
		{
			name:    "empty key",
			key:     "",
			wantErr: true,
		},
		{
			name:    "store error",
			key:     "errorKey",
			setErr:  fmt.Errorf("store error"),
			wantErr: true,
		},
		{
			name: "non-existent key",
			key:  "nonexistent",
			want: &JSONValue{Data: make(map[string]interface{})},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store.getErr = tt.setErr
			got, err := parser.Get(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_Put(t *testing.T) {
	store := newMockStore()
	parser := NewParser(store)

	tests := []struct {
		name    string
		key     string
		value   interface{}
		wantErr bool
		setErr  error
	}{
		{
			name:  "valid put",
			key:   "testKey",
			value: map[string]interface{}{"test": "value"},
		},
		{
			name:    "empty key",
			key:     "",
			value:   "test",
			wantErr: true,
		},
		{
			name:  "nil value",
			key:   "nilKey",
			value: nil,
		},
		{
			name:    "store error",
			key:     "errorKey",
			value:   "test",
			setErr:  fmt.Errorf("store error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store.putErr = tt.setErr
			err := parser.Put(tt.key, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Put() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParser_Delete(t *testing.T) {
	store := newMockStore()
	parser := NewParser(store)

	tests := []struct {
		name    string
		key     string
		wantErr bool
		setErr  error
	}{
		{
			name: "valid delete",
			key:  "testKey",
		},
		{
			name:    "empty key",
			key:     "",
			wantErr: true,
		},
		{
			name:    "store error",
			key:     "errorKey",
			setErr:  fmt.Errorf("store error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store.delErr = tt.setErr
			err := parser.Delete(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
