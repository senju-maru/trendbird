package e2etest

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"connectrpc.com/connect"

	trendbirdv1 "github.com/trendbird/backend/gen/trendbird/v1"
	"github.com/trendbird/backend/gen/trendbird/v1/trendbirdv1connect"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

func TestTopicService_ListTopics(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		t1 := seedTopic(t, env.db, withTopicName("Topic A"), withTopicGenre("technology"))
		seedUserTopic(t, env.db, user.ID, t1.ID)
		t2 := seedTopic(t, env.db, withTopicName("Topic B"), withTopicGenre("marketing"))
		seedUserTopic(t, env.db, user.ID, t2.ID)
		t3 := seedTopic(t, env.db, withTopicName("Topic C"), withTopicGenre("business"))
		seedUserTopic(t, env.db, user.ID, t3.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		resp, err := client.ListTopics(context.Background(), connect.NewRequest(&trendbirdv1.ListTopicsRequest{}))
		if err != nil {
			t.Fatalf("ListTopics: %v", err)
		}

		topics := resp.Msg.GetTopics()
		if got := len(topics); got != 3 {
			t.Fatalf("expected 3 topics, got %d", got)
		}

		wantMap := map[string]*model.Topic{
			t1.ID: t1,
			t2.ID: t2,
			t3.ID: t3,
		}
		wantSlugs := map[string]string{t1.ID: "technology", t2.ID: "marketing", t3.ID: "business"}
		for _, got := range topics {
			want, ok := wantMap[got.GetId()]
			if !ok {
				t.Errorf("unexpected topic ID: %s", got.GetId())
				continue
			}
			if got.GetName() != want.Name {
				t.Errorf("name: want %s, got %s", want.Name, got.GetName())
			}
			if got.GetGenre() != wantSlugs[got.GetId()] {
				t.Errorf("genre: want %s, got %s", wantSlugs[got.GetId()], got.GetGenre())
			}
			if int32(got.GetStatus()) != want.Status {
				t.Errorf("status: want %d, got %d", want.Status, got.GetStatus())
			}
		}

	})

	t.Run("cross_user_isolation", func(t *testing.T) {
		env := setupTest(t)
		userA := seedUser(t, env.db)
		userB := seedUser(t, env.db)

		tpA := seedTopic(t, env.db, withTopicName("Topic for A"))
		seedUserTopic(t, env.db, userA.ID, tpA.ID)
		tpB := seedTopic(t, env.db, withTopicName("Topic for B"))
		seedUserTopic(t, env.db, userB.ID, tpB.ID)

		clientA := connectClient(t, env, userA.ID, trendbirdv1connect.NewTopicServiceClient)
		resp, err := clientA.ListTopics(context.Background(), connect.NewRequest(&trendbirdv1.ListTopicsRequest{}))
		if err != nil {
			t.Fatalf("ListTopics: %v", err)
		}

		topics := resp.Msg.GetTopics()
		if got := len(topics); got != 1 {
			t.Fatalf("expected 1 topic for userA, got %d", got)
		}
		if topics[0].GetName() != "Topic for A" {
			t.Errorf("name: want %q, got %q", "Topic for A", topics[0].GetName())
		}
	})

	t.Run("empty", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		resp, err := client.ListTopics(context.Background(), connect.NewRequest(&trendbirdv1.ListTopicsRequest{}))
		if err != nil {
			t.Fatalf("ListTopics: %v", err)
		}

		if got := len(resp.Msg.GetTopics()); got != 0 {
			t.Errorf("expected 0 topics, got %d", got)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.topicClient.ListTopics(context.Background(), connect.NewRequest(&trendbirdv1.ListTopicsRequest{}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

func TestTopicService_GetTopic(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		tp := seedTopic(t, env.db,
			withTopicName("My Topic"),
			withTopicGenre("marketing"),
			withTopicKeywords([]string{"Go", "Rust"}),
		)
		seedUserTopic(t, env.db, user.ID, tp.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		resp, err := client.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{
			Id: tp.ID,
		}))
		if err != nil {
			t.Fatalf("GetTopic: %v", err)
		}

		got := resp.Msg.GetTopic()
		if got.GetId() != tp.ID {
			t.Errorf("id: want %s, got %s", tp.ID, got.GetId())
		}
		if got.GetName() != tp.Name {
			t.Errorf("name: want %s, got %s", tp.Name, got.GetName())
		}
		if got.GetGenre() != "marketing" {
			t.Errorf("genre: want marketing, got %s", got.GetGenre())
		}
		if int32(got.GetStatus()) != tp.Status {
			t.Errorf("status: want %d, got %d", tp.Status, got.GetStatus())
		}
		if got.GetNotificationEnabled() != true {
			t.Error("expected notification_enabled=true")
		}
		// keywords フィールド検証 (B2)
		keywords := got.GetKeywords()
		if len(keywords) != 2 || keywords[0] != "Go" || keywords[1] != "Rust" {
			t.Errorf("keywords: want [Go Rust], got %v", keywords)
		}
		// change_percent フィールド検証 (B2) — seed デフォルトは 0
		if got.GetChangePercent() != tp.ChangePercent {
			t.Errorf("change_percent: want %f, got %f", tp.ChangePercent, got.GetChangePercent())
		}
	})

	t.Run("not_found", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		_, err := client.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{
			Id: "00000000-0000-0000-0000-000000000000",
		}))
		assertConnectCode(t, err, connect.CodeNotFound)
	})

	t.Run("permission_denied", func(t *testing.T) {
		env := setupTest(t)
		owner := seedUser(t, env.db)
		other := seedUser(t, env.db)
		tp := seedTopic(t, env.db)
		seedUserTopic(t, env.db, owner.ID, tp.ID)
		// other has no user_topic link

		client := connectClient(t, env, other.ID, trendbirdv1connect.NewTopicServiceClient)
		_, err := client.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{
			Id: tp.ID,
		}))
		// No user_topic link → NotFound (acts as permission denied)
		assertConnectCode(t, err, connect.CodeNotFound)
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.topicClient.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{
			Id: "some-id",
		}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})

	// --- ステータス別テスト ---

	t.Run("spike_status_with_all_fields", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		now := time.Now()
		spikeStarted := now.Add(-30 * time.Minute)

		tp := seedTopic(t, env.db,
			withTopicName("Spike Topic"),
			withTopicGenre("technology"),
			withTopicKeywords([]string{"spike", "test"}),
			withTopicStatus(1), // Spike
			withTopicZScore(5.2),
			withTopicCurrentVolume(1500),
			withTopicBaselineVolume(200),
			withTopicChangePercent(650.0),
			withTopicContext("急増の背景"),
			withTopicContextSummary("AI関連のニュースが拡散中"),
			withTopicSpikeStartedAt(spikeStarted),
		)
		seedUserTopic(t, env.db, user.ID, tp.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		resp, err := client.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{
			Id: tp.ID,
		}))
		if err != nil {
			t.Fatalf("GetTopic: %v", err)
		}

		got := resp.Msg.GetTopic()
		if got.GetStatus() != trendbirdv1.TopicStatus_TOPIC_STATUS_SPIKE {
			t.Errorf("status: want SPIKE, got %v", got.GetStatus())
		}
		if got.ZScore == nil {
			t.Fatal("z_score should not be nil")
		}
		if math.Abs(got.GetZScore()-5.2) > 0.01 {
			t.Errorf("z_score: want 5.2, got %f", got.GetZScore())
		}
		if got.GetCurrentVolume() != 1500 {
			t.Errorf("current_volume: want 1500, got %d", got.GetCurrentVolume())
		}
		if got.GetBaselineVolume() != 200 {
			t.Errorf("baseline_volume: want 200, got %d", got.GetBaselineVolume())
		}
		if math.Abs(got.GetChangePercent()-650.0) > 0.01 {
			t.Errorf("change_percent: want 650.0, got %f", got.GetChangePercent())
		}
		if got.Context == nil || got.GetContext() != "急増の背景" {
			t.Errorf("context: want %q, got %v", "急増の背景", got.Context)
		}
		if got.ContextSummary == nil || got.GetContextSummary() != "AI関連のニュースが拡散中" {
			t.Errorf("context_summary: want %q, got %v", "AI関連のニュースが拡散中", got.ContextSummary)
		}
		if got.SpikeStartedAt == nil {
			t.Fatal("spike_started_at should not be nil for Spike status")
		}
		parsedSpike, err := time.Parse(time.RFC3339, got.GetSpikeStartedAt())
		if err != nil {
			t.Fatalf("spike_started_at parse: %v", err)
		}
		if parsedSpike.Sub(spikeStarted).Abs() > 2*time.Second {
			t.Errorf("spike_started_at: want ~%v, got %v", spikeStarted, parsedSpike)
		}
	})

	t.Run("rising_status", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		tp := seedTopic(t, env.db,
			withTopicStatus(2), // Rising
			withTopicZScore(2.5),
		)
		seedUserTopic(t, env.db, user.ID, tp.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		resp, err := client.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{
			Id: tp.ID,
		}))
		if err != nil {
			t.Fatalf("GetTopic: %v", err)
		}

		got := resp.Msg.GetTopic()
		if got.GetStatus() != trendbirdv1.TopicStatus_TOPIC_STATUS_RISING {
			t.Errorf("status: want RISING, got %v", got.GetStatus())
		}
		// Rising ではspikeStartedAtはnilのはず
		if got.SpikeStartedAt != nil {
			t.Errorf("spike_started_at: want nil for Rising status, got %v", got.GetSpikeStartedAt())
		}
	})

	t.Run("stable_status_no_optional_fields", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		tp := seedTopic(t, env.db,
			withTopicStatus(3), // Stable (default)
		)
		seedUserTopic(t, env.db, user.ID, tp.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		resp, err := client.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{
			Id: tp.ID,
		}))
		if err != nil {
			t.Fatalf("GetTopic: %v", err)
		}

		got := resp.Msg.GetTopic()
		if got.GetStatus() != trendbirdv1.TopicStatus_TOPIC_STATUS_STABLE {
			t.Errorf("status: want STABLE, got %v", got.GetStatus())
		}
		if got.ZScore != nil {
			t.Errorf("z_score: want nil for Stable, got %v", got.GetZScore())
		}
		if got.Context != nil {
			t.Errorf("context: want nil, got %v", got.GetContext())
		}
		if got.ContextSummary != nil {
			t.Errorf("context_summary: want nil, got %v", got.GetContextSummary())
		}
		if got.SpikeStartedAt != nil {
			t.Errorf("spike_started_at: want nil, got %v", got.GetSpikeStartedAt())
		}
	})

	// --- Sparkline テスト ---

	t.Run("weekly_sparkline_data_populated", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		tp := seedTopic(t, env.db)
		seedUserTopic(t, env.db, user.ID, tp.ID)

		now := time.Now()
		// 7件の volume データを seed（7日間以内）
		for i := 6; i >= 0; i-- {
			seedTopicVolume(t, env.db, tp.ID,
				withTopicVolumeTimestamp(now.Add(-time.Duration(i)*24*time.Hour)),
				withTopicVolumeValue(int32(100+i*10)),
			)
		}

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		resp, err := client.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{
			Id: tp.ID,
		}))
		if err != nil {
			t.Fatalf("GetTopic: %v", err)
		}

		sparkline := resp.Msg.GetTopic().GetWeeklySparklineData()
		if got := len(sparkline); got != 7 {
			t.Fatalf("weekly_sparkline_data count: want 7, got %d", got)
		}
		// ASC 順（古い→新しい）を検証
		for i := 0; i < len(sparkline)-1; i++ {
			cur, _ := time.Parse(time.RFC3339, sparkline[i].GetTimestamp())
			next, _ := time.Parse(time.RFC3339, sparkline[i+1].GetTimestamp())
			if !cur.Before(next) {
				t.Errorf("sparkline not ASC: [%d]=%v >= [%d]=%v", i, cur, i+1, next)
			}
		}
	})

	t.Run("weekly_sparkline_data_empty", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		tp := seedTopic(t, env.db)
		seedUserTopic(t, env.db, user.ID, tp.ID)
		// volume データを seed しない

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		resp, err := client.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{
			Id: tp.ID,
		}))
		if err != nil {
			t.Fatalf("GetTopic: %v", err)
		}

		sparkline := resp.Msg.GetTopic().GetWeeklySparklineData()
		if len(sparkline) != 0 {
			t.Errorf("weekly_sparkline_data count: want 0, got %d", len(sparkline))
		}
	})

	t.Run("weekly_sparkline_data_excludes_old_data", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		tp := seedTopic(t, env.db)
		seedUserTopic(t, env.db, user.ID, tp.ID)

		now := time.Now()
		// 7日窓内に2件
		seedTopicVolume(t, env.db, tp.ID,
			withTopicVolumeTimestamp(now.Add(-1*24*time.Hour)),
			withTopicVolumeValue(100),
		)
		seedTopicVolume(t, env.db, tp.ID,
			withTopicVolumeTimestamp(now.Add(-2*24*time.Hour)),
			withTopicVolumeValue(200),
		)
		// 7日窓外に1件
		seedTopicVolume(t, env.db, tp.ID,
			withTopicVolumeTimestamp(now.Add(-8*24*time.Hour)),
			withTopicVolumeValue(999),
		)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		resp, err := client.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{
			Id: tp.ID,
		}))
		if err != nil {
			t.Fatalf("GetTopic: %v", err)
		}

		sparkline := resp.Msg.GetTopic().GetWeeklySparklineData()
		if got := len(sparkline); got != 2 {
			t.Errorf("weekly_sparkline_data count: want 2 (7d window), got %d", got)
		}
	})

	t.Run("weekly_sparkline_data_high_density", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		tp := seedTopic(t, env.db)
		seedUserTopic(t, env.db, user.ID, tp.ID)

		now := time.Now()
		// 50件のデータを1時間間隔で seed（全て7日以内）
		for i := 0; i < 50; i++ {
			seedTopicVolume(t, env.db, tp.ID,
				withTopicVolumeTimestamp(now.Add(-time.Duration(i)*time.Hour)),
				withTopicVolumeValue(int32(i+1)),
			)
		}

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		resp, err := client.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{
			Id: tp.ID,
		}))
		if err != nil {
			t.Fatalf("GetTopic: %v", err)
		}

		sparkline := resp.Msg.GetTopic().GetWeeklySparklineData()
		if got := len(sparkline); got != 50 {
			t.Errorf("weekly_sparkline_data count: want 50, got %d", got)
		}
	})

	// --- SpikeHistory テスト ---

	t.Run("spike_history_populated", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		tp := seedTopic(t, env.db)
		seedUserTopic(t, env.db, user.ID, tp.ID)

		now := time.Now()
		sh1 := seedSpikeHistory(t, env.db, tp.ID,
			withSpikeTimestamp(now.Add(-3*time.Hour)),
			withSpikePeakZScore(3.0),
			withSpikeSummary("First spike"),
			withSpikeDurationMinutes(15),
		)
		sh2 := seedSpikeHistory(t, env.db, tp.ID,
			withSpikeTimestamp(now.Add(-2*time.Hour)),
			withSpikePeakZScore(4.5),
			withSpikeSummary("Second spike"),
			withSpikeDurationMinutes(45),
		)
		sh3 := seedSpikeHistory(t, env.db, tp.ID,
			withSpikeTimestamp(now.Add(-1*time.Hour)),
			withSpikePeakZScore(6.0),
			withSpikeSummary("Third spike"),
			withSpikeDurationMinutes(60),
		)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		resp, err := client.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{
			Id: tp.ID,
		}))
		if err != nil {
			t.Fatalf("GetTopic: %v", err)
		}

		history := resp.Msg.GetTopic().GetSpikeHistory()
		if got := len(history); got != 3 {
			t.Fatalf("spike_history count: want 3, got %d", got)
		}

		// DESC 順を検証（最新が先頭）
		for i := 0; i < len(history)-1; i++ {
			cur, _ := time.Parse(time.RFC3339, history[i].GetTimestamp())
			next, _ := time.Parse(time.RFC3339, history[i+1].GetTimestamp())
			if cur.Before(next) {
				t.Errorf("spike_history not DESC: [%d]=%v < [%d]=%v", i, cur, i+1, next)
			}
		}

		// 全フィールド検証（先頭=最新=sh3）
		first := history[0]
		if first.GetId() != sh3.ID {
			t.Errorf("spike_history[0].id: want %s, got %s", sh3.ID, first.GetId())
		}
		if math.Abs(first.GetPeakZScore()-6.0) > 0.01 {
			t.Errorf("spike_history[0].peak_z_score: want 6.0, got %f", first.GetPeakZScore())
		}
		if first.GetSummary() != "Third spike" {
			t.Errorf("spike_history[0].summary: want %q, got %q", "Third spike", first.GetSummary())
		}
		if first.GetDurationMinutes() != 60 {
			t.Errorf("spike_history[0].duration_minutes: want 60, got %d", first.GetDurationMinutes())
		}

		_ = sh1
		_ = sh2
	})

	t.Run("spike_history_empty", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		tp := seedTopic(t, env.db)
		seedUserTopic(t, env.db, user.ID, tp.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		resp, err := client.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{
			Id: tp.ID,
		}))
		if err != nil {
			t.Fatalf("GetTopic: %v", err)
		}

		history := resp.Msg.GetTopic().GetSpikeHistory()
		if len(history) != 0 {
			t.Errorf("spike_history count: want 0, got %d", len(history))
		}
	})

	t.Run("spike_history_mixed_statuses", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		tp := seedTopic(t, env.db)
		seedUserTopic(t, env.db, user.ID, tp.ID)

		seedSpikeHistory(t, env.db, tp.ID,
			withSpikeStatus(1), // Spike
			withSpikePeakZScore(4.0),
		)
		seedSpikeHistory(t, env.db, tp.ID,
			withSpikeStatus(2), // Rising
			withSpikePeakZScore(2.5),
		)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		resp, err := client.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{
			Id: tp.ID,
		}))
		if err != nil {
			t.Fatalf("GetTopic: %v", err)
		}

		history := resp.Msg.GetTopic().GetSpikeHistory()
		if got := len(history); got != 2 {
			t.Fatalf("spike_history count: want 2, got %d", got)
		}

		statusSet := map[trendbirdv1.TopicStatus]bool{}
		for _, h := range history {
			statusSet[h.GetStatus()] = true
		}
		if !statusSet[trendbirdv1.TopicStatus_TOPIC_STATUS_SPIKE] {
			t.Error("expected SPIKE status in spike_history")
		}
		if !statusSet[trendbirdv1.TopicStatus_TOPIC_STATUS_RISING] {
			t.Error("expected RISING status in spike_history")
		}
	})

	t.Run("spike_history_cross_topic_isolation", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		topicA := seedTopic(t, env.db)
		topicB := seedTopic(t, env.db)
		seedUserTopic(t, env.db, user.ID, topicA.ID)
		seedUserTopic(t, env.db, user.ID, topicB.ID)

		// topicA に2件
		seedSpikeHistory(t, env.db, topicA.ID)
		seedSpikeHistory(t, env.db, topicA.ID)
		// topicB に3件
		seedSpikeHistory(t, env.db, topicB.ID)
		seedSpikeHistory(t, env.db, topicB.ID)
		seedSpikeHistory(t, env.db, topicB.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)

		respA, err := client.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{
			Id: topicA.ID,
		}))
		if err != nil {
			t.Fatalf("GetTopic topicA: %v", err)
		}
		if got := len(respA.Msg.GetTopic().GetSpikeHistory()); got != 2 {
			t.Errorf("topicA spike_history: want 2, got %d", got)
		}

		respB, err := client.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{
			Id: topicB.ID,
		}))
		if err != nil {
			t.Fatalf("GetTopic topicB: %v", err)
		}
		if got := len(respB.Msg.GetTopic().GetSpikeHistory()); got != 3 {
			t.Errorf("topicB spike_history: want 3, got %d", got)
		}
	})

	// --- フィールド検証テスト ---

	t.Run("all_proto_fields_completeness", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		now := time.Now()
		spikeStarted := now.Add(-10 * time.Minute)

		tp := seedTopic(t, env.db,
			withTopicName("Full Fields Topic"),
			withTopicGenre("marketing"),
			withTopicKeywords([]string{"kw1", "kw2", "kw3"}),
			withTopicStatus(1), // Spike
			withTopicZScore(4.0),
			withTopicCurrentVolume(500),
			withTopicBaselineVolume(100),
			withTopicChangePercent(400.0),
			withTopicContext("Full context"),
			withTopicContextSummary("Full summary"),
			withTopicSpikeStartedAt(spikeStarted),
		)
		seedUserTopic(t, env.db, user.ID, tp.ID, withNotificationEnabled(true))

		// weekly sparkline 1件
		seedTopicVolume(t, env.db, tp.ID,
			withTopicVolumeTimestamp(now.Add(-1*time.Hour)),
			withTopicVolumeValue(42),
		)
		// spike history 1件
		seedSpikeHistory(t, env.db, tp.ID,
			withSpikePeakZScore(3.5),
			withSpikeSummary("Test spike"),
		)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		resp, err := client.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{
			Id: tp.ID,
		}))
		if err != nil {
			t.Fatalf("GetTopic: %v", err)
		}

		got := resp.Msg.GetTopic()

		// 全19フィールドを検証
		if got.GetId() != tp.ID {
			t.Errorf("id: want %s, got %s", tp.ID, got.GetId())
		}
		if got.GetName() != "Full Fields Topic" {
			t.Errorf("name: want %q, got %q", "Full Fields Topic", got.GetName())
		}
		kw := got.GetKeywords()
		if len(kw) != 3 || kw[0] != "kw1" || kw[1] != "kw2" || kw[2] != "kw3" {
			t.Errorf("keywords: want [kw1 kw2 kw3], got %v", kw)
		}
		if got.GetGenre() != "marketing" {
			t.Errorf("genre: want marketing, got %s", got.GetGenre())
		}
		if got.GetStatus() != trendbirdv1.TopicStatus_TOPIC_STATUS_SPIKE {
			t.Errorf("status: want SPIKE, got %v", got.GetStatus())
		}
		if math.Abs(got.GetChangePercent()-400.0) > 0.01 {
			t.Errorf("change_percent: want 400.0, got %f", got.GetChangePercent())
		}
		if got.ZScore == nil || math.Abs(got.GetZScore()-4.0) > 0.01 {
			t.Errorf("z_score: want 4.0, got %v", got.ZScore)
		}
		if got.GetCurrentVolume() != 500 {
			t.Errorf("current_volume: want 500, got %d", got.GetCurrentVolume())
		}
		if got.GetBaselineVolume() != 100 {
			t.Errorf("baseline_volume: want 100, got %d", got.GetBaselineVolume())
		}
		if got.Context == nil || got.GetContext() != "Full context" {
			t.Errorf("context: want %q, got %v", "Full context", got.Context)
		}
		if _, err := time.Parse(time.RFC3339, got.GetCreatedAt()); err != nil {
			t.Errorf("created_at parse: %v", err)
		}
		if got.ContextSummary == nil || got.GetContextSummary() != "Full summary" {
			t.Errorf("context_summary: want %q, got %v", "Full summary", got.ContextSummary)
		}
		if got.SpikeStartedAt == nil {
			t.Error("spike_started_at should not be nil")
		}
		if got := len(got.GetWeeklySparklineData()); got != 1 {
			t.Errorf("weekly_sparkline_data count: want 1, got %d", got)
		}
		if got := len(resp.Msg.GetTopic().GetSpikeHistory()); got != 1 {
			t.Errorf("spike_history count: want 1, got %d", got)
		}
		if resp.Msg.GetTopic().GetNotificationEnabled() != true {
			t.Error("notification_enabled: want true")
		}
	})

	t.Run("notification_disabled", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		tp := seedTopic(t, env.db)
		ut := seedUserTopic(t, env.db, user.ID, tp.ID)
		// GORM Create は false をゼロ値としてスキップするため raw SQL で更新
		env.db.Exec("UPDATE user_topics SET notification_enabled = false WHERE id = ?", ut.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		resp, err := client.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{
			Id: tp.ID,
		}))
		if err != nil {
			t.Fatalf("GetTopic: %v", err)
		}

		if resp.Msg.GetTopic().GetNotificationEnabled() {
			t.Error("notification_enabled: want false, got true")
		}
	})

	t.Run("z_score_boundary_values", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		// z-score = 3.0 (Spike 閾値)
		tp1 := seedTopic(t, env.db, withTopicZScore(3.0))
		seedUserTopic(t, env.db, user.ID, tp1.ID)

		// z-score = 2.0 (Rising 閾値)
		tp2 := seedTopic(t, env.db, withTopicZScore(2.0))
		seedUserTopic(t, env.db, user.ID, tp2.ID)

		// z-score = 0.0
		tp3 := seedTopic(t, env.db, withTopicZScore(0.0))
		seedUserTopic(t, env.db, user.ID, tp3.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)

		resp1, err := client.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{Id: tp1.ID}))
		if err != nil {
			t.Fatalf("GetTopic tp1: %v", err)
		}
		if resp1.Msg.GetTopic().ZScore == nil || math.Abs(resp1.Msg.GetTopic().GetZScore()-3.0) > 0.01 {
			t.Errorf("tp1 z_score: want 3.0, got %v", resp1.Msg.GetTopic().ZScore)
		}

		resp2, err := client.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{Id: tp2.ID}))
		if err != nil {
			t.Fatalf("GetTopic tp2: %v", err)
		}
		if resp2.Msg.GetTopic().ZScore == nil || math.Abs(resp2.Msg.GetTopic().GetZScore()-2.0) > 0.01 {
			t.Errorf("tp2 z_score: want 2.0, got %v", resp2.Msg.GetTopic().ZScore)
		}

		resp3, err := client.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{Id: tp3.ID}))
		if err != nil {
			t.Fatalf("GetTopic tp3: %v", err)
		}
		if resp3.Msg.GetTopic().ZScore == nil || math.Abs(resp3.Msg.GetTopic().GetZScore()-0.0) > 0.01 {
			t.Errorf("tp3 z_score: want 0.0, got %v", resp3.Msg.GetTopic().ZScore)
		}
	})

	t.Run("large_volumes", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		tp := seedTopic(t, env.db,
			withTopicCurrentVolume(math.MaxInt32),
			withTopicChangePercent(9999.99),
		)
		seedUserTopic(t, env.db, user.ID, tp.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		resp, err := client.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{
			Id: tp.ID,
		}))
		if err != nil {
			t.Fatalf("GetTopic: %v", err)
		}

		got := resp.Msg.GetTopic()
		if got.GetCurrentVolume() != math.MaxInt32 {
			t.Errorf("current_volume: want %d, got %d", int32(math.MaxInt32), got.GetCurrentVolume())
		}
		if math.Abs(got.GetChangePercent()-9999.99) > 0.01 {
			t.Errorf("change_percent: want 9999.99, got %f", got.GetChangePercent())
		}
	})

	t.Run("single_keyword", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		tp := seedTopic(t, env.db,
			withTopicKeywords([]string{"onlyone"}),
		)
		seedUserTopic(t, env.db, user.ID, tp.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		resp, err := client.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{
			Id: tp.ID,
		}))
		if err != nil {
			t.Fatalf("GetTopic: %v", err)
		}

		kw := resp.Msg.GetTopic().GetKeywords()
		if len(kw) != 1 || kw[0] != "onlyone" {
			t.Errorf("keywords: want [onlyone], got %v", kw)
		}
	})

	// --- エラーテスト ---

	t.Run("empty_topic_id", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		_, err := client.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{
			Id: "",
		}))
		// 空文字は PostgreSQL で invalid UUID エラーになる（buf/validate は Connect ミドルウェア未設定）
		assertConnectCode(t, err, connect.CodeInternal)
	})

	t.Run("invalid_uuid_topic_id", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		_, err := client.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{
			Id: "not-a-valid-uuid",
		}))
		// 不正な UUID は PostgreSQL で invalid input syntax エラーになる
		assertConnectCode(t, err, connect.CodeInternal)
	})

	// --- 共有リソーステスト ---

	t.Run("shared_topic_different_notification", func(t *testing.T) {
		env := setupTest(t)
		userA := seedUser(t, env.db)
		userB := seedUser(t, env.db)
		tp := seedTopic(t, env.db)
		seedUserTopic(t, env.db, userA.ID, tp.ID)
		utB := seedUserTopic(t, env.db, userB.ID, tp.ID)
		// GORM Create は false をゼロ値としてスキップするため raw SQL で更新
		env.db.Exec("UPDATE user_topics SET notification_enabled = false WHERE id = ?", utB.ID)

		clientA := connectClient(t, env, userA.ID, trendbirdv1connect.NewTopicServiceClient)
		respA, err := clientA.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{
			Id: tp.ID,
		}))
		if err != nil {
			t.Fatalf("GetTopic userA: %v", err)
		}
		if !respA.Msg.GetTopic().GetNotificationEnabled() {
			t.Error("userA notification_enabled: want true, got false")
		}

		clientB := connectClient(t, env, userB.ID, trendbirdv1connect.NewTopicServiceClient)
		respB, err := clientB.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{
			Id: tp.ID,
		}))
		if err != nil {
			t.Fatalf("GetTopic userB: %v", err)
		}
		if respB.Msg.GetTopic().GetNotificationEnabled() {
			t.Error("userB notification_enabled: want false, got true")
		}
	})
}

