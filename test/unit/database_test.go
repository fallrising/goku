package unit

import (
	"testing"

	"github.com/fallrising/goku-cli/internal/database"
)

func TestNewDatabase(t *testing.T) {
	db, err := database.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}

	err = db.Init()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Add more specific tests for database operations
}

// Add more tests for other database functions
