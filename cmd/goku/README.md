# Goku CLI Help Documentation

Goku CLI is a powerful command-line interface for managing bookmarks efficiently across multiple user profiles.

## Global Options

- `--db`: Path to the Goku database file (default: "<user>.db", env: GOKU_DB_PATH_<USER>)
- `--cache-db`: Path to the Goku cache database file (default: "<user>_cache.db", env: GOKU_CACHE_DB_PATH_<USER>)
- `--duckdb`: Path to the Goku DuckDB statistics file (default: "<user>_stats.duckdb", env: GOKU_DUCKDB_PATH_<USER>)
- `--user`: User profile to use (default: "goku", env: GOKU_USER)

## Commands

### add
Add a new bookmark

Usage: `goku [--user <user>] add [options]`

Options:
- `--url`: URL of the bookmark (required)
- `--title`: Title of the bookmark
- `--description`: Description of the bookmark
- `--tags`: Tags for the bookmark (comma-separated)
- `--fetch, -F`: Enable fetching additional data for the bookmark

### delete
Delete a bookmark

Usage: `goku [--user <user>] delete --id <bookmark_id>`

Options:
- `--id`: ID of the bookmark to delete (required)

### get
Get details of a specific bookmark

Usage: `goku [--user <user>] get --id <bookmark_id>`

Options:
- `--id`: ID of the bookmark to retrieve (required)

### list
List bookmarks with pagination

Usage: `goku [--user <user>] list [options]`

Options:
- `--limit`: Number of bookmarks to display per page (default: 10)
- `--offset`: Offset to start listing bookmarks from (default: 0)

### search
Search bookmarks

Usage: `goku [--user <user>] search [options] <query>`

Options:
- `--query, -q`: Search query (required)
- `--limit`: Number of results to display (default: 10)
- `--offset`: Offset for pagination (default: 0)

### update
Update an existing bookmark

Usage: `goku [--user <user>] update [options]`

Options:
- `--id`: ID of the bookmark to update (required)
- `--url`: New URL for the bookmark
- `--title`: New title for the bookmark
- `--description`: New description for the bookmark
- `--tags`: New tags for the bookmark (comma-separated)
- `--fetch, -F`: Enable fetching updated data for the bookmark

### import
Import bookmarks from a file

Usage: `goku [--user <user>] import [options]`

Options:
- `--file, -f`: Input file path (.html or .json) (required)
- `--workers, -w`: Number of worker goroutines for concurrent processing (default: 5)
- `--fetch, -F`: Enable fetching additional data for each imported bookmark

### export
Export bookmarks to a file

Usage: `goku [--user <user>] export [options]`

Options:
- `--output, -o`: Output file path (default: stdout)

### tags
Manage tags for bookmarks

Subcommands:
- `remove`: Remove a tag from a bookmark
  Usage: `goku [--user <user>] tags remove --id <bookmark_id> --tag <tag_name>`
- `list`: List all unique tags
  Usage: `goku [--user <user>] tags list`

### stats
Display bookmark statistics

Usage: `goku [--user <user>] stats`

### purge
Delete all bookmarks from the database

Usage: `goku [--user <user>] purge [options]`

Options:
- `--force`: Force purge without confirmation

### sync
Sync data from SQLite to DuckDB for statistics

Usage: `goku [--user <user>] sync`

### fetch
Fetch or update metadata for bookmarks

Usage: `goku [--user <user>] fetch [options]`

Options:
- `--id`: Fetch metadata for a specific bookmark ID
- `--all`: Fetch metadata for all bookmarks
- `--limit`: Number of bookmarks to process per batch (default: 10)
- `--skip-internal`: Skip URLs with internal IP addresses

For more detailed information on each command, use `goku <command> --help`.

## User Profiles

Goku CLI supports multiple user profiles. Each profile has its own set of databases. To use a specific profile, use the `--user` flag followed by the profile name. For example:
```
goku --user osint add --url "https://example.com" --title "OSINT Example"
```

This will add a bookmark to the "osint" user's database. If the user profile doesn't exist, it will be created automatically.

The default user profile is "goku" if no `--user` flag is specified.
