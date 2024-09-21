#!/bin/bash

set -e

# Create a temporary directory for the test database
TEST_DB="./goku.db"

cleanup() {
    echo "Cleaning up..."
    rm -f "$TEST_DB"
}

# Set up trap to ensure cleanup happens even if the script fails
trap cleanup EXIT

# Build the Goku CLI using build.sh
echo "Building Goku CLI..."
if ! ./build.sh; then
    echo "Build failed. Exiting."
    exit 1
fi

# Function to run Goku CLI commands with the test database
run_goku() {
    GOKU_DB_PATH="$TEST_DB" ./bin/goku "$@"
}

echo "Starting CRUD and Search validation for Goku CLI"

# Test 1: Basic CRUD operations
echo "1. Creating a bookmark with all details provided"
CREATE_OUTPUT=$(run_goku add --url "https://example.com" --title "Example Site" --description "An example website" --tags "example,test")
echo "$CREATE_OUTPUT"
BOOKMARK_ID=$(echo "$CREATE_OUTPUT" | sed -n 's/.*ID: \([0-9]*\).*/\1/p')

if [ -z "$BOOKMARK_ID" ]; then
    echo "Failed to extract bookmark ID"
    exit 1
fi

echo "Created bookmark with ID: $BOOKMARK_ID"

# Read the bookmark
echo "2. Reading the created bookmark"
run_goku get --id "$BOOKMARK_ID"

# Update the bookmark
echo "3. Updating the bookmark"
run_goku update --id "$BOOKMARK_ID" --title "Updated Example Site" --description "An updated example website" --tags "example,test,updated"

echo "3.1 Testing partial update"
run_goku update --id "$BOOKMARK_ID" --title "Partially Updated Example Site"
PARTIAL_UPDATE_OUTPUT=$(run_goku get --id "$BOOKMARK_ID")
if ! echo "$PARTIAL_UPDATE_OUTPUT" | grep -q "Title:Partially Updated Example Site"; then
    echo "Error: Partial update of title failed"
    exit 1
fi
if ! echo "$PARTIAL_UPDATE_OUTPUT" | grep -q "Description:An updated example website"; then
    echo "Error: Partial update changed description unexpectedly"
    exit 1
fi

# Read the updated bookmark
echo "4. Reading the updated bookmark"
run_goku get --id "$BOOKMARK_ID"

# List all bookmarks
echo "5. Listing all bookmarks"
run_goku list

# Test 2: Search functionality
echo "6. Testing search functionality"

# Search by title
echo "6.1 Searching by title"
SEARCH_OUTPUT=$(run_goku search --query "Updated Example")
echo "$SEARCH_OUTPUT"
if ! echo "$SEARCH_OUTPUT" | grep -q "Updated Example Site"; then
    echo "Error: Search by title failed"
    exit 1
fi

# Search by URL
echo "6.2 Searching by URL"
SEARCH_OUTPUT=$(run_goku search --query "example.com")
echo "$SEARCH_OUTPUT"
if ! echo "$SEARCH_OUTPUT" | grep -q "https://example.com"; then
    echo "Error: Search by URL failed"
    exit 1
fi

# Search by description
echo "6.3 Searching by description"
SEARCH_OUTPUT=$(run_goku search --query "updated example")
echo "$SEARCH_OUTPUT"
if ! echo "$SEARCH_OUTPUT" | grep -q "updated example"; then
    echo "Error: Search by description failed"
    exit 1
fi

# Search by tag
echo "6.4 Searching by tag"
SEARCH_OUTPUT=$(run_goku search --query "updated")
echo "$SEARCH_OUTPUT"
if ! echo "$SEARCH_OUTPUT" | grep -q "example test updated"; then
    echo "Error: Search by tag failed"
    exit 1
fi

# Search with no results
echo "6.5 Searching with no results"
SEARCH_OUTPUT=$(run_goku search --query "nonexistent")
echo "$SEARCH_OUTPUT"
if ! echo "$SEARCH_OUTPUT" | grep -q "No bookmarks found matching the query"; then
    echo "Error: Search with no results failed"
    exit 1
fi

# Delete the bookmark
echo "7. Deleting the bookmark"
run_goku delete --id "$BOOKMARK_ID"

# Try to read the deleted bookmark (should fail)
echo "8. Attempting to read the deleted bookmark (should fail)"
if run_goku get --id "$BOOKMARK_ID" 2>/dev/null; then
    echo "Error: Bookmark was not deleted successfully"
    exit 1
else
    echo "Bookmark deleted successfully"
fi

# Test 3: Automatic webpage content extraction
echo "9. Creating a bookmark with only URL (testing automatic content extraction)"
CREATE_OUTPUT=$(run_goku add --url "https://www.example.com")
echo "$CREATE_OUTPUT"
BOOKMARK_ID=$(echo "$CREATE_OUTPUT" | sed -n 's/.*ID: \([0-9]*\).*/\1/p')

if [ -z "$BOOKMARK_ID" ]; then
    echo "Failed to extract bookmark ID"
    exit 1
fi

echo "Created bookmark with ID: $BOOKMARK_ID"

# Read the bookmark to verify automatic content extraction
echo "10. Reading the automatically extracted bookmark content"
BOOKMARK_CONTENT=$(run_goku get --id "$BOOKMARK_ID")
echo "$BOOKMARK_CONTENT"

# Verify that title was extracted (this should always be present)
if echo "$BOOKMARK_CONTENT" | grep -q "Title:Example Domain"; then
    echo "Title extraction successful"
else
    echo "Error: Title extraction failed"
    echo "Expected title 'Example Domain' not found in output:"
    echo "$BOOKMARK_CONTENT"
    exit 1
fi

# Check if description was extracted (might be empty)
if echo "$BOOKMARK_CONTENT" | grep -q "Description:"; then
    echo "Description field is present"
    DESCRIPTION=$(echo "$BOOKMARK_CONTENT" | sed -n 's/.*Description:\([^,]*\).*/\1/p' | tr -d '[:space:]')
    if [ -n "$DESCRIPTION" ]; then
        echo "Description extraction successful: $DESCRIPTION"
    else
        echo "Description is empty, but this is acceptable"
    fi
else
    echo "Error: Description field is missing"
    exit 1
fi

# Check if tags were extracted (might be empty)
if echo "$BOOKMARK_CONTENT" | grep -q "Tags:"; then
    echo "Tags field is present"
    TAGS=$(echo "$BOOKMARK_CONTENT" | sed -n 's/.*Tags:\([^]]*\).*/\1/p' | tr -d '[:space:]')
    if [ -n "$TAGS" ]; then
        echo "Tags extraction successful: $TAGS"
    else
        echo "Tags are empty, but this is acceptable"
    fi
else
    echo "Error: Tags field is missing"
    exit 1
fi

# Delete the test bookmark
echo "11. Deleting the test bookmark"
run_goku delete --id "$BOOKMARK_ID"

# List all bookmarks again (should be empty)
echo "12. Listing all bookmarks after deletion"
run_goku list

echo "13. Testing invalid input handling"
if run_goku add --url "not_a_valid_url" 2>/dev/null; then
    echo "Error: Adding invalid URL should fail"
    exit 1
fi
echo "Invalid input handling test passed"

echo "CRUD, Search, and automatic content extraction test completed successfully"
