package commands

import (
	"context"
	"fmt"
	"github.com/fallrising/goku-cli/internal/bookmarks"
	"github.com/fallrising/goku-cli/pkg/models"
	"github.com/urfave/cli/v2"
)

func AddCommand(bookmarkService *bookmarks.BookmarkService) *cli.Command {
	return &cli.Command{
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
	}
}