func TestTopicService_CreateTopic(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		ensureGenre(t, env.db, "technology")

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		resp, err := client.CreateTopic(context.Background(), connect.NewRequest(&trendbirdv1.CreateTopicRequest{
			Name:     "AI News",
			Keywords: []string{"openai", "chatgpt"},
			Genre:    "technology",
		}))
		if err != nil {
			t.Fatalf("CreateTopic: %v", err)
		}

		got := resp.Msg.GetTopic()
		if got.GetName() != "AI News" {
			t.Errorf("name: want AI News, got %s", got.GetName())
		}
		if got.GetGenre() != "technology" {
			t.Errorf("genre: want technology, got %s", got.GetGenre())
		}

		// DB 上にトピックが作成されていることを検証
		var dbTopic model.Topic
		if err := env.db.First(&dbTopic, "id = ?", got.GetId()).Error; err != nil {
			t.Fatalf("failed to fetch topic from DB: %v", err)
		}
		if dbTopic.Name != "AI News" {
			t.Errorf("DB name: want AI News, got %s", dbTopic.Name)
		}

		// DB 上に user_topics レコードが作成されていることを検証
		var dbUserTopic model.UserTopic
		if err := env.db.Where("user_id = ? AND topic_id = ?", user.ID, got.GetId()).First(&dbUserTopic).Error; err != nil {
			t.Fatalf("failed to fetch user_topic from DB: %v", err)
		}
		if !dbUserTopic.NotificationEnabled {
			t.Error("expected notification_enabled=true for new user_topic")
		}

		// activity が記録されていることを検証
		var activity model.Activity
		if err := env.db.Where("user_id = ? AND type = ?", user.ID, 5).First(&activity).Error; err != nil {
			t.Fatalf("failed to fetch activity: %v", err)
		}
		if activity.TopicName != "AI News" {
			t.Errorf("activity topic_name: want AI News, got %s", activity.TopicName)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.topicClient.CreateTopic(context.Background(), connect.NewRequest(&trendbirdv1.CreateTopicRequest{
			Name:     "Test",
			Keywords: []string{"test"},
			Genre:    "technology",
		}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})

	t.Run("nonexistent_genre", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		_, err := client.CreateTopic(context.Background(), connect.NewRequest(&trendbirdv1.CreateTopicRequest{
			Name:     "Some Topic",
			Keywords: []string{"test"},
			Genre:    "nonexistent-genre-slug",
		}))
		assertConnectCode(t, err, connect.CodeNotFound)
	})
}

