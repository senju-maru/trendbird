package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/domain/repository"
)

// SpikeNotificationUsecase handles batch sending of spike notifications.
type SpikeNotificationUsecase struct {
	spikeHistoryRepo    repository.SpikeHistoryRepository
	userTopicRepo       repository.UserTopicRepository
	notificationSetting repository.NotificationSettingRepository
	notificationRepo    repository.NotificationRepository
}

func NewSpikeNotificationUsecase(
	spikeHistoryRepo repository.SpikeHistoryRepository,
	userTopicRepo repository.UserTopicRepository,
	notificationSetting repository.NotificationSettingRepository,
	notificationRepo repository.NotificationRepository,
) *SpikeNotificationUsecase {
	return &SpikeNotificationUsecase{
		spikeHistoryRepo:    spikeHistoryRepo,
		userTopicRepo:       userTopicRepo,
		notificationSetting: notificationSetting,
		notificationRepo:    notificationRepo,
	}
}

// Execute processes unnotified spike histories and sends in-app notifications.
func (u *SpikeNotificationUsecase) Execute(ctx context.Context) error {
	spikes, err := u.spikeHistoryRepo.ListUnnotifiedByStatus(ctx, entity.TopicSpike)
	if err != nil {
		return fmt.Errorf("list unnotified spikes: %w", err)
	}
	if len(spikes) == 0 {
		slog.Info("no unnotified spikes found")
		return nil
	}
	slog.Info("found unnotified spikes", "count", len(spikes))

	// Group spikes by topic
	topicSpikes := make(map[string][]*entity.SpikeHistory)
	for _, s := range spikes {
		topicSpikes[s.TopicID] = append(topicSpikes[s.TopicID], s)
	}

	var notifiedIDs []string
	var totalErrors int

	for topicID, topicSpikeList := range topicSpikes {
		// Get users subscribed to this topic with notification enabled
		userIDs, err := u.userTopicRepo.ListUserIDsByTopicID(ctx, topicID, true)
		if err != nil {
			slog.Error("failed to list user ids for topic", "topic_id", topicID, "error", err)
			totalErrors++
			continue
		}
		if len(userIDs) == 0 {
			// No subscribers — still mark as notified to avoid re-processing
			for _, s := range topicSpikeList {
				notifiedIDs = append(notifiedIDs, s.ID)
			}
			continue
		}

		// Filter users by notification settings
		var appNotifyUserIDs []string
		for _, uid := range userIDs {
			setting, err := u.notificationSetting.FindByUserID(ctx, uid)
			if err != nil {
				continue
			}
			if !setting.SpikeEnabled {
				continue
			}

			appNotifyUserIDs = append(appNotifyUserIDs, uid)
		}

		// Process each spike in this topic
		for _, spike := range topicSpikeList {
			// Create in-app notification for all spike-enabled users
			if len(appNotifyUserIDs) > 0 {
				topicName := spike.TopicName
				topicStatus := spike.Status
				actionURL := fmt.Sprintf("/dashboard/%s", spike.TopicID)
				actionLabel := "詳細を確認"
				notification := &entity.Notification{
					Type:        entity.NotificationTrend,
					Title:       fmt.Sprintf("%s でスパイクを検知", topicName),
					Message:     spike.Summary,
					TopicID:     &spike.TopicID,
					TopicName:   &topicName,
					TopicStatus: &topicStatus,
					ActionURL:   &actionURL,
					ActionLabel: &actionLabel,
				}
				if err := u.notificationRepo.CreateForUsers(ctx, notification, appNotifyUserIDs); err != nil {
					slog.Error("failed to create in-app notification", "spike_id", spike.ID, "error", err)
				}
			}

			notifiedIDs = append(notifiedIDs, spike.ID)
		}
	}

	// Mark successfully processed spikes as notified
	if len(notifiedIDs) > 0 {
		if err := u.spikeHistoryRepo.MarkNotified(ctx, notifiedIDs, time.Now()); err != nil {
			return fmt.Errorf("mark notified: %w", err)
		}
		slog.Info("marked spikes as notified", "count", len(notifiedIDs))
	}

	if totalErrors > 0 && len(notifiedIDs) == 0 {
		return fmt.Errorf("all spike notifications failed (%d errors)", totalErrors)
	}

	return nil
}
