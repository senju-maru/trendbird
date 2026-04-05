package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/trendbird/backend/internal/domain/apperror"
	"github.com/trendbird/backend/internal/domain/entity"
	domainrepo "github.com/trendbird/backend/internal/domain/repository"
	"github.com/trendbird/backend/internal/infrastructure/persistence"
	"github.com/trendbird/backend/internal/infrastructure/persistence/mapper"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

var _ domainrepo.XAnalyticsDailyRepository = (*xAnalyticsDailyRepository)(nil)

type xAnalyticsDailyRepository struct {
	db *gorm.DB
}

func NewXAnalyticsDailyRepository(db *gorm.DB) *xAnalyticsDailyRepository {
	return &xAnalyticsDailyRepository{db: db}
}

func (r *xAnalyticsDailyRepository) getDB(ctx context.Context) *gorm.DB {
	return persistence.GetDB(ctx, r.db).WithContext(ctx)
}

func (r *xAnalyticsDailyRepository) UpsertBatch(ctx context.Context, records []*entity.XAnalyticsDaily) (inserted, updated int32, err error) {
	if len(records) == 0 {
		return 0, 0, nil
	}

	db := r.getDB(ctx)

	// Count existing records to determine inserted vs updated
	var existingCount int64
	userID := records[0].UserID
	dates := make([]time.Time, len(records))
	for i, rec := range records {
		dates[i] = rec.Date
	}
	db.Model(&model.XAnalyticsDaily{}).Where("user_id = ? AND date IN ?", userID, dates).Count(&existingCount)

	// Build raw SQL for batch upsert
	cols := []string{
		"user_id", "date", "impressions", "likes", "engagements", "bookmarks",
		"shares", "new_follows", "unfollows", "replies", "reposts",
		"profile_visits", "posts_created", "video_views", "media_views", "updated_at",
	}

	valuePlaceholders := make([]string, len(records))
	args := make([]any, 0, len(records)*len(cols))
	now := time.Now()

	for i, rec := range records {
		placeholders := make([]string, len(cols))
		for j := range cols {
			placeholders[j] = fmt.Sprintf("$%d", i*len(cols)+j+1)
		}
		valuePlaceholders[i] = fmt.Sprintf("(%s)", strings.Join(placeholders, ", "))
		args = append(args,
			rec.UserID, rec.Date, rec.Impressions, rec.Likes, rec.Engagements, rec.Bookmarks,
			rec.Shares, rec.NewFollows, rec.Unfollows, rec.Replies, rec.Reposts,
			rec.ProfileVisits, rec.PostsCreated, rec.VideoViews, rec.MediaViews, now,
		)
	}

	updateCols := []string{
		"impressions", "likes", "engagements", "bookmarks", "shares",
		"new_follows", "unfollows", "replies", "reposts", "profile_visits",
		"posts_created", "video_views", "media_views", "updated_at",
	}
	updateSet := make([]string, len(updateCols))
	for i, col := range updateCols {
		updateSet[i] = fmt.Sprintf("%s = EXCLUDED.%s", col, col)
	}

	sql := fmt.Sprintf(
		"INSERT INTO x_analytics_daily (%s) VALUES %s ON CONFLICT (user_id, date) DO UPDATE SET %s",
		strings.Join(cols, ", "),
		strings.Join(valuePlaceholders, ", "),
		strings.Join(updateSet, ", "),
	)

	if err := db.Exec(sql, args...).Error; err != nil {
		return 0, 0, apperror.Wrap(apperror.CodeInternal, "failed to upsert daily analytics", err)
	}

	total := int32(len(records))
	updatedCount := int32(existingCount)
	insertedCount := total - updatedCount

	return insertedCount, updatedCount, nil
}

func (r *xAnalyticsDailyRepository) ListByDateRange(ctx context.Context, userID string, startDate, endDate time.Time) ([]*entity.XAnalyticsDaily, error) {
	var models []model.XAnalyticsDaily
	err := r.getDB(ctx).
		Where("user_id = ? AND date >= ? AND date <= ?", userID, startDate, endDate).
		Order("date ASC").
		Find(&models).Error
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list daily analytics", err)
	}

	entities := make([]*entity.XAnalyticsDaily, len(models))
	for i := range models {
		entities[i] = mapper.XAnalyticsDailyToEntity(&models[i])
	}
	return entities, nil
}

func (r *xAnalyticsDailyRepository) GetSummary(ctx context.Context, userID string, startDate, endDate time.Time) (*entity.AnalyticsSummary, error) {
	type summaryRow struct {
		TotalImpressions int64
		TotalLikes       int64
		TotalEngagements int64
		TotalNewFollows  int64
		TotalUnfollows   int64
		DaysCount        int32
		PostsCount       int32
	}

	var row summaryRow
	err := r.getDB(ctx).
		Model(&model.XAnalyticsDaily{}).
		Select(`
			COALESCE(SUM(impressions), 0) AS total_impressions,
			COALESCE(SUM(likes), 0) AS total_likes,
			COALESCE(SUM(engagements), 0) AS total_engagements,
			COALESCE(SUM(new_follows), 0) AS total_new_follows,
			COALESCE(SUM(unfollows), 0) AS total_unfollows,
			COUNT(*) AS days_count,
			COALESCE(SUM(posts_created), 0) AS posts_count
		`).
		Where("user_id = ? AND date >= ? AND date <= ?", userID, startDate, endDate).
		Scan(&row).Error
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to get analytics summary", err)
	}

	dailyData, err := r.ListByDateRange(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	return &entity.AnalyticsSummary{
		StartDate:        startDate.Format("2006-01-02"),
		EndDate:          endDate.Format("2006-01-02"),
		TotalImpressions: row.TotalImpressions,
		TotalLikes:       row.TotalLikes,
		TotalEngagements: row.TotalEngagements,
		TotalNewFollows:  row.TotalNewFollows,
		TotalUnfollows:   row.TotalUnfollows,
		DaysCount:        row.DaysCount,
		PostsCount:       row.PostsCount,
		DailyData:        dailyData,
	}, nil
}