func TestTopicService_CreateTopic_SharedTopicReuse(t *testing.T) {
	env := setupTest(t)
	ensureGenre(t, env.db, "technology")
	userA := seedUser(t, env.db)
	userB := seedUser(t, env.db)

	clientA := connectClient(t, env, userA.ID, trendbirdv1connect.NewTopicServiceClient)
	respA, err := clientA.CreateTopic(context.Background(), connect.NewRequest(&trendbirdv1.CreateTopicRequest{
		Name:     "Shared Topic",
		Keywords: []string{"shared"},
		Genre:    "technology",
	}))
	if err != nil {
		t.Fatalf("CreateTopic userA: %v", err)
	}
	topicIDA := respA.Msg.GetTopic().GetId()

	// userB creates the same topic (name+genre) → should reuse the same topics record
	clientB := connectClient(t, env, userB.ID, trendbirdv1connect.NewTopicServiceClient)
	respB, err := clientB.CreateTopic(context.Background(), connect.NewRequest(&trendbirdv1.CreateTopicRequest{
		Name:     "Shared Topic",
		Keywords: []string{"shared"},
		Genre:    "technology",
	}))
	if err != nil {
		t.Fatalf("CreateTopic userB: %v", err)
	}
	topicIDB := respB.Msg.GetTopic().GetId()

	if topicIDA != topicIDB {
		t.Errorf("expected same topic ID, got A=%s B=%s", topicIDA, topicIDB)
	}

	// Verify 2 user_topics records exist for the same topic
	var utCount int64
	env.db.Model(&model.UserTopic{}).Where("topic_id = ?", topicIDA).Count(&utCount)
	if utCount != 2 {
		t.Errorf("expected 2 user_topics, got %d", utCount)
	}
}

