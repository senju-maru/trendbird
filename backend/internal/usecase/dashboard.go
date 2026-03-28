package usecase

import (
	"context"
	"time"

	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/domain/repository"
)

// DashboardStatsResult holds the aggregated dashboard statistics.
type DashboardStatsResult struct {
	Detections    int32
	Generations   int32
	LastCheckedAt *time.Time
}

// DashboardUsecase handles dashboard-related operations.
type DashboardUsecase struct {
	activityRepo     repository.ActivityRepository
	spikeHistoryRepo repository.SpikeHistoryRepository
	aiGenLogRepo     repository.AIGenerationLogRepository
	topicRepo        repository.TopicRepository
}

// NewDashboardUsecase creates a new DashboardUsecase.
func NewDashboardUsecase(
	activityRepo repository.ActivityRepository,
	spikeHistoryRepo repository.SpikeHistoryRepository,
	aiGenLogRepo repository.AIGenerationLogRepository,
	topicRepo repository.TopicRepository,
) *DashboardUsecase {
	return &DashboardUsecase{
		activityRepo:     activityRepo,
		spikeHistoryRepo: spikeHistoryRepo,
		aiGenLogRepo:     aiGenLogRepo,
		topicRepo:        topicRepo,
	}
}

// GetActivities returns a paginated list of activities for the user.
func (u *DashboardUsecase) GetActivities(ctx context.Context, userID string, limit, offset int) ([]*entity.Activity, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return u.activityRepo.ListByUserID(ctx, userID, limit, offset)
}

// GetStats returns aggregated monthly statistics for the user's dashboard.
func (u *DashboardUsecase) GetStats(ctx context.Context, userID string) (*DashboardStatsResult, error) {
	detections, err := u.spikeHistoryRepo.CountByUserIDCurrentMonth(ctx, userID)
	if err != nil {
		return nil, err
	}

	generations, err := u.aiGenLogRepo.CountByUserIDCurrentMonth(ctx, userID)
	if err != nil {
		return nil, err
	}

	lastCheckedAt, err := u.topicRepo.GetLatestUpdatedAtByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &DashboardStatsResult{
		Detections:    detections,
		Generations:   generations,
		LastCheckedAt: lastCheckedAt,
	}, nil
}
