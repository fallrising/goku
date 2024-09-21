package bookmarks

import (
	"context"
	"fmt"
	"github.com/fallrising/goku-cli/internal/fetcher"

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

	// Fetch page content if title, description, or tags are not provided
	if bookmark.Title == "" || bookmark.Description == "" || len(bookmark.Tags) == 0 {
		content, err := fetcher.FetchPageContent(bookmark.URL)
		if err != nil {
			// Log the error but don't fail the bookmark creation
			fmt.Printf("Warning: Failed to fetch page content: %v\n", err)
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

func (s *BookmarkService) UpdateBookmark(ctx context.Context, bookmark *models.Bookmark) error {
	if bookmark.ID == 0 {
		return fmt.Errorf("bookmark ID is required")
	}

	return s.repo.Update(ctx, bookmark)
}

func (s *BookmarkService) DeleteBookmark(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *BookmarkService) ListBookmarks(ctx context.Context) ([]*models.Bookmark, error) {
	return s.repo.List(ctx)
}
