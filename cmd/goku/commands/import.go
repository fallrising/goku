package commands

import (
	"context"
	"fmt"
	"github.com/fallrising/goku-cli/internal/bookmarks"
	"github.com/urfave/cli/v2"
	"os"
)

func ImportCommand(bookmarkService *bookmarks.BookmarkService) *cli.Command {
	return &cli.Command{
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
	}
}
