package main

import (
	"fmt"
	"strings"

	"github.com/fallrising/goku/internal/database" // Import the database package
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all bookmarks",
	Run: func(cmd *cobra.Command, args []string) {
		bookmarks, err := database.GetAllBookmarks(database.Db)
		if err != nil {
			fmt.Println("Error retrieving bookmarks:", err)
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
