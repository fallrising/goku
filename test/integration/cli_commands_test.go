package integration

import (
	"os"
	"testing"

	"github.com/fallrising/goku-cli/cmd/goku/commands"
	"github.com/fallrising/goku-cli/internal/bookmarks"
	"github.com/fallrising/goku-cli/internal/database"
	"github.com/urfave/cli/v2"
)

func TestAddCommand(t *testing.T) {
	dbPath := "test.db"
	defer os.Remove(dbPath)

	db, err := database.NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	err = db.Init()
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	bookmarkService := bookmarks.NewBookmarkService(db)
	app := &cli.App{
		Commands: []*cli.Command{
			commands.AddCommand(bookmarkService),
		},
	}

	err = app.Run([]string{"goku", "add", "--url", "https://example.com", "--title", "Example"})
	if err != nil {
		t.Fatalf("Failed to run add command: %v", err)
	}

	// Verify the bookmark was added
	bookmarks, err := bookmarkService.ListBookmarks(nil, 1, 0)
	if err != nil {
		t.Fatalf("Failed to list bookmarks: %v", err)
	}

	if len(bookmarks) != 1 {
		t.Errorf("Expected 1 bookmark, got %d", len(bookmarks))
	}

	if bookmarks[0].URL != "https://example.com" {
		t.Errorf("Expected URL 'https://example.com', got '%s'", bookmarks[0].URL)
	}
}

// Add more tests for other CLI commands
