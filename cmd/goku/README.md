# Goku CLI Help Documentation

Goku CLI is a powerful command-line interface for managing bookmarks efficiently.

## Global Options

- `--db`: Path to the Goku database file (default: "goku.db", env: GOKU_DB_PATH)
- `--cache-db`: Path to the Goku cache database file (default: "goku_cache.db", env: GOKU_CACHE_DB_PATH)
- `--duckdb`: Path to the Goku DuckDB statistics file (default: "goku_stats.duckdb", env: GOKU_DUCKDB_PATH)

## Commands

### add
Add a new bookmark

Usage: `goku add [options]`

Options:
- `--url`: URL of the bookmark (required)
- `--title`: Title of the bookmark
- `--description`: Description of the bookmark
- `--tags`: Tags for the bookmark (comma-separated)
- `--fetch, -F`: Enable fetching additional data for the bookmark

### delete
Delete a bookmark

Usage: `goku delete --id <bookmark_id>`

Options:
- `--id`: ID of the bookmark to delete (required)

### get
Get details of a specific bookmark

Usage: `goku get --id <bookmark_id>`

Options:
- `--id`: ID of the bookmark to retrieve (required)

### list
List bookmarks with pagination

Usage: `goku list [options]`

Options:
- `--limit`: Number of bookmarks to display per page (default: 10)
- `--offset`: Offset to start listing bookmarks from (default: 0)

### search
Search bookmarks

Usage: `goku search [options] <query>`

Options:
- `--query, -q`: Search query (required)
- `--limit`: Number of results to display (default: 10)
- `--offset`: Offset for pagination (default: 0)

### update
Update an existing bookmark

Usage: `goku update [options]`

Options:
- `--id`: ID of the bookmark to update (required)
- `--url`: New URL for the bookmark
- `--title`: New title for the bookmark
- `--description`: New description for the bookmark
- `--tags`: New tags for the bookmark (comma-separated)
- `--fetch, -F`: Enable fetching updated data for the bookmark

### import
Import bookmarks from a file

Usage: `goku import [options]`

Options:
- `--file, -f`: Input file path (.html or .json) (required)
- `--workers, -w`: Number of worker goroutines for concurrent processing (default: 5)
- `--fetch, -F`: Enable fetching additional data for each imported bookmark

### export
Export bookmarks to a file

Usage: `goku export [options]`

Options:
- `--output, -o`: Output file path (default: stdout)

### tags
Manage tags for bookmarks

Subcommands:
- `remove`: Remove a tag from a bookmark
  Usage: `goku tags remove --id <bookmark_id> --tag <tag_name>`
- `list`: List all unique tags
  Usage: `goku tags list`

### stats
Display bookmark statistics

Usage: `goku stats`

### purge
Delete all bookmarks from the database

Usage: `goku purge [options]`

Options:
- `--force`: Force purge without confirmation

### sync
Sync data from SQLite to DuckDB for statistics

Usage: `goku sync`

### fetch
Fetch or update metadata for bookmarks

Usage: `goku fetch [options]`

Options:
- `--id`: Fetch metadata for a specific bookmark ID
- `--all`: Fetch metadata for all bookmarks
- `--limit`: Number of bookmarks to process per batch (default: 10)
- `--skip-internal`: Skip URLs with internal IP addresses

For more detailed information on each command, use `goku <command> --help`.