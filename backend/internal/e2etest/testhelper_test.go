package e2etest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"

	"github.com/trendbird/backend/gen/trendbird/v1/trendbirdv1connect"
	"github.com/trendbird/backend/internal/adapter/router"
	"github.com/trendbird/backend/internal/di"
	"github.com/trendbird/backend/internal/infrastructure/auth"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// testEnv: 1テスト分の環境を保持する構造体
// ---------------------------------------------------------------------------

type testEnv struct {
	server *httptest.Server
	db     *gorm.DB

	// Connect RPC clients (認証なし)
	authClient         trendbirdv1connect.AuthServiceClient
	autoDMClient       trendbirdv1connect.AutoDMServiceClient
	dashboardClient    trendbirdv1connect.DashboardServiceClient
	notificationClient trendbirdv1connect.NotificationServiceClient
	postClient         trendbirdv1connect.PostServiceClient
	settingsClient     trendbirdv1connect.SettingsServiceClient
	topicClient        trendbirdv1connect.TopicServiceClient
	twitterClient      trendbirdv1connect.TwitterServiceClient

	// Mocks
	mockTwitter *mockTwitterGateway
	mockAI      *mockAIGateway
}

// ---------------------------------------------------------------------------
// setupTest: テスト環境を構築する
// ---------------------------------------------------------------------------

func setupTest(t *testing.T) *testEnv {
	t.Helper()

	truncateAll(t, testDB)

	// Create mocks
	tw := newMockTwitterGateway()
	ai := newMockAIGateway()

	// Build container with test deps
	container := di.NewContainerForTest(di.TestDeps{
		DB:        testDB,
		TwitterGW: tw,
		AIGW:      ai,
		JWTSecret: testJWTSecret,
		JWTExpiry: testJWTExpiry,
	})

	// Start HTTP test server
	handler := router.New(container)
	server := httptest.NewServer(handler)
	t.Cleanup(func() { server.Close() })

	httpClient := server.Client()
	baseURL := server.URL

	env := &testEnv{
		server:      server,
		db:          testDB,
		mockTwitter: tw,
		mockAI:      ai,

		authClient:         trendbirdv1connect.NewAuthServiceClient(httpClient, baseURL),
		autoDMClient:       trendbirdv1connect.NewAutoDMServiceClient(httpClient, baseURL),
		dashboardClient:    trendbirdv1connect.NewDashboardServiceClient(httpClient, baseURL),
		notificationClient: trendbirdv1connect.NewNotificationServiceClient(httpClient, baseURL),
		postClient:         trendbirdv1connect.NewPostServiceClient(httpClient, baseURL),
		settingsClient:     trendbirdv1connect.NewSettingsServiceClient(httpClient, baseURL),
		topicClient:        trendbirdv1connect.NewTopicServiceClient(httpClient, baseURL),
		twitterClient:      trendbirdv1connect.NewTwitterServiceClient(httpClient, baseURL),
	}

	return env
}

// ---------------------------------------------------------------------------
// truncateAll: 全テーブルを TRUNCATE する
// ---------------------------------------------------------------------------

func truncateAll(t *testing.T, db *gorm.DB) {
	t.Helper()

	const query = `TRUNCATE TABLE
		dm_sent_logs, dm_pending_queue, auto_dm_rules,
		user_notifications, notifications, activities, ai_generation_logs, generated_posts, posts,
		posting_tips, spike_histories, topic_research, topic_volumes, user_topics, topics, user_genres, genres,
		twitter_connections, notification_settings, users
		CASCADE`

	if err := db.Exec(query).Error; err != nil {
		t.Fatalf("truncateAll: %v", err)
	}
}

// ---------------------------------------------------------------------------
// generateTestToken: テスト用 JWT トークンを生成する
// ---------------------------------------------------------------------------

