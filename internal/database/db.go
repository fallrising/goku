package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/fallrising/goku-cli/pkg/models"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{db: db}, nil
}

func (d *Database) Init() error {
	query := `CREATE TABLE IF NOT EXISTS bookmarks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		url TEXT NOT NULL,
		title TEXT,
		description TEXT,
		tags TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`

	_, err := d.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create bookmarks table: %w", err)
	}

	return nil
}

func (d *Database) Create(ctx context.Context, bookmark *models.Bookmark) error {
	query := `INSERT INTO bookmarks (url, title, description, tags) VALUES (?, ?, ?, ?)`
	tags := strings.Join(bookmark.Tags, ",")

	result, err := d.db.ExecContext(ctx, query, bookmark.URL, bookmark.Title, bookmark.Description, tags)
	if err != nil {
		return fmt.Errorf("failed to insert bookmark: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	bookmark.ID = id
	return nil
}

func (d *Database) GetByID(ctx context.Context, id int64) (*models.Bookmark, error) {
	query := `SELECT id, url, title, description, tags, created_at, updated_at FROM bookmarks WHERE id = ?`

	var bookmark models.Bookmark
	var tags string

	err := d.db.QueryRowContext(ctx, query, id).Scan(
		&bookmark.ID, &bookmark.URL, &bookmark.Title, &bookmark.Description,
		&tags, &bookmark.CreatedAt, &bookmark.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("bookmark not found")
		}
		return nil, fmt.Errorf("failed to get bookmark: %w", err)
	}

	bookmark.Tags = strings.Split(tags, ",")
	return &bookmark, nil
}

// GetByURL retrieves a bookmark by its URL
func (d *Database) GetByURL(ctx context.Context, url string) (*models.Bookmark, error) {
	query := `SELECT id, url, title, description, tags, created_at, updated_at FROM bookmarks WHERE url = ?`

	var bookmark models.Bookmark
	var tags string

	err := d.db.QueryRowContext(ctx, query, url).Scan(
		&bookmark.ID, &bookmark.URL, &bookmark.Title, &bookmark.Description,
		&tags, &bookmark.CreatedAt, &bookmark.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No bookmark found with the given URL
		}
		return nil, fmt.Errorf("failed to get bookmark by URL: %w", err)
	}

	// Split the tags string into a slice
	bookmark.Tags = strings.Split(tags, ",")
	return &bookmark, nil
}

func (d *Database) Update(ctx context.Context, bookmark *models.Bookmark) error {
	query := `UPDATE bookmarks SET url = ?, title = ?, description = ?, tags = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	tags := strings.Join(bookmark.Tags, ",")

	_, err := d.db.ExecContext(ctx, query, bookmark.URL, bookmark.Title, bookmark.Description, tags, bookmark.ID)
	if err != nil {
		return fmt.Errorf("failed to update bookmark: %w", err)
	}

	return nil
}

func (d *Database) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM bookmarks WHERE id = ?`

	_, err := d.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete bookmark: %w", err)
	}

	return nil
}

func (d *Database) List(ctx context.Context, limit, offset int) ([]*models.Bookmark, error) {
	query := `SELECT id, url, title, description, tags, created_at, updated_at FROM bookmarks LIMIT ? OFFSET ?`
	rows, err := d.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query bookmarks: %w", err)
	}
	defer rows.Close()

	var bookmarks []*models.Bookmark
	for rows.Next() {
		var bookmark models.Bookmark
		var tags string
		err := rows.Scan(
			&bookmark.ID, &bookmark.URL, &bookmark.Title, &bookmark.Description,
			&tags, &bookmark.CreatedAt, &bookmark.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bookmark row: %w", err)
		}
		bookmark.Tags = strings.Split(tags, ",")
		bookmarks = append(bookmarks, &bookmark)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bookmark rows: %w", err)
	}

	return bookmarks, nil
}

func (d *Database) Search(ctx context.Context, query string, limit, offset int) ([]*models.Bookmark, error) {
	searchQuery := `
		SELECT id, url, title, description, tags, created_at, updated_at 
		FROM bookmarks 
		WHERE url LIKE ? OR title LIKE ? OR description LIKE ? OR tags LIKE ?
		LIMIT ? OFFSET ?
	`
	searchParam := "%" + query + "%"

	rows, err := d.db.QueryContext(ctx, searchQuery, searchParam, searchParam, searchParam, searchParam, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search bookmarks: %w", err)
	}
	defer rows.Close()

	var bookmarks []*models.Bookmark
	for rows.Next() {
		var bookmark models.Bookmark
		var tags string
		err := rows.Scan(
			&bookmark.ID, &bookmark.URL, &bookmark.Title, &bookmark.Description,
			&tags, &bookmark.CreatedAt, &bookmark.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bookmark row: %w", err)
		}
		bookmark.Tags = strings.Split(tags, ",")
		bookmarks = append(bookmarks, &bookmark)
	}

	return bookmarks, nil
}

