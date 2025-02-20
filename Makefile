# Makefile for HEPop-Go

# Variables
BINARY_NAME=hepop-go
SRC_DIR=./cmd/hepop
CONFIG_FILE=config/config.yaml

# Commands
.PHONY: all build test clean run

all: build

build:
	@echo "Building project..."
	go build -o $(BINARY_NAME) $(SRC_DIR)

test:
	@echo "Running tests..."
	go test ./... -v

clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)

run: build
	@echo "Running project..."
	./$(BINARY_NAME) -config $(CONFIG_FILE)

lint:
	@echo "Running linter..."
	golangci-lint run

fmt:
	@echo "Formatting code..."
	go fmt ./...

tidy:
	@echo "Updating dependencies..."
	go mod tidy

# Additional commands
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):latest .

docker-run:
	@echo "Running Docker container..."
	docker run --rm -p 8080:8080 -p 9060:9060 $(BINARY_NAME):latest

# Help command
help:
	@echo "Available commands:"
	@echo "  make build       - Build project"
	@echo "  make test        - Run tests"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make run         - Build and run project"
	@echo "  make lint        - Run linter"
	@echo "  make fmt         - Format code"
	@echo "  make tidy        - Update dependencies"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run  - Run Docker container"