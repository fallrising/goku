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

func (d *Database) List(ctx context.Context) ([]*models.Bookmark, error) {
	query := `SELECT id, url, title, description, tags, created_at, updated_at FROM bookmarks`

	rows, err := d.db.QueryContext(ctx, query)
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