func (d *Database) ListAllTags(ctx context.Context) ([]string, error) {
	query := `SELECT tags FROM bookmarks`

	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query bookmarks for tags: %w", err)
	}
	defer rows.Close()

	tagSet := make(map[string]struct{}) // Use a set to deduplicate tags
	for rows.Next() {
		var tags string
		err := rows.Scan(&tags)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tags: %w", err)
		}

		// Split the comma-separated tags and add them to the set
		for _, tag := range strings.Split(tags, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tagSet[tag] = struct{}{}
			}
		}
	}

	// Convert the set back to a slice
	var uniqueTags []string
	for tag := range tagSet {
		uniqueTags = append(uniqueTags, tag)
	}

	return uniqueTags, nil
}

func (d *Database) PurgeAllData(ctx context.Context) error {
	query := `DELETE FROM bookmarks`

	_, err := d.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete all data: %w", err)
	}

	return nil
}

func (d *Database) CountByHostname(ctx context.Context) (map[string]int, error) {
	query := `SELECT 
		substr(url, instr(url, '://') + 3, 
			case 
				when instr(substr(url, instr(url, '://') + 3), '/') = 0 
				then length(substr(url, instr(url, '://') + 3)) 
				else instr(substr(url, instr(url, '://') + 3), '/') - 1 
			end
		) as hostname, 
		COUNT(*) as count 
	FROM bookmarks 
	GROUP BY hostname`

	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query hostnames: %w", err)
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

func (d *Database) CountByTag(ctx context.Context) (map[string]int, error) {
	query := `SELECT tag, COUNT(*) as count 
	FROM (
		SELECT trim(value) as tag
		FROM bookmarks
		CROSS JOIN json_each('["' || replace(replace(tags, ' ', ''), ',', '","') || '"]')
	)
	GROUP BY tag`

	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var tag string
		var count int
		if err := rows.Scan(&tag, &count); err != nil {
			return nil, fmt.Errorf("failed to scan tag count: %w", err)
		}
		counts[tag] = count
	}

	return counts, nil
}

func (d *Database) GetLatest(ctx context.Context, limit int) ([]*models.Bookmark, error) {
	query := `SELECT id, url, title, description, tags, created_at, updated_at 
	FROM bookmarks 
	ORDER BY created_at DESC 
	LIMIT ?`

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

func (d *Database) CountAccessibility(ctx context.Context) (map[string]int, error) {
	query := `SELECT 
		CASE 
			WHEN description LIKE 'Metadata fetch failed:%' THEN 'inaccessible'
			ELSE 'accessible'
		END as status, 
		COUNT(*) as count 
	FROM bookmarks 
	GROUP BY status`

	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query accessibility: %w", err)
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

func (d *Database) TopHostnames(ctx context.Context, limit int) ([]models.HostnameCount, error) {
	query := `SELECT 
		substr(url, instr(url, '://') + 3, 
			case 
				when instr(substr(url, instr(url, '://') + 3), '/') = 0 
				then length(substr(url, instr(url, '://') + 3)) 
				else instr(substr(url, instr(url, '://') + 3), '/') - 1 
			end
		) as hostname, 
		COUNT(*) as count 
	FROM bookmarks 
	GROUP BY hostname 
	ORDER BY count DESC 
	LIMIT ?`

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

func (d *Database) ListUniqueHostnames(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT
		substr(url, instr(url, '://') + 3, 
			case 
				when instr(substr(url, instr(url, '://') + 3), '/') = 0 
				then length(substr(url, instr(url, '://') + 3)) 
				else instr(substr(url, instr(url, '://') + 3), '/') - 1 
			end
		) as hostname
	FROM bookmarks 
	ORDER BY hostname`

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

func (d *Database) CountCreatedLastNDays(ctx context.Context, days int) (map[string]int, error) {
	query := `SELECT 
		date(created_at) as day, 
		COUNT(*) as count 
	FROM bookmarks 
	WHERE created_at >= date('now', ?)
	GROUP BY day 
	ORDER BY day DESC`

	rows, err := d.db.QueryContext(ctx, query, fmt.Sprintf("-%d days", days))
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

func (d *Database) Count(ctx context.Context) (int, error) {
	var count int
	err := d.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM bookmarks").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count bookmarks: %w", err)
	}
	return count, nil
}
