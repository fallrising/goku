package unit

import (
	"testing"

	"github.com/fallrising/goku-cli/internal/fetcher"
)

func TestFetchPageContent(t *testing.T) {
	// Note: This test requires internet connection
	url := "https://example.com"
	content, err := fetcher.FetchPageContent(url)
	if err != nil {
		t.Fatalf("Failed to fetch page content: %v", err)
	}

	if content.Title == "" {
		t.Error("Page title is empty")
	}

	if content.Description == "" {
		t.Error("Page description is empty")
	}
}

// Add more tests for other fetcher functions
