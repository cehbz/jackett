.PHONY: build test clean lint fmt vet coverage help

# Default target
all: build

# Build the project
build:
	@echo "Building..."
	go build -o jackett .

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	go clean
	rm -f jackett
	rm -f coverage.out coverage.html
	rm -f *.test

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Run golint (if installed)
lint:
	@echo "Running linter..."
	@if command -v golint >/dev/null 2>&1; then \
		golint ./...; \
	else \
		echo "golint not installed. Install with: go install golang.org/x/lint/golint@latest"; \
	fi

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Run all checks
check: fmt vet lint test

# Show help
help:
	@echo "Available targets:"
	@echo "  build     - Build the project"
	@echo "  test      - Run tests"
	@echo "  coverage  - Run tests with coverage report"
	@echo "  clean     - Clean build artifacts"
	@echo "  fmt       - Format code"
	@echo "  vet       - Run go vet"
	@echo "  lint      - Run golint"
	@echo "  deps      - Install dependencies"
	@echo "  check     - Run all checks (fmt, vet, lint, test)"
	@echo "  help      - Show this help" 