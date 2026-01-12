.PHONY: build compile install clean test lint help

BINARY_NAME=unifi
BUILD_DIR=bin

help:
	@echo "Available targets:"
	@echo "  build     - Build the binary to ./bin/unifi"
	@echo "  compile   - Alias for build"
	@echo "  install   - Install to GOPATH/bin using 'go install'"
	@echo "  clean     - Remove build artifacts"
	@echo "  test      - Run tests"
	@echo "  lint      - Run linter (requires golangci-lint)"

build: clean
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

compile: build

install:
	@echo "Installing $(BINARY_NAME)..."
	@go install
	@echo "Installation complete"

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

test:
	@echo "Running tests..."
	@go test -v ./...

lint:
	@echo "Running linter..."
	@golangci-lint run
