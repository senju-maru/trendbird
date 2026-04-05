package di

import (
	"net/http"
	"time"

	"connectrpc.com/connect"
	"gorm.io/gorm"

	"github.com/trendbird/backend/internal/adapter/handler"
	"github.com/trendbird/backend/internal/adapter/middleware"
	"github.com/trendbird/backend/internal/domain/gateway"
	"github.com/trendbird/backend/internal/infrastructure/auth"
	"github.com/trendbird/backend/internal/infrastructure/config"
	"github.com/trendbird/backend/internal/infrastructure/external"
	"github.com/trendbird/backend/internal/infrastructure/persistence"
	"github.com/trendbird/backend/internal/infrastructure/persistence/repository"
	"github.com/trendbird/backend/internal/usecase"
)

// Container は全依存の組み立て結果を保持する。
type Container struct {
	// Connect RPC handlers
	AnalyticsHandler    *handler.AnalyticsHandler
	AuthHandler         *handler.AuthHandler
	AutoDMHandler       *handler.AutoDMHandler
	AutoReplyHandler    *handler.AutoReplyHandler
	DashboardHandler    *handler.DashboardHandler
	NotificationHandler *handler.NotificationHandler
	PostHandler         *handler.PostHandler
	SettingsHandler     *handler.SettingsHandler
	TopicHandler        *handler.TopicHandler
	TwitterHandler      *handler.TwitterHandler

	// HTTP handlers (non-RPC)
	AuthHTTPHandler http.Handler

	// Interceptors
	AuthInterceptor     connect.UnaryInterceptorFunc
	LoggingInterceptor  connect.UnaryInterceptorFunc
	RecoveryInterceptor connect.UnaryInterceptorFunc
	TraceIDInterceptor  connect.UnaryInterceptorFunc

	// DB (shutdown 用)
	DB *gorm.DB
}

