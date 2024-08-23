package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
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

		// Fetch title from web if title is not provided
		if title == "" {
			fetchedTitle, err := fetchTitle(url)
			if err != nil {
				fmt.Printf("Error fetching title for %s: %v\n", url, err)
			} else {
				title = fetchedTitle
				fmt.Printf("Fetched title: %s\n", title)
			}
		}

		tags, _ := cmd.Flags().GetString("tags")
		tagList := strings.Split(tags, ",")

		// Create Bookmark object
		newBookmark := &bookmark.Bookmark{
			URL:         url,
			Title:       title,
			Description: "", // Fetch description later (optional)
			Tags:        tagList,
		}

		// Add bookmark to database
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

// fetchTitle fetches the title of a webpage from the given URL.
func fetchTitle(url string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error making HTTP request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Find the title element
	title := doc.Find("title").Text()
	return strings.TrimSpace(title), nil
}
