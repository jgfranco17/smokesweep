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

test-report:
    @echo "Running {{ PROJECT_NAME }} unit tests with reporting!"
    go clean -testcache
    go test -cover -json ./cli/... | go-test-report -o smokesweep-test-report.html -t "SmokeSweep Test Report" -g 1
    xdg-open smokesweep-test-report.html

# Sync Go modules
tidy:
    go mod tidy
    cd cli && go mod tidy
    go work sync
    @echo "{{ PROJECT_NAME }} workspace and modules synced successfully!"

# Build the binary
build:
    #!/usr/bin/env bash
    echo "Building {{ PROJECT_NAME }} binary..."
    go mod download all
    VERSION=$(jq -r .version info.json)
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-X main.version=${VERSION}" -o ./smokesweep main.go
    echo "Built binary for {{ PROJECT_NAME }} ${VERSION} successfully!"

# Build Docker image
build-docker:
    docker build -t smokesweep:dev .

# Run CLI through Docker
run-docker:
    docker run --rm smokesweep:dev --version
