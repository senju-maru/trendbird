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
)

// TopicResearchCollectionUsecase collects web search results for all topics.
type TopicResearchCollectionUsecase struct {
	topicRepo         repository.TopicRepository
	topicResearchRepo repository.TopicResearchRepository
	aiGW              gateway.AIGateway
}

// NewTopicResearchCollectionUsecase creates a new TopicResearchCollectionUsecase.
func NewTopicResearchCollectionUsecase(
	topicRepo repository.TopicRepository,
	topicResearchRepo repository.TopicResearchRepository,
	aiGW gateway.AIGateway,
) *TopicResearchCollectionUsecase {
	return &TopicResearchCollectionUsecase{
		topicRepo:         topicRepo,
		topicResearchRepo: topicResearchRepo,
		aiGW:              aiGW,
	}
}

// Execute runs the topic research collection batch job.
// It iterates over all topics, performs web search via Claude, and stores results.
func (u *TopicResearchCollectionUsecase) Execute(ctx context.Context) error {
	topics, err := u.topicRepo.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("list all topics: %w", err)
	}
	if len(topics) == 0 {
		slog.Info("no topics found, skipping topic research collection")
		return nil
	}
	slog.Info("starting topic research collection", "topic_count", len(topics))

	now := time.Now().UTC()
	var successCount, skipCount, errorCount int

	for _, topic := range topics {
		if len(topic.Keywords) == 0 {
			slog.Warn("topic has no keywords, skipping research", "topic_id", topic.ID, "topic_name", topic.Name)
			skipCount++
			continue
		}

		output, err := u.aiGW.ResearchTopic(ctx, gateway.ResearchTopicInput{
			TopicName: topic.Name,
			Keywords:  topic.Keywords,
		})
		if err != nil {
			if apperror.IsCode(err, apperror.CodeResourceExhausted) {
				slog.Warn("rate limit hit, stopping topic research collection loop",
					"processed", successCount, "remaining", len(topics)-successCount-skipCount-errorCount)
				break
			}
			slog.Error("failed to research topic", "topic_id", topic.ID, "error", err)
			errorCount++
			continue
		}

		research := &entity.TopicResearch{
			TopicID:     topic.ID,
			Query:       strings.Join(topic.Keywords, " "),
			Summary:     output.Summary,
			SourceURLs:  output.SourceURLs,
			TriggerType: entity.TriggerTypeBatch,
			SearchedAt:  now,
		}

		if err := u.topicResearchRepo.Create(ctx, research); err != nil {
			slog.Error("failed to save topic research", "topic_id", topic.ID, "error", err)
			errorCount++
			continue
		}

		slog.Info("topic research collected",
			"topic_id", topic.ID,
			"topic_name", topic.Name,
			"source_count", len(output.SourceURLs),
		)
		successCount++
	}

	slog.Info("topic research collection completed",
		"success", successCount, "skipped", skipCount, "errors", errorCount)

	if errorCount > 0 && successCount == 0 {
		return fmt.Errorf("all topics failed (%d errors)", errorCount)
	}
	return nil
}
