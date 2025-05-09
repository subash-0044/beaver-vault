# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Protobuf parameters
PROTOC=protoc
PROTO_DIR=proto
PROTO_FILES=$(wildcard $(PROTO_DIR)/*.proto)

# Install protobuf compiler and plugins
.PHONY: install-proto-tools
install-proto-tools:
	# Install protoc (protocol buffer compiler)
	@if [ "$(shell uname)" = "Darwin" ]; then \
		brew install protobuf; \
	elif [ "$(shell uname)" = "Linux" ]; then \
		sudo apt-get update && sudo apt-get install -y protobuf-compiler; \
	else \
		echo "Unsupported OS. Please install protobuf compiler manually."; \
		exit 1; \
	fi
	# Install Go plugins for protoc
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.6
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

# Generate protobuf files
.PHONY: proto
proto:
	$(PROTOC) --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		$(PROTO_FILES)

# Download Go dependencies
.PHONY: deps
deps:
	$(GOMOD) download
	$(GOMOD) verify

# Run tests
.PHONY: test
test:
	$(GOTEST) -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Clean generated files
.PHONY: clean
clean:
	rm -f coverage.out coverage.html
	find . -type f -name '*.pb.go' -delete

# Build the project
.PHONY: build
build:
	$(GOBUILD) -v ./...

# Run all necessary setup steps
.PHONY: setup
setup: install-proto-tools deps proto

# Default target
.PHONY: all
all: setup test build 