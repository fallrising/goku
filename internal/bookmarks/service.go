package bookmarks

import (
	"context"
	"fmt"
	"github.com/fallrising/goku-cli/internal/fetcher"
	"strings"

	"github.com/fallrising/goku-cli/pkg/interfaces"
	"github.com/fallrising/goku-cli/pkg/models"
)

type BookmarkService struct {
	repo interfaces.BookmarkRepository
}

func NewBookmarkService(repo interfaces.BookmarkRepository) *BookmarkService {
	return &BookmarkService{repo: repo}
}

func (s *BookmarkService) CreateBookmark(ctx context.Context, bookmark *models.Bookmark) error {
	if bookmark.URL == "" {
		return fmt.Errorf("URL is required")
	}

	// Check if URL already exists in the database
	existingBookmark, err := s.repo.GetByURL(ctx, bookmark.URL)
	if err != nil {
		return fmt.Errorf("failed to check for existing bookmark: %w", err)
	}
	if existingBookmark != nil {
		return fmt.Errorf("bookmark with this URL already exists: %s", existingBookmark.URL)
	}

	// Check if URL starts with "http://" or "https://"
	if !(strings.HasPrefix(bookmark.URL, "http://") || strings.HasPrefix(bookmark.URL, "https://")) {
		bookmark.URL = "https://" + bookmark.URL
	}

	// Fetch page content if title, description, or tags are not provided
	if bookmark.Title == "" || bookmark.Description == "" || len(bookmark.Tags) == 0 {
		content, err := fetcher.FetchPageContent(bookmark.URL)
		if err != nil {
			fmt.Printf("Warning: Failed to fetch page content: %v\n", err)
			return err
		} else {
			if bookmark.Title == "" {
				bookmark.Title = content.Title
			}
			if bookmark.Description == "" {
				bookmark.Description = content.Description
			}
			if len(bookmark.Tags) == 0 {
				bookmark.Tags = content.Tags
			}
		}
	}

	return s.repo.Create(ctx, bookmark)
}

func (s *BookmarkService) GetBookmark(ctx context.Context, id int64) (*models.Bookmark, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *BookmarkService) UpdateBookmark(ctx context.Context, updatedBookmark *models.Bookmark) error {
	if updatedBookmark.ID == 0 {
		return fmt.Errorf("bookmark ID is required")
	}

	// Fetch existing bookmark
	existingBookmark, err := s.repo.GetByID(ctx, updatedBookmark.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch existing bookmark: %w", err)
	}
	if existingBookmark == nil {
		return fmt.Errorf("bookmark not found with ID: %d", updatedBookmark.ID)
	}

	// Check if the URL has changed
	if updatedBookmark.URL != "" && updatedBookmark.URL != existingBookmark.URL {
		// Check for duplicates
		duplicate, err := s.repo.GetByURL(ctx, updatedBookmark.URL)
		if err != nil {
			return fmt.Errorf("failed to check for duplicate URL: %w", err)
		}
		if duplicate != nil {
			return fmt.Errorf("another bookmark with URL '%s' already exists", updatedBookmark.URL)
		}

		// Fetch new metadata for the new URL
		content, err := fetcher.FetchPageContent(updatedBookmark.URL)
		if err != nil {
			return fmt.Errorf("failed to fetch metadata for the updated URL: %w", err)
		}

		// Update the metadata with fetched content
		updatedBookmark.Title = content.Title
		updatedBookmark.Description = content.Description
		updatedBookmark.Tags = content.Tags
	}

	// Track changes
	updated := false

	if updatedBookmark.URL != "" && updatedBookmark.URL != existingBookmark.URL {
		existingBookmark.URL = updatedBookmark.URL
		updated = true
	}
	if updatedBookmark.Title != "" && updatedBookmark.Title != existingBookmark.Title {
		existingBookmark.Title = updatedBookmark.Title
		updated = true
	}
	if updatedBookmark.Description != "" && updatedBookmark.Description != existingBookmark.Description {
		existingBookmark.Description = updatedBookmark.Description
		updated = true
	}
	if len(updatedBookmark.Tags) > 0 && !equalTags(updatedBookmark.Tags, existingBookmark.Tags) {
		existingBookmark.Tags = updatedBookmark.Tags
		updated = true
	}

	// Update only if necessary
	if updated {
		return s.repo.Update(ctx, existingBookmark)
	}

	fmt.Println("No changes detected, bookmark update skipped.")
	return nil
}

// Helper function to check if tags are equal
func equalTags(tags1, tags2 []string) bool {
	if len(tags1) != len(tags2) {
		return false
	}
	for i, tag := range tags1 {
		if tag != tags2[i] {
			return false
		}
	}
	return true
}

func (s *BookmarkService) DeleteBookmark(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *BookmarkService) ListBookmarks(ctx context.Context, limit, offset int) ([]*models.Bookmark, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *BookmarkService) SearchBookmarks(ctx context.Context, query string, limit, offset int) ([]*models.Bookmark, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	bookmarks, err := s.repo.Search(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search bookmarks: %w", err)
	}

	if len(bookmarks) == 0 {
		fmt.Println("No bookmarks found matching the query.")
	}

	return bookmarks, nil
}

func (s *BookmarkService) RemoveTagFromBookmark(ctx context.Context, bookmarkID int64, tagToRemove string) error {
	// Fetch the existing bookmark
	bookmark, err := s.repo.GetByID(ctx, bookmarkID)
	if err != nil {
		return fmt.Errorf("failed to fetch bookmark: %w", err)
	}

	// Remove the specified tag
	bookmark.RemoveTag(tagToRemove)

	// Save the updated bookmark
	err = s.repo.Update(ctx, bookmark)
	if err != nil {
		return fmt.Errorf("failed to update bookmark after removing tag: %w", err)
	}

	return nil
}

func (s *BookmarkService) ListAllTags(ctx context.Context) ([]string, error) {
	tags, err := s.repo.ListAllTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}
	return tags, nil
}
