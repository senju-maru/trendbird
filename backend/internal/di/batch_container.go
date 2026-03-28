package di

import (
	"gorm.io/gorm"

	"github.com/trendbird/backend/internal/domain/service"
	"github.com/trendbird/backend/internal/infrastructure/config"
	"github.com/trendbird/backend/internal/infrastructure/external"
	"github.com/trendbird/backend/internal/infrastructure/persistence"
	"github.com/trendbird/backend/internal/infrastructure/persistence/repository"
	"github.com/trendbird/backend/internal/usecase"
)

// BatchContainer はバッチジョブに必要な依存の組み立て結果を保持する。
type BatchContainer struct {
	SpikeNotificationUC         *usecase.SpikeNotificationUsecase
	RisingNotificationUC        *usecase.RisingNotificationUsecase
	TrendFetchUC                *usecase.TrendFetchUsecase
	ScheduledPublishUC          *usecase.ScheduledPublishUsecase
	AutoDMBatchUC               *usecase.AutoDMBatchUsecase
	AutoReplyBatchUC            *usecase.AutoReplyBatchUsecase
	TopicResearchCollectionUC   *usecase.TopicResearchCollectionUsecase

	DB *gorm.DB
}

// NewBatchContainer は BatchConfig を受け取り、バッチ用の依存を組み立てた BatchContainer を返す。
func NewBatchContainer(cfg *config.BatchConfig) (*BatchContainer, error) {
	db, err := persistence.NewBatchDB(cfg)
	if err != nil {
		return nil, err
	}
	return newBatchContainer(cfg, db)
}

// NewBatchContainerWithDB は外部で作成した DB 接続を使ってバッチ用の依存を組み立てた BatchContainer を返す。
// スケジューラなど、DB プールサイズを呼び出し側で制御したい場合に使用する。
func NewBatchContainerWithDB(cfg *config.BatchConfig, db *gorm.DB) (*BatchContainer, error) {
	return newBatchContainer(cfg, db)
}

func newBatchContainer(cfg *config.BatchConfig, db *gorm.DB) (*BatchContainer, error) {
	// --- Repositories ---
	userRepo := repository.NewUserRepository(db)
	topicRepo := repository.NewTopicRepository(db)
	userTopicRepo := repository.NewUserTopicRepository(db)
	topicVolumeRepo := repository.NewTopicVolumeRepository(db)
	notiRepo := repository.NewNotificationRepository(db)
	notiSettingRepo := repository.NewNotificationSettingRepository(db)
	spikeHistRepo := repository.NewSpikeHistoryRepository(db)

	// --- Usecases ---
	spikeNotificationUC := usecase.NewSpikeNotificationUsecase(
		spikeHistRepo, userTopicRepo, notiSettingRepo, notiRepo,
	)
	risingNotificationUC := usecase.NewRisingNotificationUsecase(
		spikeHistRepo, userTopicRepo, notiSettingRepo, notiRepo,
	)

	// --- Claude API & Topic Research ---
	claudeClient := external.NewClaudeClient(cfg.AnthropicAPIKey)
	topicResearchRepo := repository.NewTopicResearchRepository(db)
	topicResearchCollectionUC := usecase.NewTopicResearchCollectionUsecase(
		topicRepo, topicResearchRepo, claudeClient,
	)

	// --- Trend Fetch ---
	twitterClient := external.NewTwitterClient("", "", "")
	zscoreSvc := service.NewZScoreService()
	trendFetchUC := usecase.NewTrendFetchUsecase(
		topicRepo, topicVolumeRepo, spikeHistRepo, topicResearchRepo,
		twitterClient, claudeClient, zscoreSvc, cfg.XBearerToken,
	)

	// --- Scheduled Publish ---
	oauthTwitterClient := external.NewTwitterClient(cfg.XClientID, cfg.XClientSecret, cfg.XRedirectURI)
	postRepo := repository.NewPostRepository(db)
	tweetConnRepo := repository.NewTwitterConnectionRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	scheduledPublishUC := usecase.NewScheduledPublishUsecase(
		postRepo, tweetConnRepo, activityRepo, oauthTwitterClient,
	)

	// --- Auto DM Batch ---
	autoDMRuleRepo := repository.NewAutoDMRuleRepository(db)
	dmSentLogRepo := repository.NewDMSentLogRepository(db)
	dmPendingRepo := repository.NewDMPendingQueueRepository(db)
	autoDMBatchUC := usecase.NewAutoDMBatchUsecase(
		autoDMRuleRepo, dmSentLogRepo, dmPendingRepo,
		tweetConnRepo, userRepo, oauthTwitterClient,
	)

	// --- Auto Reply Batch ---
	autoReplyRuleRepo := repository.NewAutoReplyRuleRepository(db)
	replySentLogRepo := repository.NewReplySentLogRepository(db)
	replyPendingRepo := repository.NewReplyPendingQueueRepository(db)
	autoReplyBatchUC := usecase.NewAutoReplyBatchUsecase(
		autoReplyRuleRepo, replySentLogRepo, replyPendingRepo,
		tweetConnRepo, userRepo, oauthTwitterClient,
	)

	return &BatchContainer{
		SpikeNotificationUC:       spikeNotificationUC,
		RisingNotificationUC:      risingNotificationUC,
		TrendFetchUC:              trendFetchUC,
		ScheduledPublishUC:        scheduledPublishUC,
		AutoDMBatchUC:             autoDMBatchUC,
		AutoReplyBatchUC:          autoReplyBatchUC,
		TopicResearchCollectionUC: topicResearchCollectionUC,
		DB:                        db,
	}, nil
}

// Close は DB 接続を解放する。
func (c *BatchContainer) Close() error {
	sqlDB, err := c.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
