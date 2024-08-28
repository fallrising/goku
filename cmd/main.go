// cmd/main.go
package main

import (
    "fmt"
    "log"
    "os"
    "strconv"

    "goku/internal/db"
    "goku/internal/bookmarks"
    "github.com/spf13/cobra"
)

func main() {
    db.InitDB("bookmarks.db")

    var rootCmd = &cobra.Command{Use: "goku"}

    var addCmd = &cobra.Command{
        Use:   "add [url] [title] [description] [tags]",
        Short: "Add a new bookmark",
        Args:  cobra.ExactArgs(4),
        Run: func(cmd *cobra.Command, args []string) {
            url, title, description, tags := args[0], args[1], args[2], args[3]
            err := bookmarks.AddBookmark(url, title, description, tags)
            if err != nil {
                log.Fatalf("Error adding bookmark: %v", err)
            }
            fmt.Println("Bookmark added successfully!")
        },
    }

    var updateCmd = &cobra.Command{
        Use:   "update [id] [url] [title] [description] [tags]",
        Short: "Update an existing bookmark",
        Args:  cobra.ExactArgs(5),
        Run: func(cmd *cobra.Command, args []string) {
            id, err := strconv.Atoi(args[0])
            if err != nil {
                log.Fatalf("Invalid ID: %v", err)
            }
            url, title, description, tags := args[1], args[2], args[3], args[4]
            err = bookmarks.UpdateBookmark(id, url, title, description, tags)
            if err != nil {
                log.Fatalf("Error updating bookmark: %v", err)
            }
            fmt.Println("Bookmark updated successfully!")
        },
    }

    var deleteCmd = &cobra.Command{
        Use:   "delete [id]",
        Short: "Delete a bookmark",
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            id, err := strconv.Atoi(args[0])
            if err != nil {
                log.Fatalf("Invalid ID: %v", err)
            }
            err = bookmarks.DeleteBookmark(id)
            if err != nil {
                log.Fatalf("Error deleting bookmark: %v", err)
            }
            fmt.Println("Bookmark deleted successfully!")
        },
    }

    var searchCmd = &cobra.Command{
        Use:   "search [keyword]",
        Short: "Search bookmarks",
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            keyword := args[0]
            results, err := bookmarks.SearchBookmarks(keyword)
            if err != nil {
                log.Fatalf("Error searching bookmarks: %v", err)
            }
            for _, bookmark := range results {
                fmt.Printf("ID: %d, URL: %s, Title: %s, Description: %s, Tags: %s\n", bookmark.ID, bookmark.URL, bookmark.Title, bookmark.Description, bookmark.Tags)
            }
        },
    }

    var tagCmd = &cobra.Command{
        Use:   "tag",
        Short: "Manage tags",
    }

    var addTagCmd = &cobra.Command{
        Use:   "add [name]",
        Short: "Add a new tag",
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            name := args[0]
            err := bookmarks.AddTag(name)
            if err != nil {
                log.Fatalf("Error adding tag: %v", err)
            }
            fmt.Println("Tag added successfully!")
        },
    }

    var removeTagCmd = &cobra.Command{
        Use:   "remove [name]",
        Short: "Remove a tag",
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            name := args[0]
            err := bookmarks.RemoveTag(name)
            if err != nil {
                log.Fatalf("Error removing tag: %v", err)
            }
            fmt.Println("Tag removed successfully!")
        },
    }

    var listTagsCmd = &cobra.Command{
        Use:   "list",
        Short: "List all tags",
        Run: func(cmd *cobra.Command, args []string) {
            tags, err := bookmarks.ListTags()
            if err != nil {
                log.Fatalf("Error listing tags: %v", err)
            }
            for _, tag := range tags {
                fmt.Printf("ID: %d, Name: %s\n", tag.ID, tag.Name)
            }
        },
    }

    var searchTagsCmd = &cobra.Command{
        Use:   "search [keyword]",
        Short: "Search tags",
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            keyword := args[0]
            tags, err := bookmarks.SearchTags(keyword)
            if err != nil {
                log.Fatalf("Error searching tags: %v", err)
            }
            for _, tag := range tags {
                fmt.Printf("ID: %d, Name: %s\n", tag.ID, tag.Name)
            }
        },
    }

    var importCmd = &cobra.Command{
        Use:   "import [filename]",
        Short: "Import bookmarks from a file",
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            filename := args[0]
            err := bookmarks.ImportBookmarks(filename)
            if err != nil {
                log.Fatalf("Error importing bookmarks: %v", err)
            }
            fmt.Println("Bookmarks imported successfully!")
        },
    }

    var exportCmd = &cobra.Command{
        Use:   "export [filename]",
        Short: "Export bookmarks to a file",
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            filename := args[0]
            err := bookmarks.ExportBookmarks(filename)
            if err != nil {
                log.Fatalf("Error exporting bookmarks: %v", err)
            }
            fmt.Println("Bookmarks exported successfully!")
        },
    }

    tagCmd.AddCommand(addTagCmd, removeTagCmd, listTagsCmd, searchTagsCmd)
    rootCmd.AddCommand(addCmd, updateCmd, deleteCmd, searchCmd, tagCmd, importCmd, exportCmd)
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}