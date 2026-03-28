package e2etest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	trendbirdv1 "github.com/trendbird/backend/gen/trendbird/v1"
	"github.com/trendbird/backend/gen/trendbird/v1/trendbirdv1connect"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

// ---------------------------------------------------------------------------
// TestDashboardService_GetActivities
// ---------------------------------------------------------------------------

func TestDashboardService_GetActivities(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)

		user := seedUser(t, env.db)
		// 5件のアクティビティを seed（時刻をずらして順序を確認、type と topic_name を明示的に設定）
		seedActivities := make([]struct {
			aType       int32
			topicName   string
			description string
		}, 5)
		seedActivities[0] = struct {
			aType       int32
			topicName   string
			description string
		}{5, "AI News", "トピック AI News を作成しました"}
		seedActivities[1] = struct {
			aType       int32
			topicName   string
			description string
		}{6, "Old Topic", "トピック Old Topic を削除しました"}
		seedActivities[2] = struct {
			aType       int32
			topicName   string
			description string
		}{7, "", "ログインしました"}
		seedActivities[3] = struct {
			aType       int32
			topicName   string
			description string
		}{1, "Go言語", "Go言語 のスパイクを検知しました"}
		seedActivities[4] = struct {
			aType       int32
			topicName   string
			description string
		}{3, "Go言語", "Go言語 の投稿文を生成しました"}

		for _, sa := range seedActivities {
			seedActivity(t, env.db, user.ID,
				withActivityType(sa.aType),
				func(topicName string) activityOption {
					return func(a *model.Activity) { a.TopicName = topicName }
				}(sa.topicName),
				withActivityDescription(sa.description),
			)
			time.Sleep(10 * time.Millisecond) // timestamp の順序を保証
		}

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewDashboardServiceClient)
		resp, err := client.GetActivities(context.Background(), connect.NewRequest(&trendbirdv1.GetActivitiesRequest{}))
		if err != nil {
			t.Fatalf("GetActivities: %v", err)
		}

		activities := resp.Msg.GetActivities()
		if got := len(activities); got != 5 {
			t.Fatalf("activities count: want 5, got %d", got)
		}

		// DESC 順 (最新が先頭) を検証
		for i := 0; i < len(activities)-1; i++ {
			cur, _ := time.Parse(time.RFC3339, activities[i].GetTimestamp())
			next, _ := time.Parse(time.RFC3339, activities[i+1].GetTimestamp())
			if cur.Before(next) {
				t.Errorf("activities not in DESC order: [%d]=%v < [%d]=%v", i, cur, i+1, next)
			}
		}

		// 個別フィールド検証 (B4): DESC 順なので最新（seedActivities[4]）が先頭
		// 先頭要素の type, topic_name, description を検証
		first := activities[0]
		if first.GetDescription() != seedActivities[4].description {
			t.Errorf("activities[0].description: want %q, got %q", seedActivities[4].description, first.GetDescription())
		}
		if first.GetTopicName() != seedActivities[4].topicName {
			t.Errorf("activities[0].topic_name: want %q, got %q", seedActivities[4].topicName, first.GetTopicName())
		}
		if int32(first.GetType()) != seedActivities[4].aType {
			t.Errorf("activities[0].type: want %d, got %d", seedActivities[4].aType, first.GetType())
		}

		// timestamp が RFC3339 フォーマットであることを検証
		for i, a := range activities {
			if _, parseErr := time.Parse(time.RFC3339, a.GetTimestamp()); parseErr != nil {
				t.Errorf("activities[%d].timestamp parse error: %v", i, parseErr)
			}
		}
	})

	t.Run("empty", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewDashboardServiceClient)
		resp, err := client.GetActivities(context.Background(), connect.NewRequest(&trendbirdv1.GetActivitiesRequest{}))
		if err != nil {
			t.Fatalf("GetActivities: %v", err)
		}

		if got := len(resp.Msg.GetActivities()); got != 0 {
			t.Errorf("activities count: want 0, got %d", got)
		}
	})

	t.Run("cross_user_isolation", func(t *testing.T) {
		env := setupTest(t)

		userA := seedUser(t, env.db)
		userB := seedUser(t, env.db)

		// userA に2件
		seedActivity(t, env.db, userA.ID, withActivityDescription("A activity 1"))
		seedActivity(t, env.db, userA.ID, withActivityDescription("A activity 2"))
		// userB に3件
		seedActivity(t, env.db, userB.ID, withActivityDescription("B activity 1"))
		seedActivity(t, env.db, userB.ID, withActivityDescription("B activity 2"))
		seedActivity(t, env.db, userB.ID, withActivityDescription("B activity 3"))

		client := connectClient(t, env, userA.ID, trendbirdv1connect.NewDashboardServiceClient)
		resp, err := client.GetActivities(context.Background(), connect.NewRequest(&trendbirdv1.GetActivitiesRequest{}))
		if err != nil {
			t.Fatalf("GetActivities: %v", err)
		}

		activities := resp.Msg.GetActivities()
		if got := len(activities); got != 2 {
			t.Fatalf("activities count: want 2, got %d", got)
		}
		for _, a := range activities {
			if a.GetDescription() != "A activity 1" && a.GetDescription() != "A activity 2" {
				t.Errorf("unexpected activity for userA: %q", a.GetDescription())
			}
		}
	})

	t.Run("limit_20", func(t *testing.T) {
		env := setupTest(t)

		user := seedUser(t, env.db)
		for i := 0; i < 25; i++ {
			seedActivity(t, env.db, user.ID,
				withActivityDescription(fmt.Sprintf("activity %d", i)),
			)
			time.Sleep(10 * time.Millisecond)
		}

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewDashboardServiceClient)
		resp, err := client.GetActivities(context.Background(), connect.NewRequest(&trendbirdv1.GetActivitiesRequest{}))
		if err != nil {
			t.Fatalf("GetActivities: %v", err)
		}

		if got := len(resp.Msg.GetActivities()); got != 20 {
			t.Errorf("activities count: want 20 (limit), got %d", got)
		}
	})

	t.Run("all_activity_types", func(t *testing.T) {
		env := setupTest(t)

		user := seedUser(t, env.db)

		// Proto 定義の6種類（1=Spike 〜 6=TopicRemoved）を各1件 seed
		typeTests := []struct {
			aType       int32
			topicName   string
			description string
			wantType    trendbirdv1.ActivityType
		}{
			{1, "Topic A", "Topic A のスパイクを検知", trendbirdv1.ActivityType_ACTIVITY_TYPE_SPIKE},
			{2, "Topic B", "Topic B の上昇傾向を検知", trendbirdv1.ActivityType_ACTIVITY_TYPE_RISING},
			{3, "Topic C", "Topic C の投稿文を生成", trendbirdv1.ActivityType_ACTIVITY_TYPE_AI_GENERATED},
			{4, "Topic D", "Topic D に投稿を実行", trendbirdv1.ActivityType_ACTIVITY_TYPE_POSTED},
			{5, "Topic E", "Topic E を追加", trendbirdv1.ActivityType_ACTIVITY_TYPE_TOPIC_ADDED},
			{6, "Topic F", "Topic F を削除", trendbirdv1.ActivityType_ACTIVITY_TYPE_TOPIC_REMOVED},
		}

		for _, tt := range typeTests {
			seedActivity(t, env.db, user.ID,
				withActivityType(tt.aType),
				func(topicName string) activityOption {
					return func(a *model.Activity) { a.TopicName = topicName }
				}(tt.topicName),
				withActivityDescription(tt.description),
			)
			time.Sleep(10 * time.Millisecond)
		}

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewDashboardServiceClient)
		resp, err := client.GetActivities(context.Background(), connect.NewRequest(&trendbirdv1.GetActivitiesRequest{}))
		if err != nil {
			t.Fatalf("GetActivities: %v", err)
		}

		activities := resp.Msg.GetActivities()
		if got := len(activities); got != 6 {
			t.Fatalf("activities count: want 6, got %d", got)
		}

		// DESC 順なので逆順にマッチ
		for i, a := range activities {
			want := typeTests[len(typeTests)-1-i]
			if a.GetType() != want.wantType {
				t.Errorf("activities[%d].type: want %v, got %v", i, want.wantType, a.GetType())
			}
			if a.GetTopicName() != want.topicName {
				t.Errorf("activities[%d].topic_name: want %q, got %q", i, want.topicName, a.GetTopicName())
			}
			if a.GetDescription() != want.description {
				t.Errorf("activities[%d].description: want %q, got %q", i, want.description, a.GetDescription())
			}
		}
	})

	t.Run("topic_create_records_activity", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		seedUserGenre(t, env.db, user.ID, "technology")

		topicClient := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		_, err := topicClient.CreateTopic(context.Background(), connect.NewRequest(&trendbirdv1.CreateTopicRequest{
			Name:     "Activity Test Topic",
			Keywords: []string{"go", "test"},
			Genre:    "technology",
		}))
		if err != nil {
			t.Fatalf("CreateTopic: %v", err)
		}

		dashClient := connectClient(t, env, user.ID, trendbirdv1connect.NewDashboardServiceClient)
		resp, err := dashClient.GetActivities(context.Background(), connect.NewRequest(&trendbirdv1.GetActivitiesRequest{}))
		if err != nil {
			t.Fatalf("GetActivities: %v", err)
		}

		// ACTIVITY_TYPE_TOPIC_ADDED (type=5) が存在すること
		found := false
		for _, a := range resp.Msg.GetActivities() {
			if a.GetType() == trendbirdv1.ActivityType_ACTIVITY_TYPE_TOPIC_ADDED {
				found = true
				if a.GetTopicName() != "Activity Test Topic" {
					t.Errorf("topic_name: want %q, got %q", "Activity Test Topic", a.GetTopicName())
				}
				break
			}
		}
		if !found {
			t.Error("expected ACTIVITY_TYPE_TOPIC_ADDED activity, not found")
		}
	})

	t.Run("publish_records_activity", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		topic := seedTopic(t, env.db, withTopicName("Publish Topic"))
		seedUserTopic(t, env.db, user.ID, topic.ID)
		seedTwitterConnection(t, env.db, user.ID)

		post := seedPost(t, env.db, user.ID,
			withPostStatus(1), // Draft
			withPostContent("Test publish content"),
			withPostTopicID(topic.ID),
			withPostTopicName("Publish Topic"),
		)

		postClient := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		_, err := postClient.PublishPost(context.Background(), connect.NewRequest(&trendbirdv1.PublishPostRequest{
			Id: post.ID,
		}))
		if err != nil {
			t.Fatalf("PublishPost: %v", err)
		}

		dashClient := connectClient(t, env, user.ID, trendbirdv1connect.NewDashboardServiceClient)
		resp, err := dashClient.GetActivities(context.Background(), connect.NewRequest(&trendbirdv1.GetActivitiesRequest{}))
		if err != nil {
			t.Fatalf("GetActivities: %v", err)
		}

		// ACTIVITY_TYPE_POSTED (type=4) が存在すること
		found := false
		for _, a := range resp.Msg.GetActivities() {
			if a.GetType() == trendbirdv1.ActivityType_ACTIVITY_TYPE_POSTED {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected ACTIVITY_TYPE_POSTED activity, not found")
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.dashboardClient.GetActivities(context.Background(), connect.NewRequest(&trendbirdv1.GetActivitiesRequest{}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})

	t.Run("activity_topic_name_empty_for_non_topic_type", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		// Login タイプ（type=7）は topic_name が空文字
		seedActivity(t, env.db, user.ID,
			withActivityType(7),
			func(a *model.Activity) { a.TopicName = "" },
			withActivityDescription("ログインしました"),
		)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewDashboardServiceClient)
		resp, err := client.GetActivities(context.Background(), connect.NewRequest(&trendbirdv1.GetActivitiesRequest{}))
		if err != nil {
			t.Fatalf("GetActivities: %v", err)
		}

		activities := resp.Msg.GetActivities()
		if got := len(activities); got != 1 {
			t.Fatalf("activities count: want 1, got %d", got)
		}
		if activities[0].GetTopicName() != "" {
			t.Errorf("topic_name: want empty, got %q", activities[0].GetTopicName())
		}
	})

	t.Run("activities_exact_at_limit", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		// ちょうど20件 seed
		for i := range 20 {
			seedActivity(t, env.db, user.ID,
				withActivityDescription(fmt.Sprintf("activity %d", i)),
			)
			time.Sleep(10 * time.Millisecond)
		}

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewDashboardServiceClient)
		resp, err := client.GetActivities(context.Background(), connect.NewRequest(&trendbirdv1.GetActivitiesRequest{}))
		if err != nil {
			t.Fatalf("GetActivities: %v", err)
		}

		if got := len(resp.Msg.GetActivities()); got != 20 {
			t.Errorf("activities count: want 20, got %d", got)
		}
	})

	t.Run("timestamp_rfc3339_format", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		seedActivity(t, env.db, user.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewDashboardServiceClient)
		resp, err := client.GetActivities(context.Background(), connect.NewRequest(&trendbirdv1.GetActivitiesRequest{}))
		if err != nil {
			t.Fatalf("GetActivities: %v", err)
		}

		for i, a := range resp.Msg.GetActivities() {
			if _, parseErr := time.Parse(time.RFC3339, a.GetTimestamp()); parseErr != nil {
				t.Errorf("activities[%d].timestamp not valid RFC3339: %q, error: %v", i, a.GetTimestamp(), parseErr)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// TestDashboardService_GetStats
// ---------------------------------------------------------------------------

func TestDashboardService_GetStats(t *testing.T) {
	t.Run("success_with_data", func(t *testing.T) {
		env := setupTest(t)

		user := seedUser(t, env.db)
		topic1 := seedTopic(t, env.db)
		seedUserTopic(t, env.db, user.ID, topic1.ID)
		topic2 := seedTopic(t, env.db)
		seedUserTopic(t, env.db, user.ID, topic2.ID)

		// spike_histories x3 (topic1 に 2件、topic2 に 1件)
		seedSpikeHistory(t, env.db, topic1.ID)
		seedSpikeHistory(t, env.db, topic1.ID)
		seedSpikeHistory(t, env.db, topic2.ID)

		// ai_generation_logs x5（計画書に合わせて 5件）
		for i := 0; i < 5; i++ {
			seedAIGenerationLog(t, env.db, user.ID)
		}

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewDashboardServiceClient)
		resp, err := client.GetStats(context.Background(), connect.NewRequest(&trendbirdv1.GetStatsRequest{}))
		if err != nil {
			t.Fatalf("GetStats: %v", err)
		}

		want := &trendbirdv1.GetStatsResponse{
			Stats: &trendbirdv1.DashboardStats{
				Detections:  3,
				Generations: 5,
			},
		}
		if diff := cmp.Diff(want, resp.Msg, protocmp.Transform(), protocmp.IgnoreFields(&trendbirdv1.DashboardStats{}, "last_checked_at")); diff != "" {
			t.Errorf("GetStats mismatch (-want +got):\n%s", diff)
		}

		// last_checked_at が返却され、RFC3339 でパースできることを検証
		lastChecked := resp.Msg.GetStats().GetLastCheckedAt()
		if lastChecked == "" {
			t.Fatal("last_checked_at should not be empty when topics exist")
		}
		if _, parseErr := time.Parse(time.RFC3339, lastChecked); parseErr != nil {
			t.Errorf("last_checked_at is not valid RFC3339: %q, error: %v", lastChecked, parseErr)
		}
	})

	t.Run("cross_user_isolation", func(t *testing.T) {
		env := setupTest(t)

		userA := seedUser(t, env.db)
		userB := seedUser(t, env.db)
		topic1 := seedTopic(t, env.db)
		topic2 := seedTopic(t, env.db)
		seedUserTopic(t, env.db, userA.ID, topic1.ID)
		seedUserTopic(t, env.db, userB.ID, topic2.ID)

		// userA: topic1 に spike x2, aiGen x1
		seedSpikeHistory(t, env.db, topic1.ID)
		seedSpikeHistory(t, env.db, topic1.ID)
		seedAIGenerationLog(t, env.db, userA.ID)

		// userB: topic2 に spike x3, aiGen x4
		for i := 0; i < 3; i++ {
			seedSpikeHistory(t, env.db, topic2.ID)
		}
		for i := 0; i < 4; i++ {
			seedAIGenerationLog(t, env.db, userB.ID)
		}

		client := connectClient(t, env, userA.ID, trendbirdv1connect.NewDashboardServiceClient)
		resp, err := client.GetStats(context.Background(), connect.NewRequest(&trendbirdv1.GetStatsRequest{}))
		if err != nil {
			t.Fatalf("GetStats: %v", err)
		}

		stats := resp.Msg.GetStats()
		if stats.GetDetections() != 2 {
			t.Errorf("detections: want 2, got %d", stats.GetDetections())
		}
		if stats.GetGenerations() != 1 {
			t.Errorf("generations: want 1, got %d", stats.GetGenerations())
		}
	})

	t.Run("previous_month_excluded", func(t *testing.T) {
		env := setupTest(t)

		user := seedUser(t, env.db)
		topic := seedTopic(t, env.db)
		seedUserTopic(t, env.db, user.ID, topic.ID)

		// 当月: spike x2, aiGen x3
		sp1 := seedSpikeHistory(t, env.db, topic.ID)
		sp2 := seedSpikeHistory(t, env.db, topic.ID)
		ag1 := seedAIGenerationLog(t, env.db, user.ID)
		ag2 := seedAIGenerationLog(t, env.db, user.ID)
		ag3 := seedAIGenerationLog(t, env.db, user.ID)

		// 前月分も追加（後で created_at を変更）
		spOld := seedSpikeHistory(t, env.db, topic.ID)
		agOld1 := seedAIGenerationLog(t, env.db, user.ID)
		agOld2 := seedAIGenerationLog(t, env.db, user.ID)

		// 前月の日付
		now := time.Now()
		prevMonth := time.Date(now.Year(), now.Month()-1, 15, 12, 0, 0, 0, time.Local)

		// raw SQL で前月に変更
		env.db.Exec("UPDATE spike_histories SET created_at = ? WHERE id = ?", prevMonth, spOld.ID)
		env.db.Exec("UPDATE ai_generation_logs SET created_at = ? WHERE id = ?", prevMonth, agOld1.ID)
		env.db.Exec("UPDATE ai_generation_logs SET created_at = ? WHERE id = ?", prevMonth, agOld2.ID)

		// 当月分の ID を使って未使用警告を抑制
		_ = sp1
		_ = sp2
		_ = ag1
		_ = ag2
		_ = ag3

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewDashboardServiceClient)
		resp, err := client.GetStats(context.Background(), connect.NewRequest(&trendbirdv1.GetStatsRequest{}))
		if err != nil {
			t.Fatalf("GetStats: %v", err)
		}

		stats := resp.Msg.GetStats()
		if stats.GetDetections() != 2 {
			t.Errorf("detections: want 2 (current month only), got %d", stats.GetDetections())
		}
		if stats.GetGenerations() != 3 {
			t.Errorf("generations: want 3 (current month only), got %d", stats.GetGenerations())
		}
	})

	t.Run("multiple_topics_detections", func(t *testing.T) {
		env := setupTest(t)

		user := seedUser(t, env.db)
		topic1 := seedTopic(t, env.db)
		topic2 := seedTopic(t, env.db)
		topic3 := seedTopic(t, env.db)
		seedUserTopic(t, env.db, user.ID, topic1.ID)
		seedUserTopic(t, env.db, user.ID, topic2.ID)
		seedUserTopic(t, env.db, user.ID, topic3.ID)

		// topic1: 3件, topic2: 2件, topic3: 1件 → 合計6件
		for i := 0; i < 3; i++ {
			seedSpikeHistory(t, env.db, topic1.ID)
		}
		for i := 0; i < 2; i++ {
			seedSpikeHistory(t, env.db, topic2.ID)
		}
		seedSpikeHistory(t, env.db, topic3.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewDashboardServiceClient)
		resp, err := client.GetStats(context.Background(), connect.NewRequest(&trendbirdv1.GetStatsRequest{}))
		if err != nil {
			t.Fatalf("GetStats: %v", err)
		}

		if resp.Msg.GetStats().GetDetections() != 6 {
			t.Errorf("detections: want 6 (3+2+1), got %d", resp.Msg.GetStats().GetDetections())
		}
	})

	t.Run("last_checked_at_latest", func(t *testing.T) {
		env := setupTest(t)

		user := seedUser(t, env.db)
		topic1 := seedTopic(t, env.db)
		topic2 := seedTopic(t, env.db)
		topic3 := seedTopic(t, env.db)
		seedUserTopic(t, env.db, user.ID, topic1.ID)
		seedUserTopic(t, env.db, user.ID, topic2.ID)
		seedUserTopic(t, env.db, user.ID, topic3.ID)

		// 3トピックに異なる updated_at を設定
		now := time.Now()
		oldest := now.Add(-2 * time.Hour)
		middle := now.Add(-1 * time.Hour)
		latest := now

		env.db.Exec("UPDATE topics SET updated_at = ? WHERE id = ?", oldest, topic1.ID)
		env.db.Exec("UPDATE topics SET updated_at = ? WHERE id = ?", middle, topic2.ID)
		env.db.Exec("UPDATE topics SET updated_at = ? WHERE id = ?", latest, topic3.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewDashboardServiceClient)
		resp, err := client.GetStats(context.Background(), connect.NewRequest(&trendbirdv1.GetStatsRequest{}))
		if err != nil {
			t.Fatalf("GetStats: %v", err)
		}

		lastChecked := resp.Msg.GetStats().GetLastCheckedAt()
		if lastChecked == "" {
			t.Fatal("last_checked_at should not be empty")
		}
		parsed, parseErr := time.Parse(time.RFC3339, lastChecked)
		if parseErr != nil {
			t.Fatalf("last_checked_at parse error: %v", parseErr)
		}
		// latest (topic3) の updated_at と ±1秒以内であることを検証
		diff := parsed.Sub(latest)
		if diff < -1*time.Second || diff > 1*time.Second {
			t.Errorf("last_checked_at: want ~%v, got %v (diff=%v)", latest, parsed, diff)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.dashboardClient.GetStats(context.Background(), connect.NewRequest(&trendbirdv1.GetStatsRequest{}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})

	t.Run("zero_case", func(t *testing.T) {
		env := setupTest(t)

		user := seedUser(t, env.db)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewDashboardServiceClient)
		resp, err := client.GetStats(context.Background(), connect.NewRequest(&trendbirdv1.GetStatsRequest{}))
		if err != nil {
			t.Fatalf("GetStats: %v", err)
		}

		if resp.Msg.GetStats().GetDetections() != 0 {
			t.Errorf("detections: want 0, got %d", resp.Msg.GetStats().GetDetections())
		}
		if resp.Msg.GetStats().GetGenerations() != 0 {
			t.Errorf("generations: want 0, got %d", resp.Msg.GetStats().GetGenerations())
		}
		if got := resp.Msg.GetStats().GetLastCheckedAt(); got != "" {
			t.Errorf("last_checked_at: want empty, got %q", got)
		}
	})

	t.Run("month_boundary_exact", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		topic := seedTopic(t, env.db)
		seedUserTopic(t, env.db, user.ID, topic.ID)

		now := time.Now()
		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)

		// 月初 00:00:00 ちょうど → 当月に含まれるべき
		shIncluded := seedSpikeHistory(t, env.db, topic.ID)
		env.db.Exec("UPDATE spike_histories SET created_at = ? WHERE id = ?", monthStart, shIncluded.ID)

		// 前月末 23:59:59 → 除外されるべき
		shExcluded := seedSpikeHistory(t, env.db, topic.ID)
		env.db.Exec("UPDATE spike_histories SET created_at = ? WHERE id = ?", monthStart.Add(-1*time.Second), shExcluded.ID)

		agIncluded := seedAIGenerationLog(t, env.db, user.ID)
		env.db.Exec("UPDATE ai_generation_logs SET created_at = ? WHERE id = ?", monthStart, agIncluded.ID)

		agExcluded := seedAIGenerationLog(t, env.db, user.ID)
		env.db.Exec("UPDATE ai_generation_logs SET created_at = ? WHERE id = ?", monthStart.Add(-1*time.Second), agExcluded.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewDashboardServiceClient)
		resp, err := client.GetStats(context.Background(), connect.NewRequest(&trendbirdv1.GetStatsRequest{}))
		if err != nil {
			t.Fatalf("GetStats: %v", err)
		}

		stats := resp.Msg.GetStats()
		if stats.GetDetections() != 1 {
			t.Errorf("detections: want 1 (month boundary), got %d", stats.GetDetections())
		}
		if stats.GetGenerations() != 1 {
			t.Errorf("generations: want 1 (month boundary), got %d", stats.GetGenerations())
		}
	})

	t.Run("shared_topic_detections_for_both_users", func(t *testing.T) {
		env := setupTest(t)
		userA := seedUser(t, env.db)
		userB := seedUser(t, env.db)
		sharedTopic := seedTopic(t, env.db)
		seedUserTopic(t, env.db, userA.ID, sharedTopic.ID)
		seedUserTopic(t, env.db, userB.ID, sharedTopic.ID)

		// 共有トピックに spike 2件
		seedSpikeHistory(t, env.db, sharedTopic.ID)
		seedSpikeHistory(t, env.db, sharedTopic.ID)

		// userA の検知数
		clientA := connectClient(t, env, userA.ID, trendbirdv1connect.NewDashboardServiceClient)
		respA, err := clientA.GetStats(context.Background(), connect.NewRequest(&trendbirdv1.GetStatsRequest{}))
		if err != nil {
			t.Fatalf("GetStats userA: %v", err)
		}
		if respA.Msg.GetStats().GetDetections() != 2 {
			t.Errorf("userA detections: want 2, got %d", respA.Msg.GetStats().GetDetections())
		}

		// userB も同じ共有トピックの検知を見れる
		clientB := connectClient(t, env, userB.ID, trendbirdv1connect.NewDashboardServiceClient)
		respB, err := clientB.GetStats(context.Background(), connect.NewRequest(&trendbirdv1.GetStatsRequest{}))
		if err != nil {
			t.Fatalf("GetStats userB: %v", err)
		}
		if respB.Msg.GetStats().GetDetections() != 2 {
			t.Errorf("userB detections: want 2, got %d", respB.Msg.GetStats().GetDetections())
		}
	})

	t.Run("topics_exist_but_no_detections_or_generations", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		topic := seedTopic(t, env.db)
		seedUserTopic(t, env.db, user.ID, topic.ID)
		// spike_histories も ai_generation_logs も seed しない

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewDashboardServiceClient)
		resp, err := client.GetStats(context.Background(), connect.NewRequest(&trendbirdv1.GetStatsRequest{}))
		if err != nil {
			t.Fatalf("GetStats: %v", err)
		}

		stats := resp.Msg.GetStats()
		if stats.GetDetections() != 0 {
			t.Errorf("detections: want 0, got %d", stats.GetDetections())
		}
		if stats.GetGenerations() != 0 {
			t.Errorf("generations: want 0, got %d", stats.GetGenerations())
		}
		// トピックは存在するので last_checked_at は設定されるべき
		if stats.GetLastCheckedAt() == "" {
			t.Error("last_checked_at: want non-empty (topics exist), got empty")
		}
	})
}

