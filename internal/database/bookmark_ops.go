package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/fallrising/goku-cli/pkg/models"
)

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

func (d *Database) Count(ctx context.Context) (int, error) {
	var count int
	err := d.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM bookmarks").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count bookmarks: %w", err)
	}
	return count, nil
}
