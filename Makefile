.PHONY: help certs build run test clean fmt lint

help:
	@echo "Available targets:"
	@echo "  make certs       - Generate TLS certificates (self-signed)"
	@echo "  make build       - Build the application"
	@echo "  make run         - Run the application (requires certs)"
	@echo "  make test        - Run unit tests"
	@echo "  make fmt         - Format Go code"
	@echo "  make lint        - Run linter"
	@echo "  make clean       - Clean build artifacts and certificates"

# Generate self-signed TLS certificates
certs:
	@echo "Generating self-signed TLS certificates..."
	@go run gencert.go
	@echo "Certificates generated: server.crt and server.key"

# Build the application
build: go.mod
	@echo "Building application..."
	@go build -o teleport-browser -v

# Run the application
run: certs build
	@echo "Starting server..."
	@echo "Test credentials:"
	@echo "  admin / admin123"
	@echo "  user / password"
	@echo "Navigate to: https://localhost:8443 (ignore self-signed cert warning)"
	@./teleport-browser

# Run tests
test:
	@echo "Running tests..."
	@go test -v -race -cover -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	@go vet ./...

# Clean
clean:
	@echo "Cleaning..."
	@rm -f teleport-browser server.crt server.key coverage.out cert.conf
	@go clean

# Development: run with hot reload (requires air or similar)
dev:
	@echo "Make sure 'air' is installed: go install github.com/cosmtrek/air@latest"
	@air