func generateTestToken(t *testing.T, userID string) string {
	t.Helper()

	jwtSvc := auth.NewJWTService(testJWTSecret, testJWTExpiry)
	token, err := jwtSvc.GenerateToken(userID)
	if err != nil {
		t.Fatalf("generateTestToken: %v", err)
	}
	return token
}

// ---------------------------------------------------------------------------
// generateCustomToken: カスタム claims / secret / 署名方式で JWT を生成する
// ---------------------------------------------------------------------------

func generateCustomToken(t *testing.T, signingKey any, claims jwt.MapClaims, signingMethod jwt.SigningMethod) string {
	t.Helper()

	token := jwt.NewWithClaims(signingMethod, claims)
	signed, err := token.SignedString(signingKey)
	if err != nil {
		t.Fatalf("generateCustomToken: %v", err)
	}
	return signed
}

// ---------------------------------------------------------------------------
// connectClient: 認証付き Connect RPC クライアントを生成するジェネリクス関数
// ---------------------------------------------------------------------------

func connectClient[T any](t *testing.T, env *testEnv, userID string, newClientFn func(connect.HTTPClient, string, ...connect.ClientOption) T) T {
	t.Helper()

	token := generateTestToken(t, userID)
	return newClientFn(
		env.server.Client(),
		env.server.URL,
		connect.WithInterceptors(authTokenInterceptor(token)),
	)
}

// authTokenInterceptor は Authorization ヘッダーに Bearer トークンを付与するインターセプタ。
func authTokenInterceptor(token string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+token)
			return next(ctx, req)
		}
	}
}

// cookieInterceptor は Cookie ヘッダーを設定するインターセプタ。
func cookieInterceptor(cookieValue string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Cookie", cookieValue)
			return next(ctx, req)
		}
	}
}

// findSetCookie はレスポンスヘッダーから指定名の Set-Cookie 値を検索する。
func findSetCookie(header http.Header, cookieName string) (string, bool) {
	prefix := cookieName + "="
	for _, v := range header.Values("Set-Cookie") {
		if strings.HasPrefix(v, prefix) {
			return v, true
		}
	}
	return "", false
}

// assertSetCookieCleared は指定 cookie が Max-Age=0 でクリアされていることをアサートする。
func assertSetCookieCleared(t *testing.T, header http.Header, cookieName string) {
	t.Helper()
	v, ok := findSetCookie(header, cookieName)
	if !ok {
		t.Errorf("Set-Cookie %q not found in response headers", cookieName)
		return
	}
	if !strings.Contains(v, "Max-Age=0") {
		t.Errorf("expected Set-Cookie %q to contain Max-Age=0, got %s", cookieName, v)
	}
}

// ---------------------------------------------------------------------------
// ensureGenre: 指定 slug のジャンルが存在しなければ作成し、ID を返す（冪等）
// ---------------------------------------------------------------------------

func ensureGenre(t *testing.T, db *gorm.DB, slug string) string {
	t.Helper()

	var g model.Genre
	if err := db.Where("slug = ?", slug).First(&g).Error; err == nil {
		return g.ID
	}
	g = model.Genre{Slug: slug, Label: slug}
	if err := db.Create(&g).Error; err != nil {
		t.Fatalf("ensureGenre(%s): %v", slug, err)
	}
	return g.ID
}

// ---------------------------------------------------------------------------
// seed ヘルパー群
// ---------------------------------------------------------------------------

var seq atomic.Int64

func nextSeq() int64 {
	return seq.Add(1)
}

// --- seedUser ---

type userOption func(*model.User)

func withEmail(email string) userOption {
	return func(u *model.User) { u.Email = email }
}

func withTwitterID(twitterID string) userOption {
	return func(u *model.User) { u.TwitterID = twitterID }
}

func withTwitterHandle(handle string) userOption {
	return func(u *model.User) { u.TwitterHandle = handle }
}

func withName(name string) userOption {
	return func(u *model.User) { u.Name = name }
}

func withImage(image string) userOption {
	return func(u *model.User) { u.Image = image }
}

