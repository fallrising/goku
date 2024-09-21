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
	List(ctx context.Context) ([]*models.Bookmark, error)
}