// NewContainer は Config を受け取り、全依存を組み立てた Container を返す。
func NewContainer(cfg *config.Config) (*Container, error) {
	// --- Infrastructure ---
	db, err := persistence.NewDB(cfg)
	if err != nil {
		return nil, err
	}
	jwtSvc := auth.NewJWTService(cfg.JWTSecret, cfg.JWTExpiry)

	// --- External Clients ---
	twitterClient := external.NewTwitterClient(cfg.XClientID, cfg.XClientSecret, cfg.XRedirectURI)
	claudeClient := external.NewClaudeClient(cfg.AnthropicAPIKey)

	// --- Repositories ---
	userRepo := repository.NewUserRepository(db)
	tweetConnRepo := repository.NewTwitterConnectionRepository(db)
	topicRepo := repository.NewTopicRepository(db)
	userTopicRepo := repository.NewUserTopicRepository(db)
	userGenreRepo := repository.NewUserGenreRepository(db)
	genreRepo := repository.NewGenreRepository(db)
	postRepo := repository.NewPostRepository(db)
	genPostRepo := repository.NewGeneratedPostRepository(db)
	notiRepo := repository.NewNotificationRepository(db)
	notiSettingRepo := repository.NewNotificationSettingRepository(db)
	aiGenLogRepo := repository.NewAIGenerationLogRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	spikeHistRepo := repository.NewSpikeHistoryRepository(db)
	topicVolumeRepo := repository.NewTopicVolumeRepository(db)
	autoDMRuleRepo := repository.NewAutoDMRuleRepository(db)
	dmSentLogRepo := repository.NewDMSentLogRepository(db)
	txManager := persistence.NewTransactionManager(db)

	// --- Usecases ---
	authUC := usecase.NewAuthUsecase(
		userRepo, tweetConnRepo, notiSettingRepo, activityRepo,
		twitterClient, jwtSvc, txManager,
	)
	dashboardUC := usecase.NewDashboardUsecase(activityRepo, spikeHistRepo, aiGenLogRepo, topicRepo)
	notificationUC := usecase.NewNotificationUsecase(notiRepo)
	topicResearchRepo := repository.NewTopicResearchRepository(db)
	postUC := usecase.NewPostUsecase(
		topicRepo, postRepo, genPostRepo, aiGenLogRepo, activityRepo, tweetConnRepo,
		topicResearchRepo, claudeClient, twitterClient, txManager,
	)
	settingsUC := usecase.NewSettingsUsecase(userRepo, notiSettingRepo)
	topicUC := usecase.NewTopicUsecase(topicRepo, userTopicRepo, userGenreRepo, genreRepo, activityRepo, topicVolumeRepo, spikeHistRepo, claudeClient, twitterClient, cfg.XBearerToken, txManager)
	twitterUC := usecase.NewTwitterUsecase(tweetConnRepo, twitterClient)
	autoDMUC := usecase.NewAutoDMUsecase(autoDMRuleRepo, dmSentLogRepo)
	autoReplyRuleRepo := repository.NewAutoReplyRuleRepository(db)
	replySentLogRepo := repository.NewReplySentLogRepository(db)
	autoReplyUC := usecase.NewAutoReplyUsecase(autoReplyRuleRepo, replySentLogRepo)
	analyticsDailyRepo := repository.NewXAnalyticsDailyRepository(db)
	analyticsPostRepo := repository.NewXAnalyticsPostRepository(db)
	analyticsUC := usecase.NewAnalyticsUsecase(analyticsDailyRepo, analyticsPostRepo)

	// --- Handlers ---
	secure := cfg.CookieSecure

	return &Container{
		AnalyticsHandler:    handler.NewAnalyticsHandler(analyticsUC),
		AuthHandler:         handler.NewAuthHandler(authUC, secure, cfg.CookieDomain),
		AutoDMHandler:       handler.NewAutoDMHandler(autoDMUC),
		AutoReplyHandler:    handler.NewAutoReplyHandler(autoReplyUC),
		DashboardHandler:    handler.NewDashboardHandler(dashboardUC),
		NotificationHandler: handler.NewNotificationHandler(notificationUC),
		PostHandler:         handler.NewPostHandler(postUC),
		SettingsHandler:     handler.NewSettingsHandler(settingsUC),
		TopicHandler:        handler.NewTopicHandler(topicUC),
		TwitterHandler:      handler.NewTwitterHandler(twitterUC),
		AuthHTTPHandler:     handler.NewAuthHTTPHandler(twitterClient, cfg.CookieSecure, cfg.CookieDomain),

		AuthInterceptor:     middleware.NewAuthInterceptor(jwtSvc),
		LoggingInterceptor:  middleware.NewLoggingInterceptor(),
		RecoveryInterceptor: middleware.NewRecoveryInterceptor(),
		TraceIDInterceptor:  middleware.NewTraceIDInterceptor(),

		DB: db,
	}, nil
}

// TestDeps はテスト用に外部依存を差し替えるための構造体。
type TestDeps struct {
	DB        *gorm.DB
	TwitterGW gateway.TwitterGateway
	AIGW      gateway.AIGateway
	JWTSecret string
	JWTExpiry time.Duration
}

