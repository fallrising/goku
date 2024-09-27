package integration

import (
	"context"
	"os"
	"testing"

	"github.com/fallrising/goku-cli/internal/database"
	"github.com/fallrising/goku-cli/pkg/models"
)

func TestDatabaseOperations(t *testing.T) {
	dbPath := "test_integration.db"
	defer os.Remove(dbPath)

	db, err := database.NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	err = db.Init()
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	ctx := context.Background()

	// Test Create
	bookmark := &models.Bookmark{
		URL:   "https://example.com",
		Title: "Example",
	}
	err = db.Create(ctx, bookmark)
	if err != nil {
		t.Fatalf("Failed to create bookmark: %v", err)
	}

	// Test GetByID
	retrieved, err := db.GetByID(ctx, bookmark.ID)
	if err != nil {
		t.Fatalf("Failed to get bookmark by ID: %v", err)
	}
	if retrieved.URL != bookmark.URL {
		t.Errorf("Retrieved bookmark URL mismatch. Got %s, want %s", retrieved.URL, bookmark.URL)
	}

	// Test Update
	bookmark.Title = "Updated Example"
	err = db.Update(ctx, bookmark)
	if err != nil {
		t.Fatalf("Failed to update bookmark: %v", err)
	}

	// Verify update
	updated, err := db.GetByID(ctx, bookmark.ID)
	if err != nil {
		t.Fatalf("Failed to get updated bookmark: %v", err)
	}
	if updated.Title != "Updated Example" {
		t.Errorf("Updated bookmark title mismatch. Got %s, want %s", updated.Title, "Updated Example")
	}

	// Test Delete
	err = db.Delete(ctx, bookmark.ID)
	if err != nil {
		t.Fatalf("Failed to delete bookmark: %v", err)
	}

	// Verify deletion
	deleted, err := db.GetByID(ctx, bookmark.ID)
	if err == nil || deleted != nil {
		t.Error("Bookmark should have been deleted")
	}
}

// Add more integration tests for other database operations
