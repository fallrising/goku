// internal/bookmarks/bookmarks.go
package bookmarks

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "time"
    "goku/internal/db"
)

type Bookmark struct {
    ID          int       `json:"id"`
    URL         string    `json:"url"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Tags        string    `json:"tags"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type Tag struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

func AddTag(name string) error {
    query := `INSERT INTO tags (name) VALUES (?)`
    _, err := db.DB.Exec(query, name)
    if err != nil {
        return fmt.Errorf("failed to add tag: %v", err)
    }
    return nil
}

func RemoveTag(name string) error {
    query := `DELETE FROM tags WHERE name = ?`
    _, err := db.DB.Exec(query, name)
    if err != nil {
        return fmt.Errorf("failed to remove tag: %v", err)
    }
    return nil
}

func ListTags() ([]Tag, error) {
    query := `SELECT id, name FROM tags`
    rows, err := db.DB.Query(query)
    if err != nil {
        return nil, fmt.Errorf("failed to list tags: %v", err)
    }
    defer rows.Close()

    var tags []Tag
    for rows.Next() {
        var tag Tag
        err := rows.Scan(&tag.ID, &tag.Name)
        if err != nil {
            return nil, fmt.Errorf("failed to scan tag: %v", err)
        }
        tags = append(tags, tag)
    }
    return tags, nil
}

func SearchTags(keyword string) ([]Tag, error) {
    query := `SELECT id, name FROM tags WHERE name LIKE ?`
    rows, err := db.DB.Query(query, "%"+keyword+"%")
    if err != nil {
        return nil, fmt.Errorf("failed to search tags: %v", err)
    }
    defer rows.Close()

    var tags []Tag
    for rows.Next() {
        var tag Tag
        err := rows.Scan(&tag.ID, &tag.Name)
        if err != nil {
            return nil, fmt.Errorf("failed to scan tag: %v", err)
        }
        tags = append(tags, tag)
    }
    return tags, nil
}

func AddBookmark(url, title, description, tags string) error {
    query := `INSERT INTO bookmarks (url, title, description, tags) VALUES (?, ?, ?, ?)`
    _, err := db.DB.Exec(query, url, title, description, tags)
    if err != nil {
        return fmt.Errorf("failed to add bookmark: %v", err)
    }
    return nil
}

func UpdateBookmark(id int, url, title, description, tags string) error {
    query := `UPDATE bookmarks SET url = ?, title = ?, description = ?, tags = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
    _, err := db.DB.Exec(query, url, title, description, tags, id)
    if err != nil {
        return fmt.Errorf("failed to update bookmark: %v", err)
    }
    return nil
}

func DeleteBookmark(id int) error {
    query := `DELETE FROM bookmarks WHERE id = ?`
    _, err := db.DB.Exec(query, id)
    if err != nil {
        return fmt.Errorf("failed to delete bookmark: %v", err)
    }
    return nil
}

func SearchBookmarks(keyword string) ([]Bookmark, error) {
    query := `SELECT id, url, title, description, tags, created_at, updated_at FROM bookmarks WHERE url LIKE ? OR title LIKE ? OR description LIKE ? OR tags LIKE ?`
    rows, err := db.DB.Query(query, "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
    if err != nil {
        return nil, fmt.Errorf("failed to search bookmarks: %v", err)
    }
    defer rows.Close()

    var bookmarks []Bookmark
    for rows.Next() {
        var bookmark Bookmark
        err := rows.Scan(&bookmark.ID, &bookmark.URL, &bookmark.Title, &bookmark.Description, &bookmark.Tags, &bookmark.CreatedAt, &bookmark.UpdatedAt)
        if err != nil {
            return nil, fmt.Errorf("failed to scan bookmark: %v", err)
        }
        bookmarks = append(bookmarks, bookmark)
    }
    return bookmarks, nil
}

func ImportBookmarks(filename string) error {
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return fmt.Errorf("failed to read file: %v", err)
    }

    var bookmarks []Bookmark
    err = json.Unmarshal(data, &bookmarks)
    if err != nil {
        return fmt.Errorf("failed to unmarshal bookmarks: %v", err)
    }

    for _, bookmark := range bookmarks {
        err := AddBookmark(bookmark.URL, bookmark.Title, bookmark.Description, bookmark.Tags)
        if err != nil {
            return fmt.Errorf("failed to add bookmark: %v", err)
        }
    }

    return nil
}

func ExportBookmarks(filename string) error {
    bookmarks, err := ListBookmarks()
    if err != nil {
        return fmt.Errorf("failed to list bookmarks: %v", err)
    }

    data, err := json.MarshalIndent(bookmarks, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal bookmarks: %v", err)
    }

    err = ioutil.WriteFile(filename, data, 0644)
    if err != nil {
        return fmt.Errorf("failed to write file: %v", err)
    }

    return nil
}

func ListBookmarks() ([]Bookmark, error) {
    query := `SELECT id, url, title, description, tags, created_at, updated_at FROM bookmarks`
    rows, err := db.DB.Query(query)
    if err != nil {
        return nil, fmt.Errorf("failed to list bookmarks: %v", err)
    }
    defer rows.Close()

    var bookmarks []Bookmark
    for rows.Next() {
        var bookmark Bookmark
        err := rows.Scan(&bookmark.ID, &bookmark.URL, &bookmark.Title, &bookmark.Description, &bookmark.Tags, &bookmark.CreatedAt, &bookmark.UpdatedAt)
        if err != nil {
            return nil, fmt.Errorf("failed to scan bookmark: %v", err)
        }
        bookmarks = append(bookmarks, bookmark)
    }
    return bookmarks, nil
}