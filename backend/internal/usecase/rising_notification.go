package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/domain/repository"
)

// RisingNotificationUsecase handles batch sending of rising trend notifications.
type RisingNotificationUsecase struct {
	spikeHistoryRepo    repository.SpikeHistoryRepository
	userTopicRepo       repository.UserTopicRepository
	notificationSetting repository.NotificationSettingRepository
	notificationRepo    repository.NotificationRepository
}

func NewRisingNotificationUsecase(
	spikeHistoryRepo repository.SpikeHistoryRepository,
	userTopicRepo repository.UserTopicRepository,
	notificationSetting repository.NotificationSettingRepository,
	notificationRepo repository.NotificationRepository,
) *RisingNotificationUsecase {
	return &RisingNotificationUsecase{
		spikeHistoryRepo:    spikeHistoryRepo,
		userTopicRepo:       userTopicRepo,
		notificationSetting: notificationSetting,
		notificationRepo:    notificationRepo,
	}
}

// Execute processes unnotified rising histories and sends in-app notifications.
func (u *RisingNotificationUsecase) Execute(ctx context.Context) error {
	risings, err := u.spikeHistoryRepo.ListUnnotifiedByStatus(ctx, entity.TopicRising)
	if err != nil {
		return fmt.Errorf("list unnotified risings: %w", err)
	}
	if len(risings) == 0 {
		slog.Info("no unnotified rising records found")
		return nil
	}
	slog.Info("found unnotified rising records", "count", len(risings))

	// Group by topic
	topicRisings := make(map[string][]*entity.SpikeHistory)
	for _, r := range risings {
		topicRisings[r.TopicID] = append(topicRisings[r.TopicID], r)
	}

	var notifiedIDs []string
	var totalErrors int

	for topicID, risingList := range topicRisings {
		userIDs, err := u.userTopicRepo.ListUserIDsByTopicID(ctx, topicID, true)
		if err != nil {
			slog.Error("failed to list user ids for topic", "topic_id", topicID, "error", err)
			totalErrors++
			continue
		}
		if len(userIDs) == 0 {
			for _, r := range risingList {
				notifiedIDs = append(notifiedIDs, r.ID)
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
			if !setting.RisingEnabled {
				continue
			}

			appNotifyUserIDs = append(appNotifyUserIDs, uid)
		}

		for _, rising := range risingList {
			// Create in-app notification for all rising-enabled users
			if len(appNotifyUserIDs) > 0 {
				topicName := rising.TopicName
				topicStatus := rising.Status
				actionURL := fmt.Sprintf("/dashboard/%s", rising.TopicID)
				actionLabel := "詳細を確認"
				notification := &entity.Notification{
					Type:        entity.NotificationTrend,
					Title:       fmt.Sprintf("%s で上昇トレンドを検知", topicName),
					Message:     rising.Summary,
					TopicID:     &rising.TopicID,
					TopicName:   &topicName,
					TopicStatus: &topicStatus,
					ActionURL:   &actionURL,
					ActionLabel: &actionLabel,
				}
				if err := u.notificationRepo.CreateForUsers(ctx, notification, appNotifyUserIDs); err != nil {
					slog.Error("failed to create in-app notification", "rising_id", rising.ID, "error", err)
				}
			}

			notifiedIDs = append(notifiedIDs, rising.ID)
		}
	}

	// Mark successfully processed risings as notified
	if len(notifiedIDs) > 0 {
		if err := u.spikeHistoryRepo.MarkNotified(ctx, notifiedIDs, time.Now()); err != nil {
			return fmt.Errorf("mark notified: %w", err)
		}
		slog.Info("marked rising records as notified", "count", len(notifiedIDs))
	}

	if totalErrors > 0 && len(notifiedIDs) == 0 {
		return fmt.Errorf("all rising notifications failed (%d errors)", totalErrors)
	}

	return nil
}