func withTutorialCompleted(v bool) userOption {
	return func(u *model.User) { u.TutorialCompleted = v }
}

func seedUser(t *testing.T, db *gorm.DB, opts ...userOption) *model.User {
	t.Helper()

	n := nextSeq()
	u := &model.User{
		TwitterID:         fmt.Sprintf("tw-%d", n),
		Name:              fmt.Sprintf("User %d", n),
		Email:             fmt.Sprintf("user%d@example.com", n),
		Image:             "",
		TwitterHandle:     fmt.Sprintf("user%d", n),
		TutorialCompleted: true,
	}
	for _, o := range opts {
		o(u)
	}
	if err := db.Create(u).Error; err != nil {
		t.Fatalf("seedUser: %v", err)
	}
	return u
}

// --- seedUserGenre ---

func seedUserGenre(t *testing.T, db *gorm.DB, userID string, genreSlug string) *model.UserGenre {
	t.Helper()
	genreID := ensureGenre(t, db, genreSlug)
	ug := &model.UserGenre{
		UserID:  userID,
		GenreID: genreID,
	}
	if err := db.Create(ug).Error; err != nil {
		t.Fatalf("seedUserGenre: %v", err)
	}
	return ug
}

// --- seedTopic ---

type topicOption func(*model.Topic)

func withTopicName(name string) topicOption {
	return func(tp *model.Topic) { tp.Name = name }
}

func withTopicStatus(status int32) topicOption {
	return func(tp *model.Topic) { tp.Status = status }
}

func withTopicGenre(slug string) topicOption {
	return func(tp *model.Topic) { tp.GenreID = slug }
}

func withTopicKeywords(keywords []string) topicOption {
	return func(tp *model.Topic) {
		b, _ := json.Marshal(keywords)
		tp.Keywords = string(b)
	}
}

func withTopicZScore(z float64) topicOption {
	return func(tp *model.Topic) { tp.ZScore = &z }
}

func withTopicCurrentVolume(v int32) topicOption {
	return func(tp *model.Topic) { tp.CurrentVolume = v }
}

func withTopicBaselineVolume(v int32) topicOption {
	return func(tp *model.Topic) { tp.BaselineVolume = v }
}

func withTopicChangePercent(v float64) topicOption {
	return func(tp *model.Topic) { tp.ChangePercent = v }
}

func withTopicContext(ctx string) topicOption {
	return func(tp *model.Topic) { tp.Context = &ctx }
}

func withTopicContextSummary(summary string) topicOption {
	return func(tp *model.Topic) { tp.ContextSummary = &summary }
}

func withTopicSpikeStartedAt(t time.Time) topicOption {
	return func(tp *model.Topic) { tp.SpikeStartedAt = &t }
}

func seedTopic(t *testing.T, db *gorm.DB, opts ...topicOption) *model.Topic {
	t.Helper()

	n := nextSeq()
	tp := &model.Topic{
		Name:     fmt.Sprintf("Topic %d", n),
		Keywords: `["keyword"]`,
		GenreID:  "technology", // slug placeholder; resolved below
		Status:   3,            // Stable
	}
	for _, o := range opts {
		o(tp)
	}
	// Resolve genre slug → UUID
	tp.GenreID = ensureGenre(t, db, tp.GenreID)
	if err := db.Create(tp).Error; err != nil {
		t.Fatalf("seedTopic: %v", err)
	}
	return tp
}

// --- seedUserTopic ---

type userTopicOption func(*model.UserTopic)

func withNotificationEnabled(v bool) userTopicOption {
	return func(ut *model.UserTopic) { ut.NotificationEnabled = v }
}

func withIsCreator(v bool) userTopicOption {
	return func(ut *model.UserTopic) { ut.IsCreator = v }
}

