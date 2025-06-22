# Makefile for Jackett Go client library

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOFMT=gofmt
GOVET=$(GOCMD) vet
GOLINT=golint

# Package name
PACKAGE=github.com/cehbz/jackett

# Test flags
TEST_FLAGS=-v -race -coverprofile=coverage.out

.PHONY: all build clean test coverage fmt vet lint install

# Default target
all: test build

# Build the library (no binary output for a library)
build:
	@echo "Building Jackett client library..."
	$(GOBUILD) -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f coverage.out coverage.html

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) $(TEST_FLAGS) ./...

# Generate test coverage report
coverage: test
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -w .

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

# Run golint (if installed)
lint:
	@echo "Running golint..."
	@which golint > /dev/null || (echo "golint not installed. Install with: go install golang.org/x/lint/golint@latest" && exit 1)
	$(GOLINT) ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOGET) -v ./...

# Install the library (for a library, this just ensures it can be imported)
install: deps
	@echo "Jackett client library ready for use"
	@echo "Import with: import \"$(PACKAGE)\""

# Run all checks
check: fmt vet lint test

# Show help
help:
	@echo "Available targets:"
	@echo "  all      - Run tests and build (default)"
	@echo "  build    - Build the library"
	@echo "  clean    - Remove build artifacts"
	@echo "  test     - Run tests"
	@echo "  coverage - Generate test coverage report"
	@echo "  fmt      - Format code"
	@echo "  vet      - Run go vet"
	@echo "  lint     - Run golint"
	@echo "  deps     - Install dependencies"
	@echo "  install  - Prepare library for use"
	@echo "  check    - Run all checks (fmt, vet, lint, test)"
	