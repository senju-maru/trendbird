package repository

import (
	"context"
	"time"

	"github.com/trendbird/backend/internal/domain/entity"
)

// XAnalyticsDailyRepository defines the persistence operations for daily analytics.
type XAnalyticsDailyRepository interface {
	UpsertBatch(ctx context.Context, records []*entity.XAnalyticsDaily) (inserted, updated int32, err error)
	ListByDateRange(ctx context.Context, userID string, startDate, endDate time.Time) ([]*entity.XAnalyticsDaily, error)
	GetSummary(ctx context.Context, userID string, startDate, endDate time.Time) (*entity.AnalyticsSummary, error)
}

// XAnalyticsPostRepository defines the persistence operations for per-post analytics.
type XAnalyticsPostRepository interface {
	UpsertBatch(ctx context.Context, records []*entity.XAnalyticsPost) (inserted, updated int32, err error)
	ListByUserID(ctx context.Context, userID string, sortBy string, limit int, startDate, endDate *time.Time) ([]*entity.XAnalyticsPost, int64, error)
}
