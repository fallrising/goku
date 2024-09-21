#!/bin/bash

set -e

# Create a temporary directory for the test database
TEST_DIR=$(mktemp -d)
TEST_DB="$TEST_DIR/test_goku.db"

# Cleanup function to remove the temporary directory
cleanup() {
    rm -rf "$TEST_DIR"
}

# Set up trap to ensure cleanup happens even if the script fails
trap cleanup EXIT

# Build the Goku CLI using build.sh
echo "Building Goku CLI..."
./build.sh

# Function to run Goku CLI commands with the test database
run_goku() {
    GOKU_DB_PATH="$TEST_DB" ./bin/goku "$@"
}

echo "Starting CRUD validation for Goku CLI"

# Create a bookmark
echo "1. Creating a bookmark"
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

# Read the updated bookmark
echo "4. Reading the updated bookmark"
run_goku get --id "$BOOKMARK_ID"

# List all bookmarks
echo "5. Listing all bookmarks"
run_goku list

# Delete the bookmark
echo "6. Deleting the bookmark"
run_goku delete --id "$BOOKMARK_ID"

# Try to read the deleted bookmark (should fail)
echo "7. Attempting to read the deleted bookmark (should fail)"
if run_goku get --id "$BOOKMARK_ID" 2>/dev/null; then
    echo "Error: Bookmark was not deleted successfully"
    exit 1
else
    echo "Bookmark deleted successfully"
fi

# List all bookmarks again (should be empty or not include the deleted bookmark)
echo "8. Listing all bookmarks after deletion"
run_goku list

echo "CRUD validation completed successfully"