func TestTopicService_DeleteTopic(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		tp := seedTopic(t, env.db, withTopicName("To Delete"))
		seedUserTopic(t, env.db, user.ID, tp.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		_, err := client.DeleteTopic(context.Background(), connect.NewRequest(&trendbirdv1.DeleteTopicRequest{
			Id: tp.ID,
		}))
		if err != nil {
			t.Fatalf("DeleteTopic: %v", err)
		}

		// user_topics リンクが削除されていることを検証
		var utCount int64
		env.db.Model(&model.UserTopic{}).Where("user_id = ? AND topic_id = ?", user.ID, tp.ID).Count(&utCount)
		if utCount != 0 {
			t.Errorf("expected user_topic to be deleted, got count=%d", utCount)
		}

		// 共有トピック自体は残っていることを検証
		var topicCount int64
		env.db.Model(&model.Topic{}).Where("id = ?", tp.ID).Count(&topicCount)
		if topicCount != 1 {
			t.Errorf("expected topic to remain, got count=%d", topicCount)
		}

		// activity が記録されていることを検証
		var activity model.Activity
		if err := env.db.Where("user_id = ? AND type = ?", user.ID, 6).First(&activity).Error; err != nil {
			t.Fatalf("failed to fetch activity: %v", err)
		}
		if activity.TopicName != "To Delete" {
			t.Errorf("activity topic_name: want To Delete, got %s", activity.TopicName)
		}
	})

	t.Run("not_found", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		_, err := client.DeleteTopic(context.Background(), connect.NewRequest(&trendbirdv1.DeleteTopicRequest{
			Id: "00000000-0000-0000-0000-000000000000",
		}))
		assertConnectCode(t, err, connect.CodeNotFound)
	})

	t.Run("permission_denied", func(t *testing.T) {
		env := setupTest(t)
		owner := seedUser(t, env.db)
		other := seedUser(t, env.db)
		tp := seedTopic(t, env.db)
		seedUserTopic(t, env.db, owner.ID, tp.ID)

		client := connectClient(t, env, other.ID, trendbirdv1connect.NewTopicServiceClient)
		_, err := client.DeleteTopic(context.Background(), connect.NewRequest(&trendbirdv1.DeleteTopicRequest{
			Id: tp.ID,
		}))
		assertConnectCode(t, err, connect.CodeNotFound)
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.topicClient.DeleteTopic(context.Background(), connect.NewRequest(&trendbirdv1.DeleteTopicRequest{
			Id: "some-id",
		}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})

	t.Run("empty_topic_id", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		_, err := client.DeleteTopic(context.Background(), connect.NewRequest(&trendbirdv1.DeleteTopicRequest{
			Id: "",
		}))
		assertConnectCode(t, err, connect.CodeInternal)
	})

	t.Run("invalid_uuid_topic_id", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		_, err := client.DeleteTopic(context.Background(), connect.NewRequest(&trendbirdv1.DeleteTopicRequest{
			Id: "not-a-valid-uuid",
		}))
		assertConnectCode(t, err, connect.CodeInternal)
	})
}

func TestTopicService_DeleteTopic_SharedTopicPreserved(t *testing.T) {
	env := setupTest(t)
	userA := seedUser(t, env.db)
	userB := seedUser(t, env.db)

	tp := seedTopic(t, env.db, withTopicName("Shared"))
	seedUserTopic(t, env.db, userA.ID, tp.ID)
	seedUserTopic(t, env.db, userB.ID, tp.ID)

	// userA deletes
	clientA := connectClient(t, env, userA.ID, trendbirdv1connect.NewTopicServiceClient)
	_, err := clientA.DeleteTopic(context.Background(), connect.NewRequest(&trendbirdv1.DeleteTopicRequest{
		Id: tp.ID,
	}))
	if err != nil {
		t.Fatalf("DeleteTopic userA: %v", err)
	}

	// userA の user_topics は削除
	var utCountA int64
	env.db.Model(&model.UserTopic{}).Where("user_id = ? AND topic_id = ?", userA.ID, tp.ID).Count(&utCountA)
	if utCountA != 0 {
		t.Errorf("expected userA user_topic to be deleted, got %d", utCountA)
	}

	// userB の user_topics は残っている
	var utCountB int64
	env.db.Model(&model.UserTopic{}).Where("user_id = ? AND topic_id = ?", userB.ID, tp.ID).Count(&utCountB)
	if utCountB != 1 {
		t.Errorf("expected userB user_topic to remain, got %d", utCountB)
	}

	// topics レコード自体は残っている
	var topicCount int64
	env.db.Model(&model.Topic{}).Where("id = ?", tp.ID).Count(&topicCount)
	if topicCount != 1 {
		t.Errorf("expected topic to remain, got %d", topicCount)
	}
}

