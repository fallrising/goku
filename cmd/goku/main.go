package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/fallrising/goku-cli/internal/bookmarks"
	"github.com/fallrising/goku-cli/internal/database"
	"github.com/fallrising/goku-cli/pkg/models"
	"github.com/urfave/cli/v2"
)

func main() {
	dbPath := os.Getenv("GOKU_DB_PATH")
	if dbPath == "" {
		dbPath = "goku.db" // Default to current directory if not specified
	}

	db, err := database.NewDatabase(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	if err := db.Init(); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	bookmarkService := bookmarks.NewBookmarkService(db)

	app := &cli.App{
		Name:  "goku",
		Usage: "A powerful CLI bookmark manager",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "db",
				EnvVars: []string{"GOKU_DB_PATH"},
				Value:   "goku.db",
				Usage:   "Path to the Goku database file",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add a new bookmark",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "url", Required: true},
					&cli.StringFlag{Name: "title"},
					&cli.StringFlag{Name: "description"},
					&cli.StringSliceFlag{Name: "tags"},
				},
				Action: func(c *cli.Context) error {
					bookmark := &models.Bookmark{
						URL:         c.String("url"),
						Title:       c.String("title"),
						Description: c.String("description"),
						Tags:        c.StringSlice("tags"),
					}
					err := bookmarkService.CreateBookmark(context.Background(), bookmark)
					if err != nil {
						return fmt.Errorf("failed to add bookmark: %w", err)
					}
					fmt.Printf("Bookmark added successfully with ID: %d\n", bookmark.ID)
					return nil
				},
			},
			{
				Name:  "get",
				Usage: "Get a bookmark by ID",
				Flags: []cli.Flag{
					&cli.Int64Flag{Name: "id", Required: true},
				},
				Action: func(c *cli.Context) error {
					bookmark, err := bookmarkService.GetBookmark(context.Background(), c.Int64("id"))
					if err != nil {
						return fmt.Errorf("failed to get bookmark: %w", err)
					}
					fmt.Printf("Bookmark: %+v\n", bookmark)
					return nil
				},
			},
			{
				Name:  "update",
				Usage: "Update a bookmark",
				Flags: []cli.Flag{
					&cli.Int64Flag{Name: "id", Required: true},
					&cli.StringFlag{Name: "url"},
					&cli.StringFlag{Name: "title"},
					&cli.StringFlag{Name: "description"},
					&cli.StringSliceFlag{Name: "tags"},
				},
				Action: func(c *cli.Context) error {
					bookmark := &models.Bookmark{
						ID:          c.Int64("id"),
						URL:         c.String("url"),
						Title:       c.String("title"),
						Description: c.String("description"),
						Tags:        c.StringSlice("tags"),
					}
					err := bookmarkService.UpdateBookmark(context.Background(), bookmark)
					if err != nil {
						return fmt.Errorf("failed to update bookmark: %w", err)
					}
					fmt.Println("Bookmark updated successfully")
					return nil
				},
			},
			{
				Name:  "delete",
				Usage: "Delete a bookmark",
				Flags: []cli.Flag{
					&cli.Int64Flag{Name: "id", Required: true},
				},
				Action: func(c *cli.Context) error {
					err := bookmarkService.DeleteBookmark(context.Background(), c.Int64("id"))
					if err != nil {
						return fmt.Errorf("failed to delete bookmark: %w", err)
					}
					fmt.Println("Bookmark deleted successfully")
					return nil
				},
			},
			{
				Name:  "list",
				Usage: "List all bookmarks",
				Action: func(c *cli.Context) error {
					bookmarks, err := bookmarkService.ListBookmarks(context.Background())
					if err != nil {
						return fmt.Errorf("failed to list bookmarks: %w", err)
					}
					for _, b := range bookmarks {
						fmt.Printf("ID: %d, URL: %s, Title: %s, Tags: %v, Description: %v\n", b.ID, b.URL, b.Title, b.Tags, b.Description)
					}
					return nil
				},
			},
			{
				Name:  "search",
				Usage: "Search bookmarks",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "query", Aliases: []string{"q"}, Required: true},
				},
				Action: func(c *cli.Context) error {
					query := c.String("query")
					bookmarks, err := bookmarkService.SearchBookmarks(context.Background(), query)
					if err != nil {
						return fmt.Errorf("failed to search bookmarks: %w", err)
					}
					if len(bookmarks) == 0 {
						fmt.Println("No bookmarks found matching the query.")
						return nil
					}
					fmt.Printf("Found %d bookmark(s):\n", len(bookmarks))
					for _, b := range bookmarks {
						fmt.Printf("ID: %d, URL: %s, Title: %s, Tags: %v, Description: %v\n", b.ID, b.URL, b.Title, b.Tags, b.Description)
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