func seedUserTopic(t *testing.T, db *gorm.DB, userID string, topicID string, opts ...userTopicOption) *model.UserTopic {
	t.Helper()

	ut := &model.UserTopic{
		UserID:              userID,
		TopicID:             topicID,
		NotificationEnabled: true,
	}
	for _, o := range opts {
		o(ut)
	}
	if err := db.Create(ut).Error; err != nil {
		t.Fatalf("seedUserTopic: %v", err)
	}
	return ut
}

// --- seedPost ---

type postOption func(*model.Post)

func withPostStatus(status int32) postOption {
	return func(p *model.Post) { p.Status = status }
}

func withPostContent(content string) postOption {
	return func(p *model.Post) { p.Content = content }
}

func withPostTopicID(topicID string) postOption {
	return func(p *model.Post) { p.TopicID = &topicID }
}

func withPostTopicName(name string) postOption {
	return func(p *model.Post) { p.TopicName = &name }
}

func withPostPublishedAt(t time.Time) postOption {
	return func(p *model.Post) { p.PublishedAt = &t }
}

func withPostScheduledAt(t time.Time) postOption {
	return func(p *model.Post) { p.ScheduledAt = &t }
}

func withPostTweetURL(url string) postOption {
	return func(p *model.Post) { p.TweetURL = &url }
}

func seedPost(t *testing.T, db *gorm.DB, userID string, opts ...postOption) *model.Post {
	t.Helper()

	n := nextSeq()
	p := &model.Post{
		UserID:  userID,
		Content: fmt.Sprintf("Post content %d", n),
		Status:  1, // Draft
	}
	for _, o := range opts {
		o(p)
	}
	if err := db.Create(p).Error; err != nil {
		t.Fatalf("seedPost: %v", err)
	}
	return p
}

// --- seedNotification ---
// seedNotification creates a notification AND a user_notification record.
// withNotificationRead sets is_read on the user_notifications row.

type notificationOption func(n *model.Notification, isRead *bool)

func withNotificationType(nType int32) notificationOption {
	return func(n *model.Notification, _ *bool) { n.Type = nType }
}

func withNotificationRead(read bool) notificationOption {
	return func(_ *model.Notification, isRead *bool) { *isRead = read }
}

func seedNotification(t *testing.T, db *gorm.DB, userID string, opts ...notificationOption) *model.Notification {
	t.Helper()

	seq := nextSeq()
	isRead := false
	n := &model.Notification{
		Type:    1, // Trend
		Title:   fmt.Sprintf("Notification %d", seq),
		Message: fmt.Sprintf("Message %d", seq),
	}
	for _, o := range opts {
		o(n, &isRead)
	}
	if err := db.Create(n).Error; err != nil {
		t.Fatalf("seedNotification: %v", err)
	}
	// Create user_notification link
	un := &model.UserNotification{
		UserID:         userID,
		NotificationID: n.ID,
		IsRead:         isRead,
	}
	if err := db.Create(un).Error; err != nil {
		t.Fatalf("seedNotification (user_notification): %v", err)
	}
	return n
}

// --- seedNotificationSetting ---

type notificationSettingOption func(*model.NotificationSetting)

func withSpikeEnabled(v bool) notificationSettingOption {
	return func(ns *model.NotificationSetting) { ns.SpikeEnabled = v }
}

func withRisingEnabled(v bool) notificationSettingOption {
	return func(ns *model.NotificationSetting) { ns.RisingEnabled = v }
}

