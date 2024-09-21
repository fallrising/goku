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

	// Fetch the existing bookmark
	existingBookmark, err := s.repo.GetByID(ctx, updatedBookmark.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch existing bookmark: %w", err)
	}

	// Update only the fields that are provided
	if updatedBookmark.URL != "" {
		existingBookmark.URL = updatedBookmark.URL
	}
	if updatedBookmark.Title != "" {
		existingBookmark.Title = updatedBookmark.Title
	}
	if updatedBookmark.Description != "" {
		existingBookmark.Description = updatedBookmark.Description
	}
	if len(updatedBookmark.Tags) > 0 {
		existingBookmark.Tags = updatedBookmark.Tags
	}

	// Save the updated bookmark
	return s.repo.Update(ctx, existingBookmark)
}

func (s *BookmarkService) DeleteBookmark(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *BookmarkService) ListBookmarks(ctx context.Context) ([]*models.Bookmark, error) {
	return s.repo.List(ctx)
}

func (s *BookmarkService) SearchBookmarks(ctx context.Context, query string) ([]*models.Bookmark, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	return s.repo.Search(ctx, query)
}
