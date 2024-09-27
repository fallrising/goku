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
		},
		Action: func(c *cli.Context) error {
			filePath := c.String("file")

			// Open the file
			file, err := openFile(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			// Determine import type based on file extension
			if isJSON(filePath) {
				return importFromJSON(bookmarkService, file)
			}
			return importFromHTML(bookmarkService, file)
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

// importFromJSON handles importing bookmarks from a JSON file.
func importFromJSON(bookmarkService *bookmarks.BookmarkService, file *os.File) error {
	fmt.Println("Importing from JSON")
	recordsCreated, err := bookmarkService.ImportFromJSON(context.Background(), file)
	if err != nil {
		return fmt.Errorf("failed to import JSON bookmarks: %w", err)
	}
	fmt.Printf("Import completed. %d bookmarks were successfully imported.\n", recordsCreated)
	return nil
}

// importFromHTML handles importing bookmarks from an HTML file.
func importFromHTML(bookmarkService *bookmarks.BookmarkService, file *os.File) error {
	fmt.Println("Importing from HTML")
	recordsCreated, err := bookmarkService.ImportFromHTML(context.Background(), file)
	if err != nil {
		return fmt.Errorf("failed to import HTML bookmarks: %w", err)
	}
	fmt.Printf("Import completed. %d bookmarks were successfully imported.\n", recordsCreated)
	return nil
}