func seedNotificationSetting(t *testing.T, db *gorm.DB, userID string, opts ...notificationSettingOption) *model.NotificationSetting {
	t.Helper()

	ns := &model.NotificationSetting{
		UserID:        userID,
		SpikeEnabled:  true,
		RisingEnabled: true,
	}
	for _, o := range opts {
		o(ns)
	}
	// bool の false がゼロ値として GORM の Create でスキップされ DB デフォルト (true) が
	// 使われるのを防ぐため、Raw SQL で INSERT する。
	if err := db.Exec(
		`INSERT INTO notification_settings (user_id, spike_enabled, rising_enabled)
		 VALUES (?, ?, ?)
		 RETURNING id, created_at, updated_at`,
		ns.UserID, ns.SpikeEnabled, ns.RisingEnabled,
	).Error; err != nil {
		t.Fatalf("seedNotificationSetting: %v", err)
	}
	// RETURNING で取得できないため、改めてフェッチして ID 等を埋める
	if err := db.Where("user_id = ?", ns.UserID).First(ns).Error; err != nil {
		t.Fatalf("seedNotificationSetting refetch: %v", err)
	}
	return ns
}

// --- seedTwitterConnection ---

type twitterConnectionOption func(*model.TwitterConnection)

func withTwitterConnStatus(status int32) twitterConnectionOption {
	return func(tc *model.TwitterConnection) { tc.Status = status }
}

func withAccessToken(token string) twitterConnectionOption {
	return func(tc *model.TwitterConnection) { tc.AccessToken = token }
}

func withTokenExpiresAt(t time.Time) twitterConnectionOption {
	return func(tc *model.TwitterConnection) { tc.TokenExpiresAt = t }
}

func seedTwitterConnection(t *testing.T, db *gorm.DB, userID string, opts ...twitterConnectionOption) *model.TwitterConnection {
	t.Helper()

	now := time.Now()
	tc := &model.TwitterConnection{
		UserID:         userID,
		AccessToken:    "test-access-token",
		RefreshToken:   "test-refresh-token",
		TokenExpiresAt: now.Add(1 * time.Hour),
		Status:         3, // Connected
		ConnectedAt:    &now,
	}
	for _, o := range opts {
		o(tc)
	}
	if err := db.Create(tc).Error; err != nil {
		t.Fatalf("seedTwitterConnection: %v", err)
	}
	return tc
}

// --- seedActivity ---

type activityOption func(*model.Activity)

func withActivityType(aType int32) activityOption {
	return func(a *model.Activity) { a.Type = aType }
}

func withActivityDescription(desc string) activityOption {
	return func(a *model.Activity) { a.Description = desc }
}

func seedActivity(t *testing.T, db *gorm.DB, userID string, opts ...activityOption) *model.Activity {
	t.Helper()

	n := nextSeq()
	a := &model.Activity{
		UserID:      userID,
		Type:        7, // Login
		TopicName:   "",
		Description: fmt.Sprintf("Activity %d", n),
		Timestamp:   time.Now(),
	}
	for _, o := range opts {
		o(a)
	}
	if err := db.Create(a).Error; err != nil {
		t.Fatalf("seedActivity: %v", err)
	}
	return a
}

// --- seedAIGenerationLog ---

type aiGenerationLogOption func(*model.AIGenerationLog)

func withAIGenStyle(style int32) aiGenerationLogOption {
	return func(l *model.AIGenerationLog) { l.Style = style }
}

func withAIGenCount(count int32) aiGenerationLogOption {
	return func(l *model.AIGenerationLog) { l.Count = count }
}

func seedAIGenerationLog(t *testing.T, db *gorm.DB, userID string, opts ...aiGenerationLogOption) *model.AIGenerationLog {
	t.Helper()

	l := &model.AIGenerationLog{
		UserID: userID,
		Style:  1, // Casual
		Count:  1,
	}
	for _, o := range opts {
		o(l)
	}
	if err := db.Create(l).Error; err != nil {
		t.Fatalf("seedAIGenerationLog: %v", err)
	}
	return l
}

// --- seedSpikeHistory ---

type spikeHistoryOption func(*model.SpikeHistory)

func withSpikePeakZScore(z float64) spikeHistoryOption {
	return func(sh *model.SpikeHistory) { sh.PeakZScore = z }
}

func withSpikeStatus(status int32) spikeHistoryOption {
	return func(sh *model.SpikeHistory) { sh.Status = status }
}

