.PHONY: build build-cli test clean clean-empty dist push help

# Build CLI
build: build-cli

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

# Help
help:
	@echo "Available targets:"
	@echo "  build      - Build CLI"
	@echo "  build-cli  - Build CLI"
	@echo "  test       - Run tests"
	@echo "  clean      - Clean build artifacts"
	@echo "  push       - Push to GitHub"
	@echo "  help       - Show this help" 