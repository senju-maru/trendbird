package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/trendbird/backend/internal/domain/apperror"
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/domain/gateway"
	"github.com/trendbird/backend/internal/domain/repository"
	"github.com/trendbird/backend/internal/domain/service"
)

// TrendFetchUsecase handles periodic trend detection by polling X API tweet counts.
type TrendFetchUsecase struct {
	topicRepo         repository.TopicRepository
	topicVolumeRepo   repository.TopicVolumeRepository
	spikeHistoryRepo  repository.SpikeHistoryRepository
	topicResearchRepo repository.TopicResearchRepository
	twitterGW         gateway.TwitterGateway
	aiGW              gateway.AIGateway
	zscoreSvc         *service.ZScoreService
	bearerToken       string
}

// NewTrendFetchUsecase creates a new TrendFetchUsecase.
func NewTrendFetchUsecase(
	topicRepo repository.TopicRepository,
	topicVolumeRepo repository.TopicVolumeRepository,
	spikeHistoryRepo repository.SpikeHistoryRepository,
	topicResearchRepo repository.TopicResearchRepository,
	twitterGW gateway.TwitterGateway,
	aiGW gateway.AIGateway,
	zscoreSvc *service.ZScoreService,
	bearerToken string,
) *TrendFetchUsecase {
	return &TrendFetchUsecase{
		topicRepo:         topicRepo,
		topicVolumeRepo:   topicVolumeRepo,
		spikeHistoryRepo:  spikeHistoryRepo,
		topicResearchRepo: topicResearchRepo,
		twitterGW:         twitterGW,
		aiGW:              aiGW,
		zscoreSvc:         zscoreSvc,
		bearerToken:       bearerToken,
	}
}

