package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/fallrising/goku-cli/pkg/models"
	_ "github.com/marcboeker/go-duckdb" // This line is crucial
)

type DuckDBStats struct {
	db *sql.DB
}

func NewDuckDBStats(dbPath string) (*DuckDBStats, error) {
	db, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open DuckDB: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping DuckDB: %w", err)
	}

	return &DuckDBStats{db: db}, nil
}

func (d *DuckDBStats) Init() error {
	_, err := d.db.Exec(`
		CREATE TABLE IF NOT EXISTS bookmarks (
			id INTEGER PRIMARY KEY,
			url TEXT NOT NULL,
			title TEXT,
			description TEXT,
			tags TEXT,
			created_at TIMESTAMP,
			updated_at TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create bookmarks table in DuckDB: %w", err)
	}
	return nil
}

func (d *DuckDBStats) SyncFromSQLite(sqliteDB *Database) error {
	// Start a transaction
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Clear existing data
	_, err = tx.Exec("DELETE FROM bookmarks")
	if err != nil {
		return fmt.Errorf("failed to clear existing data: %w", err)
	}

	// Fetch all bookmarks from SQLite
	bookmarks, err := sqliteDB.List(context.Background(), -1, 0) // Fetch all bookmarks
	if err != nil {
		return fmt.Errorf("failed to fetch bookmarks from SQLite: %w", err)
	}

	// Insert bookmarks into DuckDB
	stmt, err := tx.Prepare(`
		INSERT INTO bookmarks (id, url, title, description, tags, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	for _, b := range bookmarks {
		_, err = stmt.Exec(b.ID, b.URL, b.Title, b.Description, strings.Join(b.Tags, ","), b.CreatedAt, b.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to insert bookmark: %w", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Synced %d bookmarks to DuckDB", len(bookmarks))
	return nil
}

func (d *DuckDBStats) GetStatistics(ctx context.Context) (*models.Statistics, error) {
	stats := &models.Statistics{
		HostnameCounts:      make(map[string]int),
		TagCounts:           make(map[string]int),
		AccessibilityCounts: make(map[string]int),
		CreatedLastWeek:     make(map[string]int),
	}

	var err error

	// Top Hostnames
	stats.TopHostnames, err = d.getTopHostnames(ctx, 3)
	if err != nil {
		return nil, err
	}

	// Hostname Counts
	stats.HostnameCounts, err = d.getHostnameCounts(ctx)
	if err != nil {
		return nil, err
	}

	// Tag Counts
	stats.TagCounts, err = d.getTagCounts(ctx)
	if err != nil {
		return nil, err
	}

	// Latest Bookmarks
	stats.LatestBookmarks, err = d.getLatestBookmarks(ctx, 10)
	if err != nil {
		return nil, err
	}

	// Accessibility Counts
	stats.AccessibilityCounts, err = d.getAccessibilityCounts(ctx)
	if err != nil {
		return nil, err
	}

	// Unique Hostnames
	stats.UniqueHostnames, err = d.getUniqueHostnames(ctx)
	if err != nil {
		return nil, err
	}

	// Created Last Week
	stats.CreatedLastWeek, err = d.getCreatedLastWeek(ctx)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (d *DuckDBStats) getTopHostnames(ctx context.Context, limit int) ([]models.HostnameCount, error) {
	query := `
		SELECT 
			regexp_extract(url, '^(?:https?:\/\/)?(?:[^@\n]+@)?(?:www\.)?([^:\/\n?]+)', 1) as hostname,
			COUNT(*) as count 
		FROM bookmarks 
		GROUP BY hostname
		ORDER BY count DESC
		LIMIT ?
	`
	rows, err := d.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top hostnames: %w", err)
	}
	defer rows.Close()

	var topHostnames []models.HostnameCount
	for rows.Next() {
		var hc models.HostnameCount
		if err := rows.Scan(&hc.Hostname, &hc.Count); err != nil {
			return nil, fmt.Errorf("failed to scan hostname count: %w", err)
		}
		topHostnames = append(topHostnames, hc)
	}

	return topHostnames, nil
}

func (d *DuckDBStats) getHostnameCounts(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT 
			regexp_extract(url, '^(?:https?:\/\/)?(?:[^@\n]+@)?(?:www\.)?([^:\/\n?]+)', 1) as hostname,
			COUNT(*) as count 
		FROM bookmarks 
		GROUP BY hostname
	`
	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query hostname counts: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var hostname string
		var count int
		if err := rows.Scan(&hostname, &count); err != nil {
			return nil, fmt.Errorf("failed to scan hostname count: %w", err)
		}
		counts[hostname] = count
	}

	return counts, nil
}

func (d *DuckDBStats) getTagCounts(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT unnest.tag AS tag, COUNT(*) as count
		FROM bookmarks, UNNEST(string_split(tags, ',')) as unnest(tag)
		GROUP BY tag
		ORDER BY count DESC;
    `
	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tag counts: %w", err)
	}
	defer rows.Close()

	// Map to hold the counts
	counts := make(map[string]int)

	// Retrieve column names for debugging (optional)
	columnNames, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get column names: %w", err)
	}
	fmt.Printf("Columns: %v\n", columnNames)

	// Loop through the rows
	for rows.Next() {
		var tag string
		var count int64 // Using int64 since COUNT can return large numbers

		// Scan into the variables
		if err := rows.Scan(&tag, &count); err != nil {
			return nil, fmt.Errorf("failed to scan tag count: %w", err)
		}

		// Print the values for debugging
		fmt.Printf("Tag: %s, Count: %d\n", tag, count)

		// Trim the tag and store in the map
		counts[strings.TrimSpace(tag)] = int(count)
	}

	// Check for any errors during row iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return counts, nil
}

func (d *DuckDBStats) getLatestBookmarks(ctx context.Context, limit int) ([]*models.Bookmark, error) {
	query := `
		SELECT id, url, title, description, tags, created_at, updated_at
		FROM bookmarks
		ORDER BY created_at DESC
		LIMIT ?
	`
	rows, err := d.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query latest bookmarks: %w", err)
	}
	defer rows.Close()

	var bookmarks []*models.Bookmark
	for rows.Next() {
		var b models.Bookmark
		var tags string
		if err := rows.Scan(&b.ID, &b.URL, &b.Title, &b.Description, &tags, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan bookmark: %w", err)
		}
		b.Tags = strings.Split(tags, ",")
		bookmarks = append(bookmarks, &b)
	}

	return bookmarks, nil
}