// NewContainerForTest は TestDeps から全依存を組み立てた Container を返す。
// 外部クライアント生成をスキップし、モックゲートウェイを使用する。
func NewContainerForTest(deps TestDeps) *Container {
	db := deps.DB
	jwtSvc := auth.NewJWTService(deps.JWTSecret, deps.JWTExpiry)

	// --- Repositories ---
	userRepo := repository.NewUserRepository(db)
	tweetConnRepo := repository.NewTwitterConnectionRepository(db)
	topicRepo := repository.NewTopicRepository(db)
	userTopicRepo := repository.NewUserTopicRepository(db)
	userGenreRepo := repository.NewUserGenreRepository(db)
	genreRepo := repository.NewGenreRepository(db)
	postRepo := repository.NewPostRepository(db)
	genPostRepo := repository.NewGeneratedPostRepository(db)
	notiRepo := repository.NewNotificationRepository(db)
	notiSettingRepo := repository.NewNotificationSettingRepository(db)
	aiGenLogRepo := repository.NewAIGenerationLogRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	spikeHistRepo := repository.NewSpikeHistoryRepository(db)
	topicVolumeRepo := repository.NewTopicVolumeRepository(db)
	autoDMRuleRepo := repository.NewAutoDMRuleRepository(db)
	dmSentLogRepo := repository.NewDMSentLogRepository(db)
	txManager := persistence.NewTransactionManager(db)

	// --- Usecases ---
	authUC := usecase.NewAuthUsecase(
		userRepo, tweetConnRepo, notiSettingRepo, activityRepo,
		deps.TwitterGW, jwtSvc, txManager,
	)
	dashboardUC := usecase.NewDashboardUsecase(activityRepo, spikeHistRepo, aiGenLogRepo, topicRepo)
	notificationUC := usecase.NewNotificationUsecase(notiRepo)
	topicResearchRepo := repository.NewTopicResearchRepository(db)
	postUC := usecase.NewPostUsecase(
		topicRepo, postRepo, genPostRepo, aiGenLogRepo, activityRepo, tweetConnRepo,
		topicResearchRepo, deps.AIGW, deps.TwitterGW, txManager,
	)
	settingsUC := usecase.NewSettingsUsecase(userRepo, notiSettingRepo)
	topicUC := usecase.NewTopicUsecase(topicRepo, userTopicRepo, userGenreRepo, genreRepo, activityRepo, topicVolumeRepo, spikeHistRepo, deps.AIGW, deps.TwitterGW, "", txManager)
	twitterUC := usecase.NewTwitterUsecase(tweetConnRepo, deps.TwitterGW)
	autoDMUC := usecase.NewAutoDMUsecase(autoDMRuleRepo, dmSentLogRepo)
	autoReplyRuleRepo := repository.NewAutoReplyRuleRepository(db)
	replySentLogRepo := repository.NewReplySentLogRepository(db)
	autoReplyUC := usecase.NewAutoReplyUsecase(autoReplyRuleRepo, replySentLogRepo)
	analyticsDailyRepo := repository.NewXAnalyticsDailyRepository(db)
	analyticsPostRepo := repository.NewXAnalyticsPostRepository(db)
	analyticsUC := usecase.NewAnalyticsUsecase(analyticsDailyRepo, analyticsPostRepo)

	// --- Handlers (secure=false: httptest は TLS なし) ---
	return &Container{
		AnalyticsHandler:    handler.NewAnalyticsHandler(analyticsUC),
		AuthHandler:         handler.NewAuthHandler(authUC, false, ""),
		AutoDMHandler:       handler.NewAutoDMHandler(autoDMUC),
		AutoReplyHandler:    handler.NewAutoReplyHandler(autoReplyUC),
		DashboardHandler:    handler.NewDashboardHandler(dashboardUC),
		NotificationHandler: handler.NewNotificationHandler(notificationUC),
		PostHandler:         handler.NewPostHandler(postUC),
		SettingsHandler:     handler.NewSettingsHandler(settingsUC),
		TopicHandler:        handler.NewTopicHandler(topicUC),
		TwitterHandler:      handler.NewTwitterHandler(twitterUC),
		AuthHTTPHandler:     handler.NewAuthHTTPHandler(deps.TwitterGW, false, ""),

		AuthInterceptor:     middleware.NewAuthInterceptor(jwtSvc),
		LoggingInterceptor:  middleware.NewLoggingInterceptor(),
		RecoveryInterceptor: middleware.NewRecoveryInterceptor(),
		TraceIDInterceptor:  middleware.NewTraceIDInterceptor(),

		DB: db,
	}
}
