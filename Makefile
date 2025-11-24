# Rockstar Web Framework Makefile

.PHONY: build test clean fmt vet lint deps

# Build the framework
build:
	go build ./...

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	go clean ./...
	rm -f coverage.out coverage.html

# Format code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Download dependencies
deps:
	go mod download
	go mod tidy

# Run all checks
check: fmt vet test

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the framework"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  clean         - Clean build artifacts"
	@echo "  fmt           - Format code"
	@echo "  vet           - Run go vet"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  check         - Run fmt, vet, and test"
	@echo "  help          - Show this help message"