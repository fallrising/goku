// cmd/main.go
package main

import (
    "fmt"
    "log"
    "os"
    "strconv"

    "goku-cli/internal/bookmarks"
    "goku-cli/internal/db"

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

    rootCmd.AddCommand(addCmd, updateCmd, deleteCmd, searchCmd)
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}