package commands

import (
	"context"
	"fmt"
	"github.com/fallrising/goku-cli/internal/bookmarks"
	"github.com/urfave/cli/v2"
)

func DeleteCommand(bookmarkService *bookmarks.BookmarkService) *cli.Command {
	return &cli.Command{
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
	}
}
