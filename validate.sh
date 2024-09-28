#!/bin/bash

set -e

# Configuration
TEST_DB="./goku.db"
CACHE_DB="./goku_cache.db"
DUCKDB_DB="./goku_stats.duckdb"
GOKU_CMD="./bin/goku"
EXPORT_FILE="exported_bookmarks.html"

# Helper function to run goku commands
run_goku() {
    GOKU_DB_PATH="$TEST_DB" GOKU_CACHE_DB_PATH="$CACHE_DB" GOKU_DUCKDB_PATH="$DUCKDB_DB" $GOKU_CMD "$@"
}

# Helper function to log messages
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1"
}

# Cleanup function
cleanup() {
    log "Cleaning up..."
    rm -f "$TEST_DB" "$CACHE_DB" "$DUCKDB_DB" "$EXPORT_FILE"
}

test_command() {
    log "Testing: $1"
    if [ "$1" = "purge" ]; then
        # Special handling for purge command
        output=$(echo "y" | run_goku "$@" 2>&1)
    else
        output=$(run_goku "$@" 2>&1)
    fi
    exit_code=$?
    if [ $exit_code -eq 0 ]; then
        log "Success: $1"
        echo "$output"
    else
        log "Failed: $1"
        echo "Error output:"
        echo "$output"
        exit 1
    fi
}

# Main execution
trap cleanup EXIT

log "Starting validation for Goku CLI"

# Build the CLI
log "Building Goku CLI..."
if ! ./build.sh; then
    log "Build failed"
    exit 1
fi

# Run tests
test_command add --url "https://example.com" --title "Example Site" --description "An example website" --tags "example,test"
test_command get --id 1
test_command update --id 1 --title "Updated Example Site" --description "An updated example website" --tags "example,test,updated"
test_command list
test_command search --query "example"
test_command tags list
test_command stats
test_command export --output "$EXPORT_FILE"
test_command purge
test_command import --file "$EXPORT_FILE"
test_command sync
test_command delete --id 1

log "All tests completed successfully!"
exit 0