// Execute runs the trend fetch batch job.
func (u *TrendFetchUsecase) Execute(ctx context.Context) error {
	topics, err := u.topicRepo.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("list all topics: %w", err)
	}
	if len(topics) == 0 {
		slog.Info("no topics found, skipping trend fetch")
		return nil
	}
	slog.Info("starting trend fetch", "topic_count", len(topics))

	now := time.Now().UTC()
	startTime := now.Add(-6 * time.Hour).Truncate(time.Hour)
	lookbackFrom := now.Add(-7 * 24 * time.Hour)

	var successCount, skipCount, errorCount int

	for _, topic := range topics {
		if len(topic.Keywords) == 0 {
			slog.Warn("topic has no keywords, skipping", "topic_id", topic.ID, "topic_name", topic.Name)
			skipCount++
			continue
		}

		query := strings.Join(topic.Keywords, " OR ")

		points, err := u.twitterGW.GetTweetCounts(ctx, u.bearerToken, query, startTime)
		if err != nil {
			if apperror.IsCode(err, apperror.CodeResourceExhausted) {
				slog.Warn("rate limit hit, stopping trend fetch loop",
					"processed", successCount, "remaining", len(topics)-successCount-skipCount-errorCount)
				break
			}
			slog.Error("failed to get tweet counts", "topic_id", topic.ID, "error", err)
			errorCount++
			continue
		}

		if len(points) == 0 {
			skipCount++
			continue
		}

		// Convert data points to TopicVolume entities
		volumes := make([]*entity.TopicVolume, len(points))
		for i, p := range points {
			volumes[i] = &entity.TopicVolume{
				TopicID:   topic.ID,
				Timestamp: p.Start,
				Value:     int32(p.TweetCount),
			}
		}

		if err := u.topicVolumeRepo.BulkCreate(ctx, volumes); err != nil {
			slog.Error("failed to bulk create volumes", "topic_id", topic.ID, "error", err)
			errorCount++
			continue
		}

		// Get historical data for z-score calculation
		historical, err := u.topicVolumeRepo.ListByTopicIDAndRange(ctx, topic.ID, lookbackFrom, now)
		if err != nil {
			slog.Error("failed to list historical volumes", "topic_id", topic.ID, "error", err)
			errorCount++
			continue
		}

		if len(historical) == 0 {
			skipCount++
			continue
		}

		historicalValues := make([]int32, len(historical))
		for i, h := range historical {
			historicalValues[i] = h.Value
		}
		latestValue := historical[len(historical)-1].Value

		result := u.zscoreSvc.Calculate(historicalValues, latestValue)
		if result == nil {
			slog.Info("insufficient data for z-score, skipping", "topic_id", topic.ID)
			skipCount++
			continue
		}

		// Update topic fields
		previousStatus := topic.Status
		newStatus := entity.TopicStatus(result.Status)

		topic.Status = newStatus
		z := result.ZScore
		topic.ZScore = &z
		topic.CurrentVolume = result.CurrentVolume
		topic.BaselineVolume = int32(result.Mean)
		topic.ChangePercent = result.ChangePercent

		// Track spike start/end
		if newStatus == entity.TopicSpike && topic.SpikeStartedAt == nil {
			topic.SpikeStartedAt = &now
		} else if newStatus != entity.TopicSpike && topic.SpikeStartedAt != nil {
			topic.SpikeStartedAt = nil
		}

		if err := u.topicRepo.Update(ctx, topic); err != nil {
			slog.Error("failed to update topic", "topic_id", topic.ID, "error", err)
			errorCount++
			continue
		}

		// Create SpikeHistory on spike transition
		if newStatus == entity.TopicSpike && previousStatus != entity.TopicSpike {
			spike := &entity.SpikeHistory{
				TopicID:         topic.ID,
				Timestamp:       now,
				PeakZScore:      result.ZScore,
				Status:          entity.TopicSpike,
				Summary:         fmt.Sprintf("%s の言及数が急増 (z-score: %.2f, 現在: %d件/時)", topic.Name, result.ZScore, result.CurrentVolume),
				DurationMinutes: 0,
			}
			if err := u.spikeHistoryRepo.Create(ctx, spike); err != nil {
				slog.Error("failed to create spike history", "topic_id", topic.ID, "error", err)
			}
			// Best-effort: immediate web search on SPIKE transition
			u.researchTopicBestEffort(ctx, topic, entity.TriggerTypeSpike, now)
		} else if newStatus == entity.TopicRising && previousStatus != entity.TopicRising && previousStatus != entity.TopicSpike {
			// Stable → Rising の遷移のみ記録（Spike → Rising はクールダウンなので除外）
			rising := &entity.SpikeHistory{
				TopicID:         topic.ID,
				Timestamp:       now,
				PeakZScore:      result.ZScore,
				Status:          entity.TopicRising,
				Summary:         fmt.Sprintf("%s の言及数が上昇中 (z-score: %.2f, 現在: %d件/時)", topic.Name, result.ZScore, result.CurrentVolume),
				DurationMinutes: 0,
			}
			if err := u.spikeHistoryRepo.Create(ctx, rising); err != nil {
				slog.Error("failed to create rising history", "topic_id", topic.ID, "error", err)
			}
			// Best-effort: immediate web search on RISING transition
			u.researchTopicBestEffort(ctx, topic, entity.TriggerTypeRising, now)
		}

		successCount++
	}

	slog.Info("trend fetch completed",
		"success", successCount, "skipped", skipCount, "errors", errorCount)

	if errorCount > 0 && successCount == 0 {
		return fmt.Errorf("all topics failed (%d errors)", errorCount)
	}
	return nil
}

// researchTopicBestEffort performs a web search for a topic and saves the result.
// Failures are logged but do not interrupt the main trend-fetch flow.
func (u *TrendFetchUsecase) researchTopicBestEffort(ctx context.Context, topic *entity.Topic, triggerType entity.TriggerType, now time.Time) {
	if u.aiGW == nil || u.topicResearchRepo == nil {
		return
	}

	output, err := u.aiGW.ResearchTopic(ctx, gateway.ResearchTopicInput{
		TopicName: topic.Name,
		Keywords:  topic.Keywords,
	})
	if err != nil {
		slog.Warn("best-effort topic research failed", "topic_id", topic.ID, "trigger", triggerType, "error", err)
		return
	}

	research := &entity.TopicResearch{
		TopicID:     topic.ID,
		Query:       strings.Join(topic.Keywords, " "),
		Summary:     output.Summary,
		SourceURLs:  output.SourceURLs,
		TriggerType: triggerType,
		SearchedAt:  now,
	}
	if err := u.topicResearchRepo.Create(ctx, research); err != nil {
		slog.Warn("best-effort topic research save failed", "topic_id", topic.ID, "error", err)
	}
}
