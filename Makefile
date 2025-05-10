# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Download Go dependencies
.PHONY: deps
deps:
	$(GOMOD) download
	$(GOMOD) verify

# Run linter
.PHONY: lint
lint:
	golangci-lint run --timeout=5m

# Run tests
.PHONY: test
test:
	$(GOTEST) -v ./...

# Run tests with race detection
.PHONY: test-race
test-race:
	$(GOTEST) -v -race ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.txt -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.txt

# Clean generated files
.PHONY: clean
clean:
	rm -f coverage.txt

# Build the project
.PHONY: build
build:
	$(GOBUILD) -v ./...

# Run all necessary setup steps
.PHONY: setup
setup: deps

# Default target
.PHONY: all
all: setup lint test build 