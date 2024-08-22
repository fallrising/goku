package main

import (
	"fmt"
	"strings"

	"github.com/fallrising/goku/internal/bookmark"
	"github.com/fallrising/goku/internal/database"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add [URL] [TITLE] [--tags TAGS]",
	Short: "Add a new bookmark",
	Args:  cobra.RangeArgs(1, 2), // Require at least URL, allow optional title
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		title := ""
		if len(args) == 2 {
			title = args[1]
		}
		tags, _ := cmd.Flags().GetString("tags")
		tagList := strings.Split(tags, ",")

		// 1. Fetch title from web if not provided (use goquery or similar)

		// 2. Create Bookmark object
		newBookmark := &bookmark.Bookmark{
			URL:         url,
			Title:       title,
			Description: "", // Fetch description later (optional)
			Tags:        tagList,
		}

		// 3. Add bookmark to database
		if err := database.AddBookmark(database.Db, newBookmark); err != nil {
			fmt.Println("Error adding bookmark:", err)
		} else {
			fmt.Println("Bookmark added successfully!")
		}
	},
}

func init() {
	addCmd.Flags().StringP("tags", "t", "", "Comma-separated tags")
}
