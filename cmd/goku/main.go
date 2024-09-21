package main

import (
	"context"
	"fmt"
	"github.com/fallrising/goku-cli/internal/bookmarks"
	"github.com/fallrising/goku-cli/internal/database"
	"github.com/fallrising/goku-cli/pkg/models"
	"github.com/urfave/cli/v2"
	"log"
	"os"
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
				Name:  "purge",
				Usage: "Purge all bookmark data",
				Action: func(c *cli.Context) error {
					err := db.PurgeAllData(context.Background())
					if err != nil {
						return fmt.Errorf("failed to purge all data: %w", err)
					}
					fmt.Println("All bookmark data has been purged.")
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
				Usage: "List all bookmarks with pagination",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "limit", Value: 10, Usage: "Number of bookmarks to display per page"},
					&cli.IntFlag{Name: "offset", Value: 0, Usage: "Offset to start listing bookmarks from"},
				},
				Action: func(c *cli.Context) error {
					limit := c.Int("limit")
					offset := c.Int("offset")

					listBookmarks, err := bookmarkService.ListBookmarks(context.Background(), limit, offset)
					if err != nil {
						return fmt.Errorf("failed to list listBookmarks: %w", err)
					}
					if len(listBookmarks) == 0 {
						fmt.Println("No listBookmarks found.")
						return nil
					}
					fmt.Printf("Displaying %d bookmark(s):\n", len(listBookmarks))
					for _, b := range listBookmarks {
						fmt.Printf("ID: %d, URL: %s, Title: %s, Tags: %v, Description: %v\n", b.ID, b.URL, b.Title, b.Tags, b.Description)
					}
					return nil
				},
			},
			{
				Name:  "search",
				Usage: "Search bookmarks with pagination",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "query", Aliases: []string{"q"}, Required: true, Usage: "Search query"},
					&cli.IntFlag{Name: "limit", Value: 10, Usage: "Number of bookmarks to display per page"},
					&cli.IntFlag{Name: "offset", Value: 0, Usage: "Offset to start search results from"},
				},
				Action: func(c *cli.Context) error {
					query := c.String("query")
					limit := c.Int("limit")
					offset := c.Int("offset")

					searchBookmarks, err := bookmarkService.SearchBookmarks(context.Background(), query, limit, offset)
					if err != nil {
						return fmt.Errorf("failed to search searchBookmarks: %w", err)
					}
					if len(searchBookmarks) == 0 {
						fmt.Println("No searchBookmarks found matching the query.")
						return nil
					}
					fmt.Printf("Found %d bookmark(s):\n", len(searchBookmarks))
					for _, b := range searchBookmarks {
						fmt.Printf("ID: %d, URL: %s, Title: %s, Tags: %v, Description: %v\n", b.ID, b.URL, b.Title, b.Tags, b.Description)
					}
					return nil
				},
			},
			{
				Name:  "tags",
				Usage: "Manage tags for bookmarks",
				Subcommands: []*cli.Command{
					{
						Name:  "remove",
						Usage: "Remove a tag from a bookmark",
						Flags: []cli.Flag{
							&cli.Int64Flag{Name: "id", Required: true, Usage: "Bookmark ID"},
							&cli.StringFlag{Name: "tag", Required: true, Usage: "Tag to remove"},
						},
						Action: func(c *cli.Context) error {
							bookmarkID := c.Int64("id")
							tag := c.String("tag")
							err := bookmarkService.RemoveTagFromBookmark(context.Background(), bookmarkID, tag)
							if err != nil {
								return fmt.Errorf("failed to remove tag: %w", err)
							}
							fmt.Println("Tag removed successfully")
							return nil
						},
					},
					{
						Name:  "list",
						Usage: "List all unique tags",
						Action: func(c *cli.Context) error {
							tags, err := bookmarkService.ListAllTags(context.Background())
							if err != nil {
								return fmt.Errorf("failed to list tags: %w", err)
							}
							if len(tags) == 0 {
								fmt.Println("No tags found.")
								return nil
							}
							fmt.Println("Tags:")
							for _, tag := range tags {
								fmt.Println(" -", tag)
							}
							return nil
						},
					},
				},
			},
			{
				Name:  "import",
				Usage: "Import bookmarks from HTML format",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "file",
						Aliases:  []string{"f"},
						Usage:    "Input HTML file path",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					filePath := c.String("file")
					file, err := os.Open(filePath)
					if err != nil {
						return fmt.Errorf("failed to open file: %w", err)
					}
					defer file.Close()

					recordsCreated, err := bookmarkService.ImportFromHTML(context.Background(), file)
					if err != nil {
						return fmt.Errorf("failed to import bookmarks: %w", err)
					}

					fmt.Printf("Import completed. %d bookmarks were successfully imported.\n", recordsCreated)
					return nil
				},
			},
			{
				Name:  "export",
				Usage: "Export bookmarks to HTML format",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "Output file path (default: stdout)",
					},
				},
				Action: func(c *cli.Context) error {
					fmt.Println("Exporting bookmarks...")
					html, err := bookmarkService.ExportToHTML(context.Background())
					if err != nil {
						return fmt.Errorf("failed to export bookmarks: %w", err)
					}

					outputPath := c.String("output")
					if outputPath == "" {
						// Write to stdout if no output file specified
						fmt.Println(html)
					} else {
						// Write to file
						err = os.WriteFile(outputPath, []byte(html), 0644)
						if err != nil {
							return fmt.Errorf("failed to write to file: %w", err)
						}
						fmt.Printf("Bookmarks exported to %s\n", outputPath)
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
