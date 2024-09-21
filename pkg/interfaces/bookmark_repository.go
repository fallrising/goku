package interfaces

import (
	"context"

	"github.com/fallrising/goku-cli/pkg/models"
)

type BookmarkRepository interface {
	Create(ctx context.Context, bookmark *models.Bookmark) error
	GetByID(ctx context.Context, id int64) (*models.Bookmark, error)
	Update(ctx context.Context, bookmark *models.Bookmark) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int) ([]*models.Bookmark, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*models.Bookmark, error)
	ListAllTags(ctx context.Context) ([]string, error)
}