func TestTopicService_UpdateTopicNotification(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		tp := seedTopic(t, env.db)
		seedUserTopic(t, env.db, user.ID, tp.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		_, err := client.UpdateTopicNotification(context.Background(), connect.NewRequest(&trendbirdv1.UpdateTopicNotificationRequest{
			Id:      tp.ID,
			Enabled: false,
		}))
		if err != nil {
			t.Fatalf("UpdateTopicNotification: %v", err)
		}

		// DB 上の user_topics で notification_enabled が false に更新されていることを検証
		var updated model.UserTopic
		if err := env.db.Where("user_id = ? AND topic_id = ?", user.ID, tp.ID).First(&updated).Error; err != nil {
			t.Fatalf("failed to fetch user_topic: %v", err)
		}
		if updated.NotificationEnabled {
			t.Error("expected notification_enabled=false after update")
		}
	})

	t.Run("not_found", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		_, err := client.UpdateTopicNotification(context.Background(), connect.NewRequest(&trendbirdv1.UpdateTopicNotificationRequest{
			Id:      "00000000-0000-0000-0000-000000000000",
			Enabled: false,
		}))
		assertConnectCode(t, err, connect.CodeNotFound)
	})

	t.Run("permission_denied", func(t *testing.T) {
		env := setupTest(t)
		owner := seedUser(t, env.db)
		other := seedUser(t, env.db)
		tp := seedTopic(t, env.db)
		seedUserTopic(t, env.db, owner.ID, tp.ID)

		client := connectClient(t, env, other.ID, trendbirdv1connect.NewTopicServiceClient)
		_, err := client.UpdateTopicNotification(context.Background(), connect.NewRequest(&trendbirdv1.UpdateTopicNotificationRequest{
			Id:      tp.ID,
			Enabled: false,
		}))
		assertConnectCode(t, err, connect.CodeNotFound)
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.topicClient.UpdateTopicNotification(context.Background(), connect.NewRequest(&trendbirdv1.UpdateTopicNotificationRequest{
			Id:      "some-id",
			Enabled: false,
		}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})

	t.Run("empty_topic_id", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		_, err := client.UpdateTopicNotification(context.Background(), connect.NewRequest(&trendbirdv1.UpdateTopicNotificationRequest{
			Id:      "",
			Enabled: false,
		}))
		assertConnectCode(t, err, connect.CodeInternal)
	})
}

func TestTopicService_ListTopics_OnlyLinkedTopics(t *testing.T) {
	env := setupTest(t)
	user := seedUser(t, env.db)

	// Create 3 topics but only link 1 to the user
	tp1 := seedTopic(t, env.db, withTopicName("Linked"))
	seedUserTopic(t, env.db, user.ID, tp1.ID)
	seedTopic(t, env.db, withTopicName("Unlinked A"))
	seedTopic(t, env.db, withTopicName("Unlinked B"))

	client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
	resp, err := client.ListTopics(context.Background(), connect.NewRequest(&trendbirdv1.ListTopicsRequest{}))
	if err != nil {
		t.Fatalf("ListTopics: %v", err)
	}

	topics := resp.Msg.GetTopics()
	if got := len(topics); got != 1 {
		t.Fatalf("expected 1 topic, got %d", got)
	}
	if topics[0].GetName() != "Linked" {
		t.Errorf("name: want Linked, got %s", topics[0].GetName())
	}
}

