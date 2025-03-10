.PHONY: build build-server build-cli test clean help

# Build both server and CLI
build: build-server build-cli

# Build server
build-server:
	@echo "Building server..."
	@mkdir -p bin
	@go build -o bin/server ./cmd/server

# Build CLI
build-cli:
	@echo "Building CLI..."
	@mkdir -p bin
	@go build -o bin/ghrepos ./cmd/cli

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin

# Run server
run:
	@echo "Running server..."
	@go run ./cmd/server -config config.yaml

# Help
help:
	@echo "Available targets:"
	@echo "  build        - Build both server and CLI"
	@echo "  build-server - Build server only"
	@echo "  build-cli    - Build CLI only"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  help         - Show this help" 