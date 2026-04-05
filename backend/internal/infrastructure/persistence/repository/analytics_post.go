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

var _ domainrepo.XAnalyticsPostRepository = (*xAnalyticsPostRepository)(nil)

// allowedSortColumns is a whitelist of columns that can be used for sorting.
var allowedSortColumns = map[string]string{
	"impressions":  "impressions",
	"likes":        "likes",
	"engagements":  "engagements",
	"new_follows":  "new_follows",
	"bookmarks":    "bookmarks",
	"reposts":      "reposts",
	"posted_at":    "posted_at",
}

type xAnalyticsPostRepository struct {
	db *gorm.DB
}

func NewXAnalyticsPostRepository(db *gorm.DB) *xAnalyticsPostRepository {
	return &xAnalyticsPostRepository{db: db}
}

func (r *xAnalyticsPostRepository) getDB(ctx context.Context) *gorm.DB {
	return persistence.GetDB(ctx, r.db).WithContext(ctx)
}

func (r *xAnalyticsPostRepository) UpsertBatch(ctx context.Context, records []*entity.XAnalyticsPost) (inserted, updated int32, err error) {
	if len(records) == 0 {
		return 0, 0, nil
	}

	db := r.getDB(ctx)

	// Count existing records
	userID := records[0].UserID
	postIDs := make([]string, len(records))
	for i, rec := range records {
		postIDs[i] = rec.PostID
	}
	var existingCount int64
	db.Model(&model.XAnalyticsPost{}).Where("user_id = ? AND post_id IN ?", userID, postIDs).Count(&existingCount)

	cols := []string{
		"user_id", "post_id", "posted_at", "post_text", "post_url",
		"impressions", "likes", "engagements", "bookmarks", "shares",
		"new_follows", "replies", "reposts", "profile_visits",
		"detail_clicks", "url_clicks", "hashtag_clicks", "permalink_clicks", "updated_at",
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
			rec.UserID, rec.PostID, rec.PostedAt, rec.PostText, rec.PostURL,
			rec.Impressions, rec.Likes, rec.Engagements, rec.Bookmarks, rec.Shares,
			rec.NewFollows, rec.Replies, rec.Reposts, rec.ProfileVisits,
			rec.DetailClicks, rec.URLClicks, rec.HashtagClicks, rec.PermalinkClicks, now,
		)
	}

	updateCols := []string{
		"posted_at", "post_text", "post_url",
		"impressions", "likes", "engagements", "bookmarks", "shares",
		"new_follows", "replies", "reposts", "profile_visits",
		"detail_clicks", "url_clicks", "hashtag_clicks", "permalink_clicks", "updated_at",
	}
	updateSet := make([]string, len(updateCols))
	for i, col := range updateCols {
		updateSet[i] = fmt.Sprintf("%s = EXCLUDED.%s", col, col)
	}

	sql := fmt.Sprintf(
		"INSERT INTO x_analytics_posts (%s) VALUES %s ON CONFLICT (user_id, post_id) DO UPDATE SET %s",
		strings.Join(cols, ", "),
		strings.Join(valuePlaceholders, ", "),
		strings.Join(updateSet, ", "),
	)

	if err := db.Exec(sql, args...).Error; err != nil {
		return 0, 0, apperror.Wrap(apperror.CodeInternal, "failed to upsert post analytics", err)
	}

	total := int32(len(records))
	updatedCount := int32(existingCount)
	insertedCount := total - updatedCount

	return insertedCount, updatedCount, nil
}

func (r *xAnalyticsPostRepository) ListByUserID(ctx context.Context, userID string, sortBy string, limit int, startDate, endDate *time.Time) ([]*entity.XAnalyticsPost, int64, error) {
	// Validate sort column against whitelist
	sortCol, ok := allowedSortColumns[sortBy]
	if !ok {
		sortCol = "impressions"
	}

	q := r.getDB(ctx).Model(&model.XAnalyticsPost{}).Where("user_id = ?", userID)
	if startDate != nil {
		q = q.Where("posted_at >= ?", *startDate)
	}
	if endDate != nil {
		q = q.Where("posted_at <= ?", *endDate)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperror.Wrap(apperror.CodeInternal, "failed to count post analytics", err)
	}

	var models []model.XAnalyticsPost
	err := q.Order(fmt.Sprintf("%s DESC", sortCol)).Limit(limit).Find(&models).Error
	if err != nil {
		return nil, 0, apperror.Wrap(apperror.CodeInternal, "failed to list post analytics", err)
	}

	entities := make([]*entity.XAnalyticsPost, len(models))
	for i := range models {
		entities[i] = mapper.XAnalyticsPostToEntity(&models[i])
	}
	return entities, total, nil
}