func TestTopicService_CreateTopic_CustomTopicLimit(t *testing.T) {
	t.Run("can_add_existing_topic", func(t *testing.T) {
		env := setupTest(t)
		ensureGenre(t, env.db, "technology")

		// First user creates the topic
		proUser := seedUser(t, env.db)
		proClient := connectClient(t, env, proUser.ID, trendbirdv1connect.NewTopicServiceClient)
		resp, err := proClient.CreateTopic(context.Background(), connect.NewRequest(&trendbirdv1.CreateTopicRequest{
			Name:     "Existing Topic",
			Keywords: []string{"existing"},
			Genre:    "technology",
		}))
		if err != nil {
			t.Fatalf("Pro CreateTopic: %v", err)
		}
		topicID := resp.Msg.GetTopic().GetId()

		// Second user adds the same existing topic
		freeUser := seedUser(t, env.db)
		freeClient := connectClient(t, env, freeUser.ID, trendbirdv1connect.NewTopicServiceClient)
		resp2, err := freeClient.CreateTopic(context.Background(), connect.NewRequest(&trendbirdv1.CreateTopicRequest{
			Name:     "Existing Topic",
			Keywords: []string{"existing"},
			Genre:    "technology",
		}))
		if err != nil {
			t.Fatalf("Free CreateTopic (existing): %v", err)
		}
		if resp2.Msg.GetTopic().GetId() != topicID {
			t.Errorf("expected same topic ID %s, got %s", topicID, resp2.Msg.GetTopic().GetId())
		}

		// Verify is_creator=false for the free user
		var ut model.UserTopic
		if err := env.db.Where("user_id = ? AND topic_id = ?", freeUser.ID, topicID).First(&ut).Error; err != nil {
			t.Fatalf("failed to fetch user_topic: %v", err)
		}
		if ut.IsCreator {
			t.Error("expected is_creator=false for second user adding existing topic")
		}
	})

	t.Run("unlimited_custom_topics", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		ensureGenre(t, env.db, "technology")

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)

		// Create 5 custom topics — all should succeed
		for i := 1; i <= 5; i++ {
			_, err := client.CreateTopic(context.Background(), connect.NewRequest(&trendbirdv1.CreateTopicRequest{
				Name:     fmt.Sprintf("Pro Topic %d", i),
				Keywords: []string{"test"},
				Genre:    "technology",
			}))
			if err != nil {
				t.Fatalf("CreateTopic %d: %v", i, err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// TestTopicService_SuggestTopics
// ---------------------------------------------------------------------------

func TestTopicService_SuggestTopics(t *testing.T) {
	t.Run("query_match", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		seedTopic(t, env.db, withTopicName("AI News"), withTopicGenre("technology"))
		seedTopic(t, env.db, withTopicName("Blockchain Report"), withTopicGenre("technology"))
		seedTopic(t, env.db, withTopicName("Marketing Tips"), withTopicGenre("marketing"))

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		resp, err := client.SuggestTopics(context.Background(), connect.NewRequest(&trendbirdv1.SuggestTopicsRequest{
			Query: "AI News",
			Limit: 10,
		}))
		if err != nil {
			t.Fatalf("SuggestTopics: %v", err)
		}

		suggestions := resp.Msg.GetSuggestions()
		if len(suggestions) == 0 {
			t.Fatal("expected at least 1 suggestion")
		}
		// 最も類似度の高い結果が "AI News" であること
		if suggestions[0].GetName() != "AI News" {
			t.Errorf("first suggestion name: want AI News, got %s", suggestions[0].GetName())
		}
	})

	t.Run("genre_filter", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		seedTopic(t, env.db, withTopicName("Go Language"), withTopicGenre("technology"))
		seedTopic(t, env.db, withTopicName("Rust Language"), withTopicGenre("technology"))
		seedTopic(t, env.db, withTopicName("SEO Tips"), withTopicGenre("marketing"))

		genre := "technology"
		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		resp, err := client.SuggestTopics(context.Background(), connect.NewRequest(&trendbirdv1.SuggestTopicsRequest{
			Genre: &genre,
			Limit: 10,
		}))
		if err != nil {
			t.Fatalf("SuggestTopics genre_filter: %v", err)
		}

		suggestions := resp.Msg.GetSuggestions()
		if len(suggestions) != 2 {
			t.Fatalf("expected 2 suggestions, got %d", len(suggestions))
		}
		for _, s := range suggestions {
			if s.GetGenre() != "technology" {
				t.Errorf("suggestion genre: want technology, got %s", s.GetGenre())
			}
		}
	})

	t.Run("exclude_existing", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		tp1 := seedTopic(t, env.db, withTopicName("My Topic"), withTopicGenre("technology"))
		seedUserTopic(t, env.db, user.ID, tp1.ID) // ユーザーが既に持っているトピック
		seedTopic(t, env.db, withTopicName("Other Topic"), withTopicGenre("technology"))

		genre := "technology"
		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		resp, err := client.SuggestTopics(context.Background(), connect.NewRequest(&trendbirdv1.SuggestTopicsRequest{
			Genre: &genre,
			Limit: 10,
		}))
		if err != nil {
			t.Fatalf("SuggestTopics exclude_existing: %v", err)
		}

		suggestions := resp.Msg.GetSuggestions()
		if len(suggestions) != 1 {
			t.Fatalf("expected 1 suggestion (existing excluded), got %d", len(suggestions))
		}
		if suggestions[0].GetName() != "Other Topic" {
			t.Errorf("suggestion name: want Other Topic, got %s", suggestions[0].GetName())
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.topicClient.SuggestTopics(context.Background(), connect.NewRequest(&trendbirdv1.SuggestTopicsRequest{
			Query: "test",
		}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

// ---------------------------------------------------------------------------
// TestTopicService_ListGenres
// ---------------------------------------------------------------------------

func TestTopicService_ListGenres(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		ensureGenre(t, env.db, "technology")
		ensureGenre(t, env.db, "marketing")

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		resp, err := client.ListGenres(context.Background(), connect.NewRequest(&trendbirdv1.ListGenresRequest{}))
		if err != nil {
			t.Fatalf("ListGenres: %v", err)
		}

		genres := resp.Msg.GetGenres()
		if len(genres) < 2 {
			t.Fatalf("expected at least 2 genres, got %d", len(genres))
		}

		slugs := make(map[string]bool)
		for _, g := range genres {
			slugs[g.GetSlug()] = true
		}
		if !slugs["technology"] {
			t.Error("expected technology genre in results")
		}
		if !slugs["marketing"] {
			t.Error("expected marketing genre in results")
		}
	})

	t.Run("empty", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		// genres テーブルは truncateAll で TRUNCATE 済みなので 0 件のはず
		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		resp, err := client.ListGenres(context.Background(), connect.NewRequest(&trendbirdv1.ListGenresRequest{}))
		if err != nil {
			t.Fatalf("ListGenres: %v", err)
		}

		if got := len(resp.Msg.GetGenres()); got != 0 {
			t.Errorf("expected 0 genres, got %d", got)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.topicClient.ListGenres(context.Background(), connect.NewRequest(&trendbirdv1.ListGenresRequest{}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

// ---------------------------------------------------------------------------
// Idempotency Tests (3a-3d)
// ---------------------------------------------------------------------------

func TestTopicService_CreateTopic_Idempotent(t *testing.T) {
	t.Run("create_same_topic_twice_same_user", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		ensureGenre(t, env.db, "technology")

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)

		// 1st create
		resp1, err := client.CreateTopic(context.Background(), connect.NewRequest(&trendbirdv1.CreateTopicRequest{
			Name:     "Duplicate Topic",
			Keywords: []string{"dup"},
			Genre:    "technology",
		}))
		if err != nil {
			t.Fatalf("CreateTopic (1st): %v", err)
		}
		topicID1 := resp1.Msg.GetTopic().GetId()

		// 2nd create (same name+genre, same user)
		resp2, err := client.CreateTopic(context.Background(), connect.NewRequest(&trendbirdv1.CreateTopicRequest{
			Name:     "Duplicate Topic",
			Keywords: []string{"dup"},
			Genre:    "technology",
		}))
		if err != nil {
			t.Fatalf("CreateTopic (2nd): %v", err)
		}
		topicID2 := resp2.Msg.GetTopic().GetId()

		// Same topic ID
		if topicID1 != topicID2 {
			t.Errorf("expected same topic ID, got %s and %s", topicID1, topicID2)
		}

		// Only 1 user_topics record
		var utCount int64
		env.db.Model(&model.UserTopic{}).Where("user_id = ? AND topic_id = ?", user.ID, topicID1).Count(&utCount)
		if utCount != 1 {
			t.Errorf("expected 1 user_topic, got %d", utCount)
		}
	})
}

func TestTopicService_Idempotent(t *testing.T) {
	t.Run("delete_already_deleted", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		tp := seedTopic(t, env.db)
		seedUserTopic(t, env.db, user.ID, tp.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)

		// 1st delete → success
		_, err := client.DeleteTopic(context.Background(), connect.NewRequest(&trendbirdv1.DeleteTopicRequest{
			Id: tp.ID,
		}))
		if err != nil {
			t.Fatalf("DeleteTopic (1st): %v", err)
		}

		// 2nd delete → CodeNotFound (user_topic link already gone)
		_, err = client.DeleteTopic(context.Background(), connect.NewRequest(&trendbirdv1.DeleteTopicRequest{
			Id: tp.ID,
		}))
		assertConnectCode(t, err, connect.CodeNotFound)
	})

	t.Run("disable_notification_twice", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		tp := seedTopic(t, env.db)
		seedUserTopic(t, env.db, user.ID, tp.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)

		// Disable 1st time
		_, err := client.UpdateTopicNotification(context.Background(), connect.NewRequest(&trendbirdv1.UpdateTopicNotificationRequest{
			Id:      tp.ID,
			Enabled: false,
		}))
		if err != nil {
			t.Fatalf("UpdateTopicNotification (1st false): %v", err)
		}

		// Disable 2nd time → still succeeds
		_, err = client.UpdateTopicNotification(context.Background(), connect.NewRequest(&trendbirdv1.UpdateTopicNotificationRequest{
			Id:      tp.ID,
			Enabled: false,
		}))
		if err != nil {
			t.Fatalf("UpdateTopicNotification (2nd false): %v", err)
		}

		// DB is false
		var ut model.UserTopic
		env.db.Where("user_id = ? AND topic_id = ?", user.ID, tp.ID).First(&ut)
		if ut.NotificationEnabled {
			t.Error("expected notification_enabled=false")
		}
	})

	t.Run("enable_already_enabled", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		tp := seedTopic(t, env.db)
		seedUserTopic(t, env.db, user.ID, tp.ID) // default: notification_enabled=true

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)

		// Enable when already true → success
		_, err := client.UpdateTopicNotification(context.Background(), connect.NewRequest(&trendbirdv1.UpdateTopicNotificationRequest{
			Id:      tp.ID,
			Enabled: true,
		}))
		if err != nil {
			t.Fatalf("UpdateTopicNotification (true on true): %v", err)
		}

		// DB is still true
		var ut model.UserTopic
		env.db.Where("user_id = ? AND topic_id = ?", user.ID, tp.ID).First(&ut)
		if !ut.NotificationEnabled {
			t.Error("expected notification_enabled=true")
		}
	})
}

// ---------------------------------------------------------------------------
// Lifecycle Tests (4a-4c)
// ---------------------------------------------------------------------------

func TestTopicService_Lifecycle(t *testing.T) {
	t.Run("create_delete_recreate", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		ensureGenre(t, env.db, "technology")

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)

		// Create
		resp1, err := client.CreateTopic(context.Background(), connect.NewRequest(&trendbirdv1.CreateTopicRequest{
			Name:     "Lifecycle Topic",
			Keywords: []string{"life"},
			Genre:    "technology",
		}))
		if err != nil {
			t.Fatalf("CreateTopic: %v", err)
		}
		topicID := resp1.Msg.GetTopic().GetId()

		// Delete
		_, err = client.DeleteTopic(context.Background(), connect.NewRequest(&trendbirdv1.DeleteTopicRequest{
			Id: topicID,
		}))
		if err != nil {
			t.Fatalf("DeleteTopic: %v", err)
		}

		// ListTopics → 0
		listResp, err := client.ListTopics(context.Background(), connect.NewRequest(&trendbirdv1.ListTopicsRequest{}))
		if err != nil {
			t.Fatalf("ListTopics: %v", err)
		}
		if got := len(listResp.Msg.GetTopics()); got != 0 {
			t.Fatalf("expected 0 topics after delete, got %d", got)
		}

		// Recreate (same name+genre) → reuses same topic ID
		resp2, err := client.CreateTopic(context.Background(), connect.NewRequest(&trendbirdv1.CreateTopicRequest{
			Name:     "Lifecycle Topic",
			Keywords: []string{"life"},
			Genre:    "technology",
		}))
		if err != nil {
			t.Fatalf("CreateTopic (recreate): %v", err)
		}
		if resp2.Msg.GetTopic().GetId() != topicID {
			t.Errorf("expected reuse topic ID %s, got %s", topicID, resp2.Msg.GetTopic().GetId())
		}

		// ListTopics → 1
		listResp2, err := client.ListTopics(context.Background(), connect.NewRequest(&trendbirdv1.ListTopicsRequest{}))
		if err != nil {
			t.Fatalf("ListTopics (after recreate): %v", err)
		}
		if got := len(listResp2.Msg.GetTopics()); got != 1 {
			t.Errorf("expected 1 topic after recreate, got %d", got)
		}
	})

	t.Run("genre_topic_full_cycle", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		ensureGenre(t, env.db, "technology")

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)

		// AddGenre
		_, err := client.AddGenre(context.Background(), connect.NewRequest(&trendbirdv1.AddGenreRequest{
			Genre: "technology",
		}))
		if err != nil {
			t.Fatalf("AddGenre: %v", err)
		}

		// CreateTopic
		resp1, err := client.CreateTopic(context.Background(), connect.NewRequest(&trendbirdv1.CreateTopicRequest{
			Name:     "Genre Cycle Topic",
			Keywords: []string{"gc"},
			Genre:    "technology",
		}))
		if err != nil {
			t.Fatalf("CreateTopic: %v", err)
		}
		topicID := resp1.Msg.GetTopic().GetId()

		// RemoveGenre (cascade deletes user_topics for this genre)
		_, err = client.RemoveGenre(context.Background(), connect.NewRequest(&trendbirdv1.RemoveGenreRequest{
			Genre: "technology",
		}))
		if err != nil {
			t.Fatalf("RemoveGenre: %v", err)
		}

		// ListTopics → 0
		listResp, err := client.ListTopics(context.Background(), connect.NewRequest(&trendbirdv1.ListTopicsRequest{}))
		if err != nil {
			t.Fatalf("ListTopics (after RemoveGenre): %v", err)
		}
		if got := len(listResp.Msg.GetTopics()); got != 0 {
			t.Fatalf("expected 0 topics after RemoveGenre, got %d", got)
		}

		// Re-add genre
		_, err = client.AddGenre(context.Background(), connect.NewRequest(&trendbirdv1.AddGenreRequest{
			Genre: "technology",
		}))
		if err != nil {
			t.Fatalf("AddGenre (re-add): %v", err)
		}

		// Recreate topic (reuse)
		resp2, err := client.CreateTopic(context.Background(), connect.NewRequest(&trendbirdv1.CreateTopicRequest{
			Name:     "Genre Cycle Topic",
			Keywords: []string{"gc"},
			Genre:    "technology",
		}))
		if err != nil {
			t.Fatalf("CreateTopic (reuse): %v", err)
		}
		if resp2.Msg.GetTopic().GetId() != topicID {
			t.Errorf("expected reuse topic ID %s, got %s", topicID, resp2.Msg.GetTopic().GetId())
		}

		// ListTopics → 1
		listResp2, err := client.ListTopics(context.Background(), connect.NewRequest(&trendbirdv1.ListTopicsRequest{}))
		if err != nil {
			t.Fatalf("ListTopics (after re-add): %v", err)
		}
		if got := len(listResp2.Msg.GetTopics()); got != 1 {
			t.Errorf("expected 1 topic after re-add, got %d", got)
		}
	})

	t.Run("notification_toggle_lifecycle", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		tp := seedTopic(t, env.db)
		seedUserTopic(t, env.db, user.ID, tp.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)

		// GetTopic → true (default)
		resp1, err := client.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{Id: tp.ID}))
		if err != nil {
			t.Fatalf("GetTopic: %v", err)
		}
		if !resp1.Msg.GetTopic().GetNotificationEnabled() {
			t.Error("initial: expected notification_enabled=true")
		}

		// Update to false
		_, err = client.UpdateTopicNotification(context.Background(), connect.NewRequest(&trendbirdv1.UpdateTopicNotificationRequest{
			Id: tp.ID, Enabled: false,
		}))
		if err != nil {
			t.Fatalf("UpdateTopicNotification (false): %v", err)
		}

		// GetTopic → false
		resp2, err := client.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{Id: tp.ID}))
		if err != nil {
			t.Fatalf("GetTopic (after false): %v", err)
		}
		if resp2.Msg.GetTopic().GetNotificationEnabled() {
			t.Error("after disable: expected notification_enabled=false")
		}

		// Update to true
		_, err = client.UpdateTopicNotification(context.Background(), connect.NewRequest(&trendbirdv1.UpdateTopicNotificationRequest{
			Id: tp.ID, Enabled: true,
		}))
		if err != nil {
			t.Fatalf("UpdateTopicNotification (true): %v", err)
		}

		// GetTopic → true
		resp3, err := client.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{Id: tp.ID}))
		if err != nil {
			t.Fatalf("GetTopic (after true): %v", err)
		}
		if !resp3.Msg.GetTopic().GetNotificationEnabled() {
			t.Error("after re-enable: expected notification_enabled=true")
		}
	})
}

