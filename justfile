# ccc — Copilot Config CLI
# Run `just` to see all available recipes

set dotenv-load := false

mod         := "github.com/jsburckhardt/co-config"
binary      := "ccc"
cmd         := "./cmd/ccc"
version     := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`

# List available recipes
default:
    @just --list

# Build the ccc binary
build:
    go build -ldflags "-X main.version={{version}}" -o {{binary}} {{cmd}}

# Run ccc (builds first)
run *ARGS: build
    ./{{binary}} {{ARGS}}

# Run all tests
test:
    go test ./...

# Run tests with verbose output
test-verbose:
    go test -v ./...

# Run tests with coverage
test-cover:
    go test -cover ./...

# Run tests with race detector
test-race:
    go test -race ./...

# Run only unit tests (skip integration)
test-unit:
    go test -short ./...

# Run only integration tests
test-integration:
    go test -v -run Integration ./...

# Run go vet
vet:
    go vet ./...

# Format all Go files
fmt:
    gofmt -w .

# Check formatting (no changes)
fmt-check:
    @test -z "$(gofmt -l .)" || (echo "Files need formatting:" && gofmt -l . && exit 1)

# Tidy go.mod and go.sum
tidy:
    go mod tidy

# Full check: fmt, vet, test
check: fmt-check vet test

# Clean build artifacts
clean:
    rm -f {{binary}}

# Install the binary to $GOPATH/bin
install:
    go install -ldflags "-X main.version={{version}}" {{cmd}}

# Show current dependencies
deps:
    go list -m all

# Update all dependencies
deps-update:
    go get -u ./...
    go mod tidy
