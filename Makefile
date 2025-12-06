.PHONY: all build run test bench clean help

# Binary name
BINARY_NAME=engine-server

# Build the application
build:
	@echo "Building..."
	go build -o bin/$(BINARY_NAME) cmd/server/main.go

# Run the application
run:
	@echo "Running..."
	go run cmd/server/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./tests/...

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. ./tests/...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin
	go clean

# Show help
help:
	@echo "Available targets:"
	@echo "  build  - Build the application binary"
	@echo "  run    - Run the application directly"
	@echo "  test   - Run all tests"
	@echo "  bench  - Run benchmarks"
	@echo "  clean  - Remove build artifacts"
	@echo "  help   - Show this help message"

