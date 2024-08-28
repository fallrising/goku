#!/bin/bash

# Define file names
IMPORT_FILE="import_bookmarks.json"
EXPORT_FILE="export_bookmarks.json"

# Create a sample JSON file with bookmarks
cat <<EOL > $IMPORT_FILE
[
    {
        "id": 1,
        "url": "https://example.com",
        "title": "Example",
        "description": "An example bookmark",
        "tags": "example, test",
        "created_at": "2023-01-01T00:00:00Z",
        "updated_at": "2023-01-01T00:00:00Z"
    },
    {
        "id": 2,
        "url": "https://example.org",
        "title": "Example Org",
        "description": "Another example bookmark",
        "tags": "example, org",
        "created_at": "2023-01-02T00:00:00Z",
        "updated_at": "2023-01-02T00:00:00Z"
    }
]
EOL

# Import bookmarks from the JSON file
./build/goku-cli import $IMPORT_FILE

# Export bookmarks to a new JSON file
./build/goku-cli export $EXPORT_FILE

# Compare the original and exported JSON files
if diff $IMPORT_FILE $EXPORT_FILE; then
    echo "Import and export test passed!"
else
    echo "Import and export test failed!"
fi

# Clean up
rm $IMPORT_FILE $EXPORT_FILE
