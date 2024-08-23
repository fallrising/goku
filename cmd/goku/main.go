package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"

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

		var wg sync.WaitGroup
		wg.Add(1)

		// Fetch title and description asynchronously
		go func() {
			defer wg.Done()
			fetchedTitle, fetchedDesc, err := fetchBookmarkData(url)
			if err != nil {
				fmt.Printf("Error fetching data for %s: %v\n", url, err)
			} else {
				if title == "" {
					title = fetchedTitle // Use fetched title if not provided
				}
				fmt.Printf("Fetched title: %s\n", fetchedTitle)
				fmt.Printf("Fetched description: %s\n", fetchedDesc)
			}
		}()

		tags, _ := cmd.Flags().GetString("tags")
		tagList := strings.Split(tags, ",")

		// Wait for the web request to complete
		wg.Wait()

		// Create Bookmark object
		newBookmark := &bookmark.Bookmark{
			URL:         url,
			Title:       title,
			Description: "", // Now fetched asynchronously
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

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search [KEYWORD]",
	Short: "Search for bookmarks",
	Args:  cobra.ExactArgs(1), // Require one keyword argument
	Run: func(cmd *cobra.Command, args []string) {
		keyword := args[0]
		bookmarks, err := database.SearchBookmarks(database.Db, keyword, "url", "title", "tags", "description")
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

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete [ID]",
	Short: "Delete a bookmark by ID",
	Args:  cobra.ExactArgs(1), // Require one argument (the ID)
	Run: func(cmd *cobra.Command, args []string) {
		// Parse the ID from the command arguments
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("Invalid ID format. Please provide a number.")
			return
		}

		// Delete the bookmark
		err = database.DeleteBookmark(database.Db, id)
		if err != nil {
			fmt.Println("Error deleting bookmark:", err)
			return
		}

		fmt.Println("Bookmark deleted successfully!")
	},
}

var updateCmd = &cobra.Command{
	Use:   "update [ID] [--url URL] [--title TITLE] [--description DESCRIPTION] [--tags TAGS] [--locked]",
	Short: "Update a bookmark by ID",
	Args:  cobra.ExactArgs(1), // Require one argument (the ID)
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("Invalid ID format. Please provide a number.")
			return
		}

		// Get the bookmark from the database
		bm, err := database.GetBookmarkByID(database.Db, id)
		if err != nil {
			fmt.Println("Error retrieving bookmark:", err)
			return
		}

		// Update bookmark fields based on provided flags
		url, _ := cmd.Flags().GetString("url")
		if url != "" {
			bm.URL = url
		}

		title, _ := cmd.Flags().GetString("title")
		if title != "" {
			bm.Title = title
		}

		description, _ := cmd.Flags().GetString("description")
		if description != "" {
			bm.Description = description
		}

		tags, _ := cmd.Flags().GetString("tags")
		if tags != "" {
			bm.Tags = strings.Split(tags, ",")
		}

		// Update the bookmark in the database
		if err := database.UpdateBookmark(database.Db, bm); err != nil {
			fmt.Println("Error updating bookmark:", err)
			return
		}

		fmt.Println("Bookmark updated successfully!")
	},
}

// browseCmd represents the browse command
var browseCmd = &cobra.Command{
	Use:   "browse [ID]",
	Short: "Open a bookmark in the browser",
	Args:  cobra.ExactArgs(1), // Require one argument (the ID)
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("Invalid ID format. Please provide a number.")
			return
		}

		// Retrieve the bookmark from the database
		bookmark, err := database.GetBookmarkByID(database.Db, id)
		if err != nil {
			fmt.Println("Error retrieving bookmark:", err)
			return
		}

		// Open the bookmark URL in the default browser
		openBrowser(bookmark.URL)
	},
}

// Function to open a URL in the default browser
func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		fmt.Println("Error opening browser:", err)
	}
}

func init() {
	addCmd.Flags().StringP("tags", "t", "", "Comma-separated tags")
	updateCmd.Flags().String("url", "", "New URL for the bookmark")
	updateCmd.Flags().String("title", "", "New title for the bookmark")
	updateCmd.Flags().String("description", "", "New description for the bookmark")
	updateCmd.Flags().String("tags", "", "Comma-separated tags for the bookmark")
}

// fetchTitle fetches the title of a webpage from the given URL.
func fetchBookmarkData(url string) (string, string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", "", fmt.Errorf("error making HTTP request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", "", fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Find the title element
	title := doc.Find("title").Text()
	title = strings.TrimSpace(title)

	// Find the description (using Open Graph meta tag for now)
	description := doc.Find("meta[property='og:description']").AttrOr("content", "")

	return title, description, nil
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "goku",
		Short: "A command-line bookmark manager.",
		Long:  `Goku is a fast and flexible tool for managing your bookmarks from the command line.`,
	}

	// Initialize the database
	var err error
	database.Db, err = database.InitDB() // Update to database.InitDB()
	if err != nil {
		fmt.Println("Error initializing database:", err)
		os.Exit(1)
	}
	defer database.Db.Close()

	// Add subcommands
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(deleteCmd) // Add the delete command
	rootCmd.AddCommand(browseCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
