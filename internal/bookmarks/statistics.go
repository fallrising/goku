package bookmarks

import (
	"context"
	"github.com/fallrising/goku-cli/internal/database"
	"github.com/fallrising/goku-cli/pkg/models"
)

func (s *BookmarkService) GetStatistics(ctx context.Context) (*models.Statistics, error) {
	// Use DuckDB for statistics
	return s.duckDBStats.GetStatistics(ctx)
}

// Add a method to sync data from SQLite to DuckDB
func (s *BookmarkService) SyncToDuckDB() error {
	return s.duckDBStats.SyncFromSQLite(s.repo.(*database.Database))
}
