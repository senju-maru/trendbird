package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/domain/repository"
)

// RecordActivity records an activity log, logging on failure without returning an error (best-effort).
func RecordActivity(ctx context.Context, repo repository.ActivityRepository, userID string, activityType entity.ActivityType, topicName, description string) {
	activity := &entity.Activity{
		UserID:      userID,
		Type:        activityType,
		TopicName:   topicName,
		Description: description,
		Timestamp:   time.Now(),
	}
	if err := repo.Create(ctx, activity); err != nil {
		slog.Warn("failed to record activity", "userID", userID, "type", activityType, "error", err)
	}
}
