#!/bin/bash

set -e

# Configuration
DEFAULT_USER="goku"
TEST_USER="dev"
TEST_DB="./${DEFAULT_USER}.db"
CACHE_DB="./${DEFAULT_USER}_cache.db"
DUCKDB_DB="./${DEFAULT_USER}_stats.duckdb"
DUCKDB_DB_WAL="./${DEFAULT_USER}_stats.duckdb.wal"
TEST_USER_DB="./${TEST_USER}.db"
TEST_USER_CACHE_DB="./${TEST_USER}_cache.db"
TEST_USER_DUCKDB_DB="./${TEST_USER}_stats.duckdb"
TEST_USER_DUCKDB_DB_WAL="./${TEST_USER}_stats.duckdb.wal"
GOKU_CMD="./bin/goku"
EXPORT_FILE="exported_bookmarks.html"
TEST_USER_EXPORT_FILE="exported_${TEST_USER}_bookmarks.html"
TXT_FILE="bookmarks.txt"
TEST_USER_TXT_FILE="bookmarks_${TEST_USER}.txt"

# Helper function to run goku commands
run_goku() {
    GOKU_DB_PATH="$TEST_DB" GOKU_CACHE_DB_PATH="$CACHE_DB" GOKU_DUCKDB_PATH="$DUCKDB_DB" $GOKU_CMD "$@"
}

# Helper function to run goku commands for test user
run_goku_test_user() {
    GOKU_DB_PATH="$TEST_USER_DB" GOKU_CACHE_DB_PATH="$TEST_USER_CACHE_DB" GOKU_DUCKDB_PATH="$TEST_USER_DUCKDB_DB" $GOKU_CMD --user "$TEST_USER" "$@"
}

# Helper function to log messages
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1"
}

# Cleanup function
cleanup() {
    log "Cleaning up..."
    rm -f "$TEST_DB" "$CACHE_DB" "$DUCKDB_DB" "$DUCKDB_DB_WAL" "$EXPORT_FILE"
    rm -f "$TEST_USER_DB" "$TEST_USER_CACHE_DB" "$TEST_USER_DUCKDB_DB" "$TEST_USER_DUCKDB_DB_WAL" "$TEST_USER_EXPORT_FILE"
}

test_command() {
    local cmd="$1"
    shift
    log "Testing: $cmd"
    if [ "$cmd" = "purge" ]; then
        # Special handling for purge command
        output=$(echo "y" | run_goku "$cmd" "$@" 2>&1)
    else
        output=$(run_goku "$cmd" "$@" 2>&1)
    fi
    exit_code=$?
    if [ $exit_code -eq 0 ]; then
        log "Success: $cmd"
        echo "$output"
    else
        log "Failed: $cmd"
        echo "Error output:"
        echo "$output"
        exit 1
    fi
}

test_command_test_user() {
    local cmd="$1"
    shift
    log "Testing for $TEST_USER: $cmd"
    if [ "$cmd" = "purge" ]; then
        # Special handling for purge command
        output=$(echo "y" | run_goku_test_user "$cmd" "$@" 2>&1)
    else
        output=$(run_goku_test_user "$cmd" "$@" 2>&1)
    fi
    exit_code=$?
    if [ $exit_code -eq 0 ]; then
        log "Success for $TEST_USER: $cmd"
        echo "$output"
    else
        log "Failed for $TEST_USER: $cmd"
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

# Run tests for default user
log "Testing with default user ($DEFAULT_USER)"
test_command add --url "https://github.com" --title "Example Site" --description "An github website" --tags "github,test"
test_command get --id 1
test_command update --id 1 --title "Updated github Site" --description "An updated github website" --tags "github,test,updated"
test_command list
test_command search --query "example"
test_command tags list
test_command stats
test_command export --output "$EXPORT_FILE"
test_command purge
test_command import --file "$EXPORT_FILE"
test_command import --file "$TXT_FILE"
test_command sync
test_command delete --id 1

# Run tests for test user
log "Testing with test user ($TEST_USER)"
test_command_test_user add --url "https://dev.to/" --title "dev Example" --description "An dev example" --tags "dev,test"
test_command_test_user get --id 1
test_command_test_user update --id 1 --title "Updated dev Example" --description "An updated dev example" --tags "dev,test,updated"
test_command_test_user list
test_command_test_user search --query "dev"
test_command_test_user tags list
test_command_test_user stats
test_command_test_user export --output "$TEST_USER_EXPORT_FILE"
test_command_test_user purge
test_command_test_user import --file "$TEST_USER_TXT_FILE"
test_command_test_user import --file "$TEST_USER_EXPORT_FILE"
test_command_test_user sync
test_command_test_user delete --id 1

# Verify separate databases
log "Verifying separate databases"
if [ -f "$TEST_DB" ] && [ -f "$TEST_USER_DB" ]; then
    log "Separate databases created successfully"
else
    log "Failed to create separate databases"
    exit 1
fi

log "All tests completed successfully!"
exit 0