// ---------------------------------------------------------------------------
// Cross-User Isolation Tests (5a-5d)
// ---------------------------------------------------------------------------

func TestTopicService_CrossUserIsolation(t *testing.T) {
	t.Run("delete_isolation_via_list", func(t *testing.T) {
		env := setupTest(t)
		tp := seedTopic(t, env.db, withTopicName("Shared"))
		userA := seedUser(t, env.db)
		userB := seedUser(t, env.db)
		seedUserTopic(t, env.db, userA.ID, tp.ID)
		seedUserTopic(t, env.db, userB.ID, tp.ID)

		clientA := connectClient(t, env, userA.ID, trendbirdv1connect.NewTopicServiceClient)
		clientB := connectClient(t, env, userB.ID, trendbirdv1connect.NewTopicServiceClient)

		// userA deletes
		_, err := clientA.DeleteTopic(context.Background(), connect.NewRequest(&trendbirdv1.DeleteTopicRequest{
			Id: tp.ID,
		}))
		if err != nil {
			t.Fatalf("DeleteTopic userA: %v", err)
		}

		// userA.ListTopics → 0
		listA, err := clientA.ListTopics(context.Background(), connect.NewRequest(&trendbirdv1.ListTopicsRequest{}))
		if err != nil {
			t.Fatalf("ListTopics userA: %v", err)
		}
		if got := len(listA.Msg.GetTopics()); got != 0 {
			t.Errorf("userA: expected 0 topics, got %d", got)
		}

		// userB.ListTopics → 1
		listB, err := clientB.ListTopics(context.Background(), connect.NewRequest(&trendbirdv1.ListTopicsRequest{}))
		if err != nil {
			t.Fatalf("ListTopics userB: %v", err)
		}
		if got := len(listB.Msg.GetTopics()); got != 1 {
			t.Errorf("userB: expected 1 topic, got %d", got)
		}
	})

	t.Run("notification_isolation_via_api", func(t *testing.T) {
		env := setupTest(t)
		tp := seedTopic(t, env.db)
		userA := seedUser(t, env.db)
		userB := seedUser(t, env.db)
		seedUserTopic(t, env.db, userA.ID, tp.ID)
		seedUserTopic(t, env.db, userB.ID, tp.ID)

		clientA := connectClient(t, env, userA.ID, trendbirdv1connect.NewTopicServiceClient)
		clientB := connectClient(t, env, userB.ID, trendbirdv1connect.NewTopicServiceClient)

		// userA disables notification
		_, err := clientA.UpdateTopicNotification(context.Background(), connect.NewRequest(&trendbirdv1.UpdateTopicNotificationRequest{
			Id: tp.ID, Enabled: false,
		}))
		if err != nil {
			t.Fatalf("UpdateTopicNotification userA: %v", err)
		}

		// userB GetTopic → still true
		respB, err := clientB.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{Id: tp.ID}))
		if err != nil {
			t.Fatalf("GetTopic userB: %v", err)
		}
		if !respB.Msg.GetTopic().GetNotificationEnabled() {
			t.Error("userB: expected notification_enabled=true")
		}
	})

	t.Run("genre_removal_isolation", func(t *testing.T) {
		env := setupTest(t)
		ensureGenre(t, env.db, "technology")

		userA := seedUser(t, env.db)
		userB := seedUser(t, env.db)

		// Both users have technology genre and a shared topic
		seedUserGenre(t, env.db, userA.ID, "technology")
		seedUserGenre(t, env.db, userB.ID, "technology")
		tp := seedTopic(t, env.db, withTopicGenre("technology"), withTopicName("Tech Topic"))
		seedUserTopic(t, env.db, userA.ID, tp.ID)
		seedUserTopic(t, env.db, userB.ID, tp.ID)

		clientA := connectClient(t, env, userA.ID, trendbirdv1connect.NewTopicServiceClient)
		clientB := connectClient(t, env, userB.ID, trendbirdv1connect.NewTopicServiceClient)

		// userA removes genre (cascade)
		_, err := clientA.RemoveGenre(context.Background(), connect.NewRequest(&trendbirdv1.RemoveGenreRequest{
			Genre: "technology",
		}))
		if err != nil {
			t.Fatalf("RemoveGenre userA: %v", err)
		}

		// userB ListTopics → still 1
		listB, err := clientB.ListTopics(context.Background(), connect.NewRequest(&trendbirdv1.ListTopicsRequest{}))
		if err != nil {
			t.Fatalf("ListTopics userB: %v", err)
		}
		if got := len(listB.Msg.GetTopics()); got != 1 {
			t.Errorf("userB: expected 1 topic, got %d", got)
		}

		// userB ListUserGenres → still has "technology"
		genresB, err := clientB.ListUserGenres(context.Background(), connect.NewRequest(&trendbirdv1.ListUserGenresRequest{}))
		if err != nil {
			t.Fatalf("ListUserGenres userB: %v", err)
		}
		if len(genresB.Msg.GetGenres()) != 1 || genresB.Msg.GetGenres()[0] != "technology" {
			t.Errorf("userB: expected [technology], got %v", genresB.Msg.GetGenres())
		}
	})

	t.Run("suggest_excludes_only_own", func(t *testing.T) {
		env := setupTest(t)
		ensureGenre(t, env.db, "technology")

		userA := seedUser(t, env.db)
		userB := seedUser(t, env.db)

		tpA := seedTopic(t, env.db, withTopicName("TopicForA"), withTopicGenre("technology"))
		seedUserTopic(t, env.db, userA.ID, tpA.ID)
		tpB := seedTopic(t, env.db, withTopicName("TopicForB"), withTopicGenre("technology"))
		seedUserTopic(t, env.db, userB.ID, tpB.ID)

		genre := "technology"
		clientA := connectClient(t, env, userA.ID, trendbirdv1connect.NewTopicServiceClient)
		respA, err := clientA.SuggestTopics(context.Background(), connect.NewRequest(&trendbirdv1.SuggestTopicsRequest{
			Genre: &genre, Limit: 10,
		}))
		if err != nil {
			t.Fatalf("SuggestTopics userA: %v", err)
		}
		// userA should see TopicForB (not TopicForA)
		sugA := respA.Msg.GetSuggestions()
		if len(sugA) != 1 {
			t.Fatalf("userA: expected 1 suggestion, got %d", len(sugA))
		}
		if sugA[0].GetName() != "TopicForB" {
			t.Errorf("userA: expected TopicForB, got %s", sugA[0].GetName())
		}

		clientB := connectClient(t, env, userB.ID, trendbirdv1connect.NewTopicServiceClient)
		respB, err := clientB.SuggestTopics(context.Background(), connect.NewRequest(&trendbirdv1.SuggestTopicsRequest{
			Genre: &genre, Limit: 10,
		}))
		if err != nil {
			t.Fatalf("SuggestTopics userB: %v", err)
		}
		// userB should see TopicForA (not TopicForB)
		sugB := respB.Msg.GetSuggestions()
		if len(sugB) != 1 {
			t.Fatalf("userB: expected 1 suggestion, got %d", len(sugB))
		}
		if sugB[0].GetName() != "TopicForA" {
			t.Errorf("userB: expected TopicForA, got %s", sugB[0].GetName())
		}
	})
}

// ---------------------------------------------------------------------------
// Shared Resource Tests (6a-6c)
// ---------------------------------------------------------------------------

