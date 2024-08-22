package database

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/fallrising/goku/internal/bookmark" // Update import path
	_ "github.com/mattn/go-sqlite3"                // Import SQLite driver
)

// Db is the database connection, now in the database package
var Db *sql.DB

// InitDB initializes the database connection and creates the table if it doesn't exist.
func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "bookmarks.db")
	if err != nil {
		return nil, err
	}

	// Create the bookmarks table
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS bookmarks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		url TEXT NOT NULL UNIQUE,
		title TEXT,
		description TEXT,
		tags TEXT
	);
	`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, fmt.Errorf("error creating bookmarks table: %w", err)
	}

	return db, nil
}

// AddBookmark adds a new bookmark to the database.
func AddBookmark(db *sql.DB, bookmark *bookmark.Bookmark) error { // Update type
	// Convert tags slice to a comma-separated string for storage
	tagsString := strings.Join(bookmark.Tags, ",")

	stmt, err := db.Prepare("INSERT INTO bookmarks(url, title, description, tags) VALUES(?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(bookmark.URL, bookmark.Title, bookmark.Description, tagsString)
	if err != nil {
		return fmt.Errorf("error executing statement: %w", err)
	}

	return nil
}

// GetAllBookmarks retrieves all bookmarks from the database.
func GetAllBookmarks(db *sql.DB) ([]*bookmark.Bookmark, error) { // Update type
	rows, err := db.Query("SELECT id, url, title, description, tags FROM bookmarks")
	if err != nil {
		return nil, fmt.Errorf("error querying bookmarks: %w", err)
	}
	defer rows.Close()

	var bookmarks []*bookmark.Bookmark // Update type
	for rows.Next() {
		var bm bookmark.Bookmark // Update type
		var tagsString string
		err := rows.Scan(&bm.ID, &bm.URL, &bm.Title, &bm.Description, &tagsString)
		if err != nil {
			return nil, fmt.Errorf("error scanning bookmark row: %w", err)
		}

		// Convert comma-separated string back to tags slice
		bm.Tags = strings.Split(tagsString, ",")
		bookmarks = append(bookmarks, &bm)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return bookmarks, nil
}

// SearchBookmarks searches for bookmarks matching a keyword in the URL, title, or tags.
func SearchBookmarks(db *sql.DB, keyword string) ([]*bookmark.Bookmark, error) { // Update type
	// Use a LIKE query for basic keyword search
	query := fmt.Sprintf("SELECT id, url, title, description, tags FROM bookmarks WHERE url LIKE '%%%s%%' OR title LIKE '%%%s%%' OR tags LIKE '%%%s%%'", keyword, keyword, keyword)
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error searching bookmarks: %w", err)
	}
	defer rows.Close()

	var bookmarks []*bookmark.Bookmark // Update type
	for rows.Next() {
		var bm bookmark.Bookmark // Update type
		var tagsString string
		err := rows.Scan(&bm.ID, &bm.URL, &bm.Title, &bm.Description, &tagsString)
		if err != nil {
			return nil, fmt.Errorf("error scanning bookmark row: %w", err)
		}

		// Convert comma-separated string back to tags slice
		bm.Tags = strings.Split(tagsString, ",")
		bookmarks = append(bookmarks, &bm)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return bookmarks, nil
}