func (d *DuckDBStats) getAccessibilityCounts(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT 
			CASE 
				WHEN description LIKE 'Metadata fetch failed:%' THEN 'inaccessible'
				ELSE 'accessible'
			END as status, 
			COUNT(*) as count 
		FROM bookmarks 
		GROUP BY status
	`
	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query accessibility counts: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan accessibility count: %w", err)
		}
		counts[status] = count
	}

	return counts, nil
}

func (d *DuckDBStats) getUniqueHostnames(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT regexp_extract(url, '^(?:https?:\/\/)?(?:[^@\n]+@)?(?:www\.)?([^:\/\n?]+)', 1) as hostname
		FROM bookmarks
		ORDER BY hostname
	`
	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query unique hostnames: %w", err)
	}
	defer rows.Close()

	var hostnames []string
	for rows.Next() {
		var hostname string
		if err := rows.Scan(&hostname); err != nil {
			return nil, fmt.Errorf("failed to scan hostname: %w", err)
		}
		hostnames = append(hostnames, hostname)
	}

	return hostnames, nil
}

func (d *DuckDBStats) getCreatedLastWeek(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT 
			strftime(created_at, '%Y-%m-%d') as day, 
			COUNT(*) as count 
		FROM bookmarks 
		WHERE created_at >= current_date - interval '7 days'
		GROUP BY day 
		ORDER BY day DESC
	`
	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query created counts: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var day string
		var count int
		if err := rows.Scan(&day, &count); err != nil {
			return nil, fmt.Errorf("failed to scan day count: %w", err)
		}
		counts[day] = count
	}

	return counts, nil
}
