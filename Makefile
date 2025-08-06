.PHONY: build clean test install

# Build the application
build:
	go build -o bin/nimby_shapetopoi ./cmd/nimby_shapetopoi

# Build for multiple platforms
build-all:
	GOOS=windows GOARCH=amd64 go build -o bin/nimby_shapetopoi-windows-amd64.exe ./cmd/nimby_shapetopoi
	GOOS=darwin GOARCH=amd64 go build -o bin/nimby_shapetopoi-darwin-amd64 ./cmd/nimby_shapetopoi
	GOOS=darwin GOARCH=arm64 go build -o bin/nimby_shapetopoi-darwin-arm64 ./cmd/nimby_shapetopoi
	GOOS=linux GOARCH=amd64 go build -o bin/nimby_shapetopoi-linux-amd64 ./cmd/nimby_shapetopoi

# Install dependencies
deps:
	go mod download
	go mod tidy

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f *.zip *.tsv

# Run tests
test:
	go test ./...

# Run tests with verbose output
test-verbose:
	go test -v ./...

# Run tests with race detection
test-race:
	go test -race ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run benchmarks
bench:
	go test -bench=. ./...

# Install the binary to GOBIN
install:
	go install ./cmd/nimby_shapetopoi

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Run linter with fixes
lint-fix:
	golangci-lint run --fix

# Development build (with race detector)
build-dev:
	go build -race -o bin/nimby_shapetopoi-dev ./cmd/nimby_shapetopoi