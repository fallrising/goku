package commands

import (
	"context"
	"fmt"
	"github.com/fallrising/goku-cli/internal/bookmarks"
	"github.com/urfave/cli/v2"
)

func ListCommand(bookmarkService *bookmarks.BookmarkService) *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all bookmarks with pagination",
		Flags: []cli.Flag{
			&cli.IntFlag{Name: "limit", Value: 10, Usage: "Number of bookmarks to display per page"},
			&cli.IntFlag{Name: "offset", Value: 0, Usage: "Offset to start listing bookmarks from"},
		},
		Action: func(c *cli.Context) error {
			limit := c.Int("limit")
			offset := c.Int("offset")

			bookmarks, err := bookmarkService.ListBookmarks(context.Background(), limit, offset)
			if err != nil {
				return fmt.Errorf("failed to list bookmarks: %w", err)
			}
			if len(bookmarks) == 0 {
				fmt.Println("No bookmarks found.")
				return nil
			}
			fmt.Printf("Displaying %d bookmark(s):\n", len(bookmarks))
			for _, b := range bookmarks {
				fmt.Printf("ID: %d, URL: %s, Title: %s, Tags: %v, Description: %v\n", b.ID, b.URL, b.Title, b.Tags, b.Description)
			}
			return nil
		},
	}
}
