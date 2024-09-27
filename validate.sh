#!/bin/bash

set -e

# Configuration
TEST_DB="./goku.db"
GOKU_CMD="./bin/goku"

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

test_basic_crud() {
    log "Testing basic CRUD operations"

    log "1. Creating a bookmark with all details provided"
    CREATE_OUTPUT=$(run_goku add --url "https://example.com" --title "Example Site" --description "An example website" --tags "example,test")
    BOOKMARK_ID=$(echo "$CREATE_OUTPUT" | sed -n 's/.*ID: \([0-9]*\).*/\1/p')
    assert $? "Failed to extract bookmark ID"
    log "Created bookmark with ID: $BOOKMARK_ID"

    log "2. Reading the created bookmark"
    run_goku get --id "$BOOKMARK_ID"
    assert $? "Failed to read bookmark"

    log "3. Updating the bookmark"
    run_goku update --id "$BOOKMARK_ID" --title "Updated Example Site" --description "An updated example website" --tags "example,test,updated"
    assert $? "Failed to update bookmark"

    log "4. Reading the updated bookmark"
    run_goku get --id "$BOOKMARK_ID"
    assert $? "Failed to read updated bookmark"

    log "5. Listing all bookmarks"
    run_goku list
    assert $? "Failed to list bookmarks"
}

test_partial_updates() {
    log "Testing partial updates"

    log "3.1 Testing partial update for title"
    run_goku update --id "$BOOKMARK_ID" --title "Partially Updated Example Site"
    PARTIAL_UPDATE_OUTPUT=$(run_goku get --id "$BOOKMARK_ID")
    echo "$PARTIAL_UPDATE_OUTPUT" | grep -q "Title:Partially Updated Example Site"
    assert $? "Partial update of title failed"
    echo "$PARTIAL_UPDATE_OUTPUT" | grep -q "Description:An updated example website"
    assert $? "Partial update changed description unexpectedly"

    log "3.2 Testing partial update for URL"
    run_goku update --id "$BOOKMARK_ID" --url "https://www.yahoo.com"
    PARTIAL_UPDATE_OUTPUT=$(run_goku get --id "$BOOKMARK_ID")
    echo "$PARTIAL_UPDATE_OUTPUT" | grep -q "URL:https://www.yahoo.com"
    assert $? "Partial update of URL failed"

    log "3.3 Testing partial update for tags"
    run_goku update --id "$BOOKMARK_ID" --tags "updated,test"
    PARTIAL_UPDATE_OUTPUT=$(run_goku get --id "$BOOKMARK_ID")
    echo "$PARTIAL_UPDATE_OUTPUT" | grep -q "updated test"
    assert $? "Partial update of tags failed"
}

test_search() {
    log "Testing search functionality"

    log "6.1 Searching by title"
    SEARCH_OUTPUT=$(run_goku search --query "Updated")
    echo "$SEARCH_OUTPUT" | grep -q "updated test"
    assert $? "Search by title failed"

    log "6.2 Searching by URL"
    SEARCH_OUTPUT=$(run_goku search --query "yahoo.com")
    echo "$SEARCH_OUTPUT" | grep -q "https://www.yahoo.com"
    assert $? "Search by URL failed"

    log "6.3 Searching by description"
    SEARCH_OUTPUT=$(run_goku search --query "updated example")
    echo "$SEARCH_OUTPUT" | grep -q "An updated example website"
    assert $? "Search by description failed"

    log "6.4 Searching by tag"
    SEARCH_OUTPUT=$(run_goku search --query "updated")
    echo "$SEARCH_OUTPUT" | grep -q "updated test"
    assert $? "Search by tag failed"

    log "6.5 Searching with no results"
    SEARCH_OUTPUT=$(run_goku search --query "nonexistent")
    echo "$SEARCH_OUTPUT" | grep -q "No bookmarks found matching the query"
    assert $? "Search with no results failed"
}

test_tag_management() {
    log "Testing tag management"

    log "7.1 Removing a tag from the bookmark"
    run_goku tags remove --id "$BOOKMARK_ID" --tag "test"
    TAG_OUTPUT=$(run_goku get --id "$BOOKMARK_ID")
    echo "$TAG_OUTPUT" | grep -q "test"
    assert $? "Failed to remove tag"

    log "7.2 Listing all unique tags"
    run_goku tags list
    assert $? "Failed to list tags"
}

test_import_export() {
    log "Testing import and export functionality"

    log "8. Exporting bookmarks"
    run_goku export --output "exported_bookmarks.html"
    assert $? "Failed to export bookmarks"

    log "9. Importing bookmarks"
    run_goku import --file "exported_bookmarks.html"
    assert $? "Failed to import bookmarks"

    rm -f "exported_bookmarks.html"
}

test_statistics() {
    log "Testing statistics functionality"

    log "10. Generating statistics"
    run_goku stats
    assert $? "Failed to generate statistics"
}

# Main execution
trap cleanup EXIT

log "Starting CRUD, Search, and additional feature validation for Goku CLI"

test_build
test_basic_crud
test_partial_updates
test_search
test_tag_management
test_import_export
test_statistics

log "All tests completed successfully!"
exit 0