package commands

import (
	"context"
	"fmt"
	"github.com/fallrising/goku-cli/internal/bookmarks"
	"github.com/urfave/cli/v2"
)

func GetCommand(bookmarkService *bookmarks.BookmarkService) *cli.Command {
	return &cli.Command{
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
	}
}