func TestTopicService_SharedResource(t *testing.T) {
	t.Run("three_users_share_topic", func(t *testing.T) {
		env := setupTest(t)
		ensureGenre(t, env.db, "technology")

		users := make([]*model.User, 3)
		for i := range users {
			users[i] = seedUser(t, env.db)
		}

		var topicIDs [3]string
		for i, u := range users {
			client := connectClient(t, env, u.ID, trendbirdv1connect.NewTopicServiceClient)
			resp, err := client.CreateTopic(context.Background(), connect.NewRequest(&trendbirdv1.CreateTopicRequest{
				Name:     "Shared By 3",
				Keywords: []string{"shared"},
				Genre:    "technology",
			}))
			if err != nil {
				t.Fatalf("CreateTopic user %d: %v", i, err)
			}
			topicIDs[i] = resp.Msg.GetTopic().GetId()
		}

		// All 3 should have the same topic ID
		if topicIDs[0] != topicIDs[1] || topicIDs[1] != topicIDs[2] {
			t.Errorf("expected all same ID, got %v", topicIDs)
		}

		// Only first user should be creator
		var ut0 model.UserTopic
		env.db.Where("user_id = ? AND topic_id = ?", users[0].ID, topicIDs[0]).First(&ut0)
		if !ut0.IsCreator {
			t.Error("user[0]: expected is_creator=true")
		}

		for i := 1; i < 3; i++ {
			var ut model.UserTopic
			env.db.Where("user_id = ? AND topic_id = ?", users[i].ID, topicIDs[i]).First(&ut)
			if ut.IsCreator {
				t.Errorf("user[%d]: expected is_creator=false", i)
			}
		}
	})

	t.Run("survives_progressive_deletion", func(t *testing.T) {
		env := setupTest(t)
		tp := seedTopic(t, env.db, withTopicName("Resilient"))
		users := make([]*model.User, 3)
		for i := range users {
			users[i] = seedUser(t, env.db)
			seedUserTopic(t, env.db, users[i].ID, tp.ID)
		}

		// User1 deletes
		client1 := connectClient(t, env, users[0].ID, trendbirdv1connect.NewTopicServiceClient)
		_, err := client1.DeleteTopic(context.Background(), connect.NewRequest(&trendbirdv1.DeleteTopicRequest{Id: tp.ID}))
		if err != nil {
			t.Fatalf("DeleteTopic user1: %v", err)
		}

		// User2 deletes
		client2 := connectClient(t, env, users[1].ID, trendbirdv1connect.NewTopicServiceClient)
		_, err = client2.DeleteTopic(context.Background(), connect.NewRequest(&trendbirdv1.DeleteTopicRequest{Id: tp.ID}))
		if err != nil {
			t.Fatalf("DeleteTopic user2: %v", err)
		}

		// User3 can still GetTopic
		client3 := connectClient(t, env, users[2].ID, trendbirdv1connect.NewTopicServiceClient)
		resp, err := client3.GetTopic(context.Background(), connect.NewRequest(&trendbirdv1.GetTopicRequest{Id: tp.ID}))
		if err != nil {
			t.Fatalf("GetTopic user3: %v", err)
		}
		if resp.Msg.GetTopic().GetName() != "Resilient" {
			t.Errorf("topic name: want Resilient, got %s", resp.Msg.GetTopic().GetName())
		}

		// topics record still exists
		var topicCount int64
		env.db.Model(&model.Topic{}).Where("id = ?", tp.ID).Count(&topicCount)
		if topicCount != 1 {
			t.Errorf("expected topic record to remain, got count=%d", topicCount)
		}
	})

	t.Run("reuse_preserves_original_keywords", func(t *testing.T) {
		env := setupTest(t)
		ensureGenre(t, env.db, "technology")

		userA := seedUser(t, env.db)
		userB := seedUser(t, env.db)

		clientA := connectClient(t, env, userA.ID, trendbirdv1connect.NewTopicServiceClient)
		respA, err := clientA.CreateTopic(context.Background(), connect.NewRequest(&trendbirdv1.CreateTopicRequest{
			Name:     "Keyword Topic",
			Keywords: []string{"a", "b"},
			Genre:    "technology",
		}))
		if err != nil {
			t.Fatalf("CreateTopic userA: %v", err)
		}
		topicIDA := respA.Msg.GetTopic().GetId()

		clientB := connectClient(t, env, userB.ID, trendbirdv1connect.NewTopicServiceClient)
		respB, err := clientB.CreateTopic(context.Background(), connect.NewRequest(&trendbirdv1.CreateTopicRequest{
			Name:     "Keyword Topic",
			Keywords: []string{"c", "d"},
			Genre:    "technology",
		}))
		if err != nil {
			t.Fatalf("CreateTopic userB: %v", err)
		}

		// Same topic ID
		if respB.Msg.GetTopic().GetId() != topicIDA {
			t.Errorf("expected same topic ID, got A=%s B=%s", topicIDA, respB.Msg.GetTopic().GetId())
		}

		// Keywords preserved from original creator
		var dbTopic model.Topic
		env.db.First(&dbTopic, "id = ?", topicIDA)
		// JSON encoding may include spaces: ["a", "b"] or ["a","b"]
		if dbTopic.Keywords != `["a","b"]` && dbTopic.Keywords != `["a", "b"]` {
			t.Errorf("expected original keywords [a,b], got %s", dbTopic.Keywords)
		}
	})
}

// ---------------------------------------------------------------------------
// TestTopicService_CreateTopic_ConcurrentSameName
// ---------------------------------------------------------------------------

func TestTopicService_CreateTopic_ConcurrentSameName(t *testing.T) {
	env := setupTest(t)
	ensureGenre(t, env.db, "technology")

	user1 := seedUser(t, env.db)
	user2 := seedUser(t, env.db)
	seedUserGenre(t, env.db, user1.ID, "technology")
	seedUserGenre(t, env.db, user2.ID, "technology")

	topicName := fmt.Sprintf("Concurrent Topic %d", nextSeq())

	type result struct {
		topicID string
		err     error
	}
	ch := make(chan result, 2)

	// goroutine 1: user1 が CreateTopic
	go func() {
		client := connectClient(t, env, user1.ID, trendbirdv1connect.NewTopicServiceClient)
		resp, err := client.CreateTopic(context.Background(), connect.NewRequest(&trendbirdv1.CreateTopicRequest{
			Name:     topicName,
			Keywords: []string{"concurrent"},
			Genre:    "technology",
		}))
		if err != nil {
			ch <- result{err: err}
			return
		}
		ch <- result{topicID: resp.Msg.GetTopic().GetId()}
	}()

	// goroutine 2: user2 が同名で CreateTopic
	go func() {
		client := connectClient(t, env, user2.ID, trendbirdv1connect.NewTopicServiceClient)
		resp, err := client.CreateTopic(context.Background(), connect.NewRequest(&trendbirdv1.CreateTopicRequest{
			Name:     topicName,
			Keywords: []string{"concurrent"},
			Genre:    "technology",
		}))
		if err != nil {
			ch <- result{err: err}
			return
		}
		ch <- result{topicID: resp.Msg.GetTopic().GetId()}
	}()

	// 結果収集
	r1 := <-ch
	r2 := <-ch

	if r1.err != nil {
		t.Errorf("goroutine 1 failed: %v", r1.err)
	}
	if r2.err != nil {
		t.Errorf("goroutine 2 failed: %v", r2.err)
	}
	if r1.err != nil || r2.err != nil {
		t.FailNow()
	}

	// 両方成功し、同じ topic ID を返すはず（ON CONFLICT DO NOTHING + 再取得パターン）
	if r1.topicID != r2.topicID {
		t.Errorf("expected same topic ID, got %s and %s", r1.topicID, r2.topicID)
	}

	// topics テーブルに1行のみ
	var topicCount int64
	env.db.Model(&model.Topic{}).Where("name = ?", topicName).Count(&topicCount)
	if topicCount != 1 {
		t.Errorf("topics count: want 1, got %d", topicCount)
	}

	// 両 user の user_topics にリンクが存在
	var ut1Count int64
	env.db.Model(&model.UserTopic{}).Where("user_id = ? AND topic_id = ?", user1.ID, r1.topicID).Count(&ut1Count)
	if ut1Count != 1 {
		t.Errorf("user1 user_topics: want 1, got %d", ut1Count)
	}

	var ut2Count int64
	env.db.Model(&model.UserTopic{}).Where("user_id = ? AND topic_id = ?", user2.ID, r2.topicID).Count(&ut2Count)
	if ut2Count != 1 {
		t.Errorf("user2 user_topics: want 1, got %d", ut2Count)
	}
}

// ---------------------------------------------------------------------------
// TestTopicService_CreateTopic_AutoAddGenreAndNotificationSettings
// ---------------------------------------------------------------------------

func TestTopicService_CreateTopic_AutoAddGenreAndNotificationSettings(t *testing.T) {
	env := setupTest(t)

	user := seedUser(t, env.db)
	genreID := ensureGenre(t, env.db, "technology")
	// genre 未追加、notification_settings なし

	client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
	resp, err := client.CreateTopic(context.Background(), connect.NewRequest(&trendbirdv1.CreateTopicRequest{
		Name:     "Atomicity Test Topic",
		Keywords: []string{"test"},
		Genre:    "technology",
	}))
	if err != nil {
		t.Fatalf("CreateTopic: %v", err)
	}

	topicID := resp.Msg.GetTopic().GetId()

	// user_genres 自動追加
	var ugCount int64
	env.db.Model(&model.UserGenre{}).Where("user_id = ? AND genre_id = ?", user.ID, genreID).Count(&ugCount)
	if ugCount != 1 {
		t.Errorf("user_genres: want 1 (auto-added), got %d", ugCount)
	}

	// user_topics 作成
	var ut model.UserTopic
	if err := env.db.Where("user_id = ? AND topic_id = ?", user.ID, topicID).First(&ut).Error; err != nil {
		t.Fatalf("user_topic not found: %v", err)
	}

	// notification_enabled=true
	if !ut.NotificationEnabled {
		t.Error("user_topic.notification_enabled: want true, got false")
	}
}