func withSpikeSummary(summary string) spikeHistoryOption {
	return func(sh *model.SpikeHistory) { sh.Summary = summary }
}

func withSpikeTimestamp(ts time.Time) spikeHistoryOption {
	return func(sh *model.SpikeHistory) { sh.Timestamp = ts }
}

func withSpikeDurationMinutes(d int32) spikeHistoryOption {
	return func(sh *model.SpikeHistory) { sh.DurationMinutes = d }
}

func seedSpikeHistory(t *testing.T, db *gorm.DB, topicID string, opts ...spikeHistoryOption) *model.SpikeHistory {
	t.Helper()

	n := nextSeq()
	sh := &model.SpikeHistory{
		TopicID:         topicID,
		Timestamp:       time.Now(),
		PeakZScore:      3.5,
		Status:          1, // Spike
		Summary:         fmt.Sprintf("Spike summary %d", n),
		DurationMinutes: 30,
	}
	for _, o := range opts {
		o(sh)
	}
	if err := db.Create(sh).Error; err != nil {
		t.Fatalf("seedSpikeHistory: %v", err)
	}
	return sh
}

// --- seedTopicVolume ---

type topicVolumeOption func(*model.TopicVolume)

func withTopicVolumeTimestamp(ts time.Time) topicVolumeOption {
	return func(tv *model.TopicVolume) { tv.Timestamp = ts }
}

func withTopicVolumeValue(v int32) topicVolumeOption {
	return func(tv *model.TopicVolume) { tv.Value = v }
}

func seedTopicVolume(t *testing.T, db *gorm.DB, topicID string, opts ...topicVolumeOption) *model.TopicVolume {
	t.Helper()

	tv := &model.TopicVolume{
		TopicID:   topicID,
		Timestamp: time.Now(),
		Value:     100,
	}
	for _, o := range opts {
		o(tv)
	}
	if err := db.Create(tv).Error; err != nil {
		t.Fatalf("seedTopicVolume: %v", err)
	}
	return tv
}

// --- seedAutoDMRule ---

type autoDMRuleOption func(*model.AutoDMRule)

func withRuleEnabled(v bool) autoDMRuleOption {
	return func(r *model.AutoDMRule) { r.Enabled = v }
}

func withRuleTriggerKeywords(keywords []string) autoDMRuleOption {
	return func(r *model.AutoDMRule) { r.TriggerKeywords = model.StringArray(keywords) }
}

func withRuleTemplateMessage(msg string) autoDMRuleOption {
	return func(r *model.AutoDMRule) { r.TemplateMessage = msg }
}

func seedAutoDMRule(t *testing.T, db *gorm.DB, userID string, opts ...autoDMRuleOption) *model.AutoDMRule {
	t.Helper()

	n := nextSeq()
	r := &model.AutoDMRule{
		UserID:          userID,
		Enabled:         true,
		TriggerKeywords: model.StringArray{"keyword" + fmt.Sprintf("%d", n)},
		TemplateMessage: fmt.Sprintf("Auto DM template %d", n),
	}
	for _, o := range opts {
		o(r)
	}
	if err := db.Create(r).Error; err != nil {
		t.Fatalf("seedAutoDMRule: %v", err)
	}
	return r
}

// --- seedDMSentLog ---

type dmSentLogOption func(*model.DMSentLog)

func withDMLogRuleID(ruleID string) dmSentLogOption {
	return func(l *model.DMSentLog) { l.RuleID = ruleID }
}

func withDMLogRecipientID(recipientID string) dmSentLogOption {
	return func(l *model.DMSentLog) { l.RecipientTwitterID = recipientID }
}

func withDMLogTriggerKeyword(keyword string) dmSentLogOption {
	return func(l *model.DMSentLog) { l.TriggerKeyword = keyword }
}

