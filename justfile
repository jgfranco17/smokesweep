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
