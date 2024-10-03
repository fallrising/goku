package commands

import (
	"fmt"
	"github.com/fallrising/goku-cli/internal/bookmarks"
	"github.com/urfave/cli/v2"
)

func SyncCommand() *cli.Command {
	return &cli.Command{
		Name: "sync",
		Usage: "Sync data from SQLite to DuckDB for statistics\n\n" +
			"Example:\n" +
			"  goku sync",
		Action: func(c *cli.Context) error {
			fmt.Println("Syncing data to DuckDB...")
			bookmarkService := c.App.Metadata["bookmarkService"].(*bookmarks.BookmarkService)
			err := bookmarkService.SyncToDuckDB()
			if err != nil {
				return fmt.Errorf("failed to sync data to DuckDB: %w", err)
			}
			fmt.Println("Sync completed successfully")
			return nil
		},
	}
}