func seedDMSentLog(t *testing.T, db *gorm.DB, userID string, ruleID string, opts ...dmSentLogOption) *model.DMSentLog {
	t.Helper()

	n := nextSeq()
	l := &model.DMSentLog{
		UserID:             userID,
		RuleID:             ruleID,
		RecipientTwitterID: fmt.Sprintf("recipient-%d", n),
		ReplyTweetID:       fmt.Sprintf("reply-tweet-%d", n),
		TriggerKeyword:     "keyword",
		DMText:             fmt.Sprintf("DM text %d", n),
		SentAt:             time.Now(),
	}
	for _, o := range opts {
		o(l)
	}
	if err := db.Create(l).Error; err != nil {
		t.Fatalf("seedDMSentLog: %v", err)
	}
	return l
}

// --- seedDMPendingQueue ---

type dmPendingQueueOption func(*model.DMPendingQueue)

func withDMPendingStatus(status int) dmPendingQueueOption {
	return func(q *model.DMPendingQueue) { q.Status = status }
}

func withDMPendingRecipientID(recipientID string) dmPendingQueueOption {
	return func(q *model.DMPendingQueue) { q.RecipientTwitterID = recipientID }
}

func withDMPendingReplyTweetID(tweetID string) dmPendingQueueOption {
	return func(q *model.DMPendingQueue) { q.ReplyTweetID = tweetID }
}

func withDMPendingTriggerKeyword(keyword string) dmPendingQueueOption {
	return func(q *model.DMPendingQueue) { q.TriggerKeyword = keyword }
}

func seedDMPendingQueue(t *testing.T, db *gorm.DB, userID string, ruleID string, opts ...dmPendingQueueOption) *model.DMPendingQueue {
	t.Helper()

	n := nextSeq()
	q := &model.DMPendingQueue{
		UserID:             userID,
		RuleID:             ruleID,
		RecipientTwitterID: fmt.Sprintf("recipient-%d", n),
		ReplyTweetID:       fmt.Sprintf("reply-tweet-%d", n),
		TriggerKeyword:     "keyword",
		Status:             1, // Pending
	}
	for _, o := range opts {
		o(q)
	}
	if err := db.Create(q).Error; err != nil {
		t.Fatalf("seedDMPendingQueue: %v", err)
	}
	return q
}

// ---------------------------------------------------------------------------
// httpGetWithCookie: Cookie 認証で HTTP GET を送信するヘルパー
// ---------------------------------------------------------------------------

func httpGetWithCookie(t *testing.T, env *testEnv, path string, userID string) *http.Response {
	t.Helper()

	token := generateTestToken(t, userID)
	req, err := http.NewRequest(http.MethodGet, env.server.URL+path, nil)
	if err != nil {
		t.Fatalf("httpGetWithCookie: create request: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: "tb_jwt", Value: token})

	resp, err := env.server.Client().Do(req)
	if err != nil {
		t.Fatalf("httpGetWithCookie: do request: %v", err)
	}
	t.Cleanup(func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	})

	return resp
}

// ---------------------------------------------------------------------------
// postWebhook: Webhook エンドポイントへ HTTP POST を送信するヘルパー
// ---------------------------------------------------------------------------

func postWebhook(t *testing.T, env *testEnv, path string, body []byte, headers map[string]string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodPost, env.server.URL+path, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("postWebhook: create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := env.server.Client().Do(req)
	if err != nil {
		t.Fatalf("postWebhook: do request: %v", err)
	}
	// body を読み切っておく（呼び出し側で Close 不要にする）
	t.Cleanup(func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	})

	return resp
}

// ---------------------------------------------------------------------------
// assertConnectCode: Connect RPC エラーコードのアサーション
// ---------------------------------------------------------------------------

func assertConnectCode(t *testing.T, err error, wantCode connect.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error with code %v, got nil", wantCode)
	}
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("expected connect.Error, got %T: %v", err, err)
	}
	if connectErr.Code() != wantCode {
		t.Errorf("code: want %v, got %v", wantCode, connectErr.Code())
	}
}
