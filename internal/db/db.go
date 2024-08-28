// internal/db/db.go
package db

import (
    "database/sql"
    "log"

    _ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB(filepath string) {
    var err error
    DB, err = sql.Open("sqlite3", filepath)
    if err != nil {
        log.Fatalf("Failed to open database: %v", err)
    }

    createTable()
}

func createTable() {
    createBookmarksTableSQL := `CREATE TABLE IF NOT EXISTS bookmarks (
        "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        "url" TEXT NOT NULL UNIQUE,
        "title" TEXT,
        "description" TEXT,
        "tags" TEXT,
        "created_at" DATETIME DEFAULT CURRENT_TIMESTAMP,
        "updated_at" DATETIME DEFAULT CURRENT_TIMESTAMP
    );`

    createTagsTableSQL := `CREATE TABLE IF NOT EXISTS tags (
        "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        "name" TEXT NOT NULL UNIQUE
    );`

    createBookmarkTagsTableSQL := `CREATE TABLE IF NOT EXISTS bookmark_tags (
        "bookmark_id" INTEGER NOT NULL,
        "tag_id" INTEGER NOT NULL,
        FOREIGN KEY (bookmark_id) REFERENCES bookmarks(id),
        FOREIGN KEY (tag_id) REFERENCES tags(id)
    );`

    _, err := DB.Exec(createBookmarksTableSQL)
    if err != nil {
        log.Fatalf("Failed to create bookmarks table: %v", err)
    }

    _, err = DB.Exec(createTagsTableSQL)
    if err != nil {
        log.Fatalf("Failed to create tags table: %v", err)
    }

    _, err = DB.Exec(createBookmarkTagsTableSQL)
    if err != nil {
        log.Fatalf("Failed to create bookmark_tags table: %v", err)
    }
}