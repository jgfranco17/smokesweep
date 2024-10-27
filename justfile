# Helper recipes
PROJECT_NAME := "SmokeSweep"

# Default command
default:
    @just --list --unsorted

# Execute unit tests
test:
    @echo "Running {{ PROJECT_NAME }} unit tests!"
    go clean -testcache
    go test -cover ./cli/...

# Sync Go modules
tidy:
    go mod tidy
    cd cli && go mod tidy
    go work sync
    @echo "{{ PROJECT_NAME }} workspace and modules synced successfully!"

# Build the binary
build:
    @echo "Building {{ PROJECT_NAME }} binary..."
    go mod download all
    CGO_ENABLED=0 GOOS=linux go build -o ./smokesweep main.go
    @echo "{{ PROJECT_NAME }} binary built successfully!"

# Build Docker image
build-docker:
    docker build -t smokesweep:dev .

# Run CLI through Docker
run-docker:
    docker run --rm smokesweep:dev --version
