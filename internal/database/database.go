package database

import (
	"database/sql"
	"fmt"
	"io"
	"strings"

	"github.com/fallrising/goku/internal/bookmark" // Update import path
	_ "github.com/mattn/go-sqlite3"                // Import SQLite driver
	"golang.org/x/net/html"
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
func SearchBookmarks(db *sql.DB, keyword string, fields ...string) ([]*bookmark.Bookmark, error) { // Update type
	// Use a LIKE query for basic keyword search
	// Construct the WHERE clause based on specified fields
	whereClause := "WHERE "
	for i, field := range fields {
		if i > 0 {
			whereClause += " OR " // Add OR for multiple fields
		}
		whereClause += fmt.Sprintf("%s LIKE '%%%s%%'", field, keyword)
	}

	query := fmt.Sprintf("SELECT id, url, title, description, tags FROM bookmarks %s", whereClause)
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

// DeleteBookmark deletes a bookmark by its ID.
func DeleteBookmark(db *sql.DB, id int) error {
	stmt, err := db.Prepare("DELETE FROM bookmarks WHERE id = ?")
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("error executing statement: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no bookmark found with ID %d", id)
	}

	return nil
}

func GetBookmarkByID(db *sql.DB, id int) (*bookmark.Bookmark, error) {
	row := db.QueryRow("SELECT id, url, title, description, tags FROM bookmarks WHERE id = ?", id)

	var bm bookmark.Bookmark
	var tagsString string
	err := row.Scan(&bm.ID, &bm.URL, &bm.Title, &bm.Description, &tagsString)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("bookmark with ID %d not found", id)
		}
		return nil, fmt.Errorf("error scanning bookmark row: %w", err)
	}

	// Convert comma-separated string back to tags slice
	bm.Tags = strings.Split(tagsString, ",")
	return &bm, nil
}

// UpdateBookmark updates an existing bookmark.
func UpdateBookmark(db *sql.DB, bookmark *bookmark.Bookmark) error {
	stmt, err := db.Prepare("UPDATE bookmarks SET url = ?, title = ?, description = ?, tags = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(bookmark.URL, bookmark.Title, bookmark.Description, bookmark.GetTagsString(), bookmark.ID)
	if err != nil {
		return fmt.Errorf("error executing statement: %w", err)
	}

	return nil
}

// ImportBookmarksFromHTML imports bookmarks from an HTML file.
func ImportBookmarksFromHTML(db *sql.DB, reader io.Reader) error {
	z := html.NewTokenizer(reader)

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, return
			return nil
		case tt == html.StartTagToken:
			t := z.Token()

			isAnchor := t.Data == "a"
			if !isAnchor {
				continue
			}

			// Extract href (URL) and text content (Title)
			var href, title string
			for _, a := range t.Attr {
				if a.Key == "href" {
					href = a.Val
				}
			}

			// Get the text content of the anchor tag
			if z.Next() == html.TextToken {
				title = strings.TrimSpace(z.Token().Data)
			}

			if href != "" {
				// Add bookmark to the database
				newBookmark := &bookmark.Bookmark{
					URL:         href,
					Title:       title,
					Description: "",
					Tags:        []string{}, // No tags from HTML import (for now)
				}

				if err := AddBookmark(db, newBookmark); err != nil {
					// Handle error, maybe log it and continue?
					fmt.Printf("Error adding bookmark %s: %v\n", href, err)
				}
			}
		}
	}
}

// ExportBookmarksToHTML exports bookmarks to an HTML file.
func ExportBookmarksToHTML(db *sql.DB, writer io.Writer) error {
	bookmarks, err := GetAllBookmarks(db)
	if err != nil {
		return fmt.Errorf("error retrieving bookmarks: %w", err)
	}

	// Write the HTML header
	_, err = writer.Write([]byte(`<!DOCTYPE html>
<html>
<head>
<title>Goku Bookmarks Export</title>
</head>
<body>
<h1>Goku Bookmarks</h1>
<ul>
`))
	if err != nil {
		return fmt.Errorf("error writing HTML header: %w", err)
	}

	// Write each bookmark as an <li> item
	for _, bm := range bookmarks {
		listItem := fmt.Sprintf(`<li><a href="%s">%s</a></li>`, bm.URL, bm.Title)
		_, err := writer.Write([]byte(listItem))
		if err != nil {
			return fmt.Errorf("error writing bookmark to HTML: %w", err)
		}
	}

	// Write the HTML footer
	_, err = writer.Write([]byte(`</ul>
</body>
</html>
`))
	if err != nil {
		return fmt.Errorf("error writing HTML footer: %w", err)
	}

	return nil
}
