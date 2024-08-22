package main

import (
	"fmt"
	"strings"

	"github.com/fallrising/goku/internal/database" // Import the database package
	"github.com/spf13/cobra"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search [KEYWORD]",
	Short: "Search for bookmarks",
	Args:  cobra.ExactArgs(1), // Require one keyword argument
	Run: func(cmd *cobra.Command, args []string) {
		keyword := args[0]
		bookmarks, err := database.SearchBookmarks(database.Db, keyword)
		if err != nil {
			fmt.Println("Error searching bookmarks:", err)
			return
		}

		for _, bm := range bookmarks {
			fmt.Printf("ID: %d\n", bm.ID)
			fmt.Printf("URL: %s\n", bm.URL)
			fmt.Printf("Title: %s\n", bm.Title)
			fmt.Printf("Description: %s\n", bm.Description)
			fmt.Printf("Tags: %s\n", strings.Join(bm.Tags, ","))
			fmt.Println("--------------------")
		}
	},
}
