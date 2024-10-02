package commands

import (
	"context"
	"fmt"
	"github.com/fallrising/goku-cli/internal/bookmarks"
	"github.com/urfave/cli/v2"
	"os"
	"strings"
)

func ImportCommand(bookmarkService *bookmarks.BookmarkService) *cli.Command {
	return &cli.Command{
		Name:  "import",
		Usage: "Import bookmarks from either HTML or JSON format",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "file",
				Aliases:  []string{"f"},
				Usage:    "Input file path (.html or .json)",
				Required: true,
			},
			&cli.IntFlag{
				Name:    "workers",
				Aliases: []string{"w"},
				Usage:   "Number of worker goroutines for concurrent processing",
				Value:   5, // Default value
			},
			&cli.BoolFlag{
				Name:    "fetch",
				Aliases: []string{"F"},
				Usage:   "Enable fetching additional data for each bookmark",
				Value:   false, // Disabled by default
			},
		},
		Action: func(c *cli.Context) error {
			filePath := c.String("file")
			numWorkers := c.Int("workers")
			fetchData := c.Bool("fetch")

			// Open the file
			file, err := openFile(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			// Create a context with the import options
			ctx := context.WithValue(context.Background(), "numWorkers", numWorkers)
			ctx = context.WithValue(ctx, "fetchData", fetchData)

			// Determine import type based on file extension
			var recordsCreated int
			if isJSON(filePath) {
				recordsCreated, err = bookmarkService.ImportFromJSON(ctx, file)
			} else {
				recordsCreated, err = bookmarkService.ImportFromHTML(ctx, file)
			}

			if err != nil {
				return fmt.Errorf("failed to import bookmarks: %w", err)
			}

			fmt.Printf("Import completed. %d bookmarks were successfully imported.\n", recordsCreated)
			if fetchData {
				fmt.Println("Additional data was fetched for each bookmark.")
			}
			return nil
		},
	}
}

// openFile opens the file and returns an error if it fails.
func openFile(filePath string) (*os.File, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}

// isJSON checks if the file is a JSON file based on the file extension.
func isJSON(filePath string) bool {
	return strings.HasSuffix(filePath, ".json")
}
