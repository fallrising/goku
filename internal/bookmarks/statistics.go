package bookmarks

import (
	"context"
	"fmt"
	"github.com/fallrising/goku-cli/pkg/models"
)

func (s *BookmarkService) GetStatistics(ctx context.Context) (*models.Statistics, error) {
	stats := &models.Statistics{}
	var err error

	stats.HostnameCounts, err = s.repo.CountByHostname(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname counts: %w", err)
	}

	stats.TagCounts, err = s.repo.CountByTag(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag counts: %w", err)
	}

	stats.LatestBookmarks, err = s.repo.GetLatest(ctx, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest bookmarks: %w", err)
	}

	stats.AccessibilityCounts, err = s.repo.CountAccessibility(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get accessibility counts: %w", err)
	}

	stats.TopHostnames, err = s.repo.TopHostnames(ctx, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to get top hostnames: %w", err)
	}

	stats.UniqueHostnames, err = s.repo.ListUniqueHostnames(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get unique hostnames: %w", err)
	}

	stats.CreatedLastWeek, err = s.repo.CountCreatedLastNDays(ctx, 7)
	if err != nil {
		return nil, fmt.Errorf("failed to get created counts for last week: %w", err)
	}

	return stats, nil
}
