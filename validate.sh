#!/bin/bash

set -e

# Configuration
TEST_DB="./goku.db"
CACHE_DB="./goku_cache.db"
GOKU_CMD="./bin/goku"
GOKU_LOG="./goku.log"

# Helper functions
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1"
}

run_goku() {
    GOKU_DB_PATH="$TEST_DB" $GOKU_CMD "$@"
}

cleanup() {
    log "Cleaning up..."
    rm -f "$TEST_DB"
    rm -f "$CACHE_DB"
    rm -f "$GOKU_LOG"
    rm -f "exported_bookmarks.html"
}

assert() {
    if [ $1 -ne 0 ]; then
        log "Assertion failed: $2"
        exit 1
    fi
}

# Test functions
test_build() {
    log "Building Goku CLI..."
    if ! ./build.sh; then
        log "Build failed. Exiting."
        exit 1
    fi
}

test_add_command() {
    log "Testing Add command"
    log "Running command: $GOKU_CMD add --url \"https://example.com\" --title \"Example Site\" --description \"An example website\" --tags \"example,test\""
    CREATE_OUTPUT=$(run_goku add --url "https://example.com" --title "Example Site" --description "An example website" --tags "example,test")
    log "Command output: $CREATE_OUTPUT"
    BOOKMARK_ID=$(echo "$CREATE_OUTPUT" | sed -n 's/.*ID: \([0-9]*\).*/\1/p')
    log "Extracted bookmark ID: $BOOKMARK_ID"
    assert $? "Failed to add bookmark"
    log "Created bookmark with ID: $BOOKMARK_ID"
}

test_get_command() {
    log "Testing Get command"
    run_goku get --id "$BOOKMARK_ID"
    assert $? "Failed to get bookmark"
}

test_update_command() {
    log "Testing Update command"
    run_goku update --id "$BOOKMARK_ID" --title "Updated Example Site" --description "An updated example website" --tags "example,test,updated"
    assert $? "Failed to update bookmark"
}

test_list_command() {
    log "Testing List command"
    run_goku list
    assert $? "Failed to list bookmarks"
}

test_search_command() {
    log "Testing Search command"
    SEARCH_OUTPUT=$(run_goku search --query "Updated")
    echo "$SEARCH_OUTPUT" | grep -q "Updated Example Site"
    assert $? "Search failed"
}

test_delete_command() {
    log "Testing Delete command"
    run_goku delete --id "$BOOKMARK_ID"
    assert $? "Failed to delete bookmark"

    # Verify deletion
    if run_goku get --id "$BOOKMARK_ID" &>/dev/null; then
        log "Error: Bookmark was not properly deleted"
        exit 1
    else
        log "Bookmark was properly deleted"
    fi
}

test_tags_command() {
    log "Testing Tags command"

    # Add a bookmark with tags
    run_goku add --url "https://tagtest.com" --title "Tag Test" --tags "test,tags"
    assert $? "Failed to add bookmark for tag test"

    # List tags
    TAG_OUTPUT=$(run_goku tags list)
    echo "$TAG_OUTPUT" | grep -q "test"
    assert $? "Failed to list tags"

    # Remove a tag
    run_goku tags remove --id 1 --tag "test"
    assert $? "Failed to remove tag"

    # Verify tag removal
    TAG_OUTPUT=$(run_goku tags list)
    echo "$TAG_OUTPUT" | grep -qv "test"
    assert $? "Tag was not properly removed"
}

test_stats_command() {
    log "Testing Stats command"
    STATS_OUTPUT=$(run_goku stats)
    echo "$STATS_OUTPUT" | grep -q "Bookmark Statistics"
    assert $? "Failed to generate statistics"
}

test_add_again_command() {
    log "Testing Add command"
    log "Running command: $GOKU_CMD add --url \"https://google.com\" --title \"Example Site\" --description \"An example website\" --tags \"example,test\""
    CREATE_OUTPUT=$(run_goku add --url "https://google.com" --title "Example Site" --description "An example website" --tags "example,test")
    log "Command output: $CREATE_OUTPUT"
    BOOKMARK_ID=$(echo "$CREATE_OUTPUT" | sed -n 's/.*ID: \([0-9]*\).*/\1/p')
    log "Extracted bookmark ID: $BOOKMARK_ID"
    assert $? "Failed to add bookmark"
    log "Created bookmark with ID: $BOOKMARK_ID"
}

test_import_commands() {
    log "Testing Import command"
    run_goku import --file "exported_bookmarks.html"
    assert $? "Failed to import bookmarks"
}

test_export_commands() {
    log "Testing Export command"
    run_goku export --output "exported_bookmarks.html"
    assert $? "Failed to export bookmarks"
}

test_purge_command() {
    log "Testing Purge command"
    echo "y" | run_goku purge
    assert $? "Failed to purge bookmarks"

    # Verify purge
    BOOKMARK_COUNT=$(run_goku list | grep -c "ID:")
    if [ "$BOOKMARK_COUNT" -ne 0 ]; then
        log "Error: Bookmarks were not properly purged"
        exit 1
    else
        log "All bookmarks were properly purged"
    fi
}

# Main execution
trap cleanup EXIT

log "Starting comprehensive validation for Goku CLI"

cleanup
test_build
test_add_command
test_get_command
test_update_command
test_list_command
test_search_command
test_tags_command
test_stats_command
test_delete_command
test_add_again_command
test_export_commands
test_purge_command
test_import_commands

log "All tests completed successfully!"
exit 0