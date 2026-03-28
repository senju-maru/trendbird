package di

import (
	"context"
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/trendbird/backend/internal/infrastructure/config"
	"github.com/trendbird/backend/internal/infrastructure/external"
	"github.com/trendbird/backend/internal/infrastructure/persistence"
	"github.com/trendbird/backend/internal/infrastructure/persistence/repository"
	"github.com/trendbird/backend/internal/usecase"
)

// MCPContainer は MCP サーバーに必要な依存の組み立て結果を保持する。
type MCPContainer struct {
	PostUC         *usecase.PostUsecase
	TopicUC        *usecase.TopicUsecase
	NotificationUC *usecase.NotificationUsecase
	AutoReplyUC    *usecase.AutoReplyUsecase
	AutoDMUC       *usecase.AutoDMUsecase

	// UserID はローカルユーザーの ID。MCP_USER_ID または DB から自動検出。
	UserID string

	DB *gorm.DB
}

// NewMCPContainer は MCPConfig を受け取り、MCP 用の依存を組み立てた MCPContainer を返す。
func NewMCPContainer(cfg *config.MCPConfig) (*MCPContainer, error) {
	db, err := persistence.NewMCPDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// --- Repositories ---
	userRepo := repository.NewUserRepository(db)
	topicRepo := repository.NewTopicRepository(db)
	userTopicRepo := repository.NewUserTopicRepository(db)
	userGenreRepo := repository.NewUserGenreRepository(db)
	genreRepo := repository.NewGenreRepository(db)
	postRepo := repository.NewPostRepository(db)
	genPostRepo := repository.NewGeneratedPostRepository(db)
	notiRepo := repository.NewNotificationRepository(db)
	aiGenLogRepo := repository.NewAIGenerationLogRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	spikeHistRepo := repository.NewSpikeHistoryRepository(db)
	topicVolumeRepo := repository.NewTopicVolumeRepository(db)
	tweetConnRepo := repository.NewTwitterConnectionRepository(db)
	topicResearchRepo := repository.NewTopicResearchRepository(db)
	autoDMRuleRepo := repository.NewAutoDMRuleRepository(db)
	dmSentLogRepo := repository.NewDMSentLogRepository(db)
	autoReplyRuleRepo := repository.NewAutoReplyRuleRepository(db)
	replySentLogRepo := repository.NewReplySentLogRepository(db)
	txManager := persistence.NewTransactionManager(db)

	// --- External Clients ---
	twitterClient := external.NewTwitterClient(cfg.XClientID, cfg.XClientSecret, cfg.XRedirectURI)
	claudeClient := external.NewClaudeClient(cfg.AnthropicAPIKey)

	// --- Usecases ---
	postUC := usecase.NewPostUsecase(
		topicRepo, postRepo, genPostRepo, aiGenLogRepo, activityRepo, tweetConnRepo,
		topicResearchRepo, claudeClient, twitterClient, txManager,
	)
	topicUC := usecase.NewTopicUsecase(
		topicRepo, userTopicRepo, userGenreRepo, genreRepo, activityRepo,
		topicVolumeRepo, spikeHistRepo, claudeClient, twitterClient, cfg.XBearerToken, txManager,
	)
	notificationUC := usecase.NewNotificationUsecase(notiRepo)
	autoReplyUC := usecase.NewAutoReplyUsecase(autoReplyRuleRepo, replySentLogRepo)
	autoDMUC := usecase.NewAutoDMUsecase(autoDMRuleRepo, dmSentLogRepo)

	// --- User Resolution ---
	userID := cfg.MCPUserID
	if userID == "" {
		user, err := userRepo.FindFirst(context.Background())
		if err != nil {
			return nil, fmt.Errorf("ユーザーが見つかりません。先にブラウザ (http://localhost:3000) でXログインしてください: %w", err)
		}
		userID = user.ID
		slog.Info("MCP: auto-detected user", "user_id", userID, "name", user.Name)
	}

	return &MCPContainer{
		PostUC:         postUC,
		TopicUC:        topicUC,
		NotificationUC: notificationUC,
		AutoReplyUC:    autoReplyUC,
		AutoDMUC:       autoDMUC,
		UserID:         userID,
		DB:             db,
	}, nil
}

// Close は DB 接続を解放する。
func (c *MCPContainer) Close() error {
	sqlDB, err := c.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
