package e2etest

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"

	trendbirdv1 "github.com/trendbird/backend/gen/trendbird/v1"
	"github.com/trendbird/backend/gen/trendbird/v1/trendbirdv1connect"
)

func TestAnalyticsService_ImportDailyAnalytics(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		client := connectClient(t, env, user.ID, trendbirdv1connect.NewAnalyticsServiceClient)

		resp, err := client.ImportDailyAnalytics(context.Background(), connect.NewRequest(&trendbirdv1.ImportDailyAnalyticsRequest{
			Records: []*trendbirdv1.DailyAnalytics{
				{Date: "2026-03-25", Impressions: 100, Likes: 10, Engagements: 20, NewFollows: 5},
				{Date: "2026-03-26", Impressions: 200, Likes: 20, Engagements: 40, NewFollows: 8},
			},
		}))
		if err != nil {
			t.Fatalf("ImportDailyAnalytics: %v", err)
		}
		if resp.Msg.GetImportedCount() != 2 {
			t.Errorf("imported_count: want 2, got %d", resp.Msg.GetImportedCount())
		}
	})

	t.Run("upsert_idempotent", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		client := connectClient(t, env, user.ID, trendbirdv1connect.NewAnalyticsServiceClient)

		records := []*trendbirdv1.DailyAnalytics{
			{Date: "2026-03-25", Impressions: 100, Likes: 10},
		}

		// First import
		_, err := client.ImportDailyAnalytics(context.Background(), connect.NewRequest(&trendbirdv1.ImportDailyAnalyticsRequest{Records: records}))
		if err != nil {
			t.Fatalf("first import: %v", err)
		}

		// Second import (same date, updated values)
		records[0].Impressions = 200
		resp, err := client.ImportDailyAnalytics(context.Background(), connect.NewRequest(&trendbirdv1.ImportDailyAnalyticsRequest{Records: records}))
		if err != nil {
			t.Fatalf("second import: %v", err)
		}
		if resp.Msg.GetUpdatedCount() != 1 {
			t.Errorf("updated_count: want 1, got %d", resp.Msg.GetUpdatedCount())
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)
		_, err := env.analyticsClient.ImportDailyAnalytics(context.Background(), connect.NewRequest(&trendbirdv1.ImportDailyAnalyticsRequest{
			Records: []*trendbirdv1.DailyAnalytics{{Date: "2026-03-25", Impressions: 100}},
		}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

func TestAnalyticsService_ImportPostAnalytics(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		client := connectClient(t, env, user.ID, trendbirdv1connect.NewAnalyticsServiceClient)

		resp, err := client.ImportPostAnalytics(context.Background(), connect.NewRequest(&trendbirdv1.ImportPostAnalyticsRequest{
			Records: []*trendbirdv1.PostAnalytics{
				{PostId: "12345", PostedAt: "2026-03-25T10:00:00Z", PostText: "テスト投稿", Impressions: 500, Likes: 10},
				{PostId: "12346", PostedAt: "2026-03-26T10:00:00Z", PostText: "もうひとつ", Impressions: 300, Likes: 5},
			},
		}))
		if err != nil {
			t.Fatalf("ImportPostAnalytics: %v", err)
		}
		if resp.Msg.GetImportedCount() != 2 {
			t.Errorf("imported_count: want 2, got %d", resp.Msg.GetImportedCount())
		}
	})
}

func TestAnalyticsService_GetAnalyticsSummary(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		client := connectClient(t, env, user.ID, trendbirdv1connect.NewAnalyticsServiceClient)

		// Import data first
		_, err := client.ImportDailyAnalytics(context.Background(), connect.NewRequest(&trendbirdv1.ImportDailyAnalyticsRequest{
			Records: []*trendbirdv1.DailyAnalytics{
				{Date: "2026-03-25", Impressions: 100, Likes: 10, Engagements: 20, NewFollows: 5, Unfollows: 1},
				{Date: "2026-03-26", Impressions: 200, Likes: 20, Engagements: 40, NewFollows: 8, Unfollows: 2},
			},
		}))
		if err != nil {
			t.Fatalf("import: %v", err)
		}

		start := "2026-03-25"
		end := "2026-03-26"
		resp, err := client.GetAnalyticsSummary(context.Background(), connect.NewRequest(&trendbirdv1.GetAnalyticsSummaryRequest{
			StartDate: &start,
			EndDate:   &end,
		}))
		if err != nil {
			t.Fatalf("GetAnalyticsSummary: %v", err)
		}

		s := resp.Msg.GetSummary()
		if s.GetTotalImpressions() != 300 {
			t.Errorf("total_impressions: want 300, got %d", s.GetTotalImpressions())
		}
		if s.GetTotalLikes() != 30 {
			t.Errorf("total_likes: want 30, got %d", s.GetTotalLikes())
		}
		if s.GetDaysCount() != 2 {
			t.Errorf("days_count: want 2, got %d", s.GetDaysCount())
		}
		if len(s.GetDailyData()) != 2 {
			t.Errorf("daily_data count: want 2, got %d", len(s.GetDailyData()))
		}
	})

	t.Run("empty", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		client := connectClient(t, env, user.ID, trendbirdv1connect.NewAnalyticsServiceClient)

		resp, err := client.GetAnalyticsSummary(context.Background(), connect.NewRequest(&trendbirdv1.GetAnalyticsSummaryRequest{}))
		if err != nil {
			t.Fatalf("GetAnalyticsSummary: %v", err)
		}

		if resp.Msg.GetSummary().GetDaysCount() != 0 {
			t.Errorf("days_count: want 0, got %d", resp.Msg.GetSummary().GetDaysCount())
		}
	})
}

func TestAnalyticsService_ListPostAnalytics(t *testing.T) {
	t.Run("success_sorted", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		client := connectClient(t, env, user.ID, trendbirdv1connect.NewAnalyticsServiceClient)

		// Import posts
		_, err := client.ImportPostAnalytics(context.Background(), connect.NewRequest(&trendbirdv1.ImportPostAnalyticsRequest{
			Records: []*trendbirdv1.PostAnalytics{
				{PostId: "1", PostedAt: "2026-03-25T10:00:00Z", PostText: "低い", Impressions: 100},
				{PostId: "2", PostedAt: "2026-03-25T11:00:00Z", PostText: "高い", Impressions: 500},
				{PostId: "3", PostedAt: "2026-03-25T12:00:00Z", PostText: "中間", Impressions: 300},
			},
		}))
		if err != nil {
			t.Fatalf("import: %v", err)
		}

		sortBy := "impressions"
		limit := int32(10)
		resp, err := client.ListPostAnalytics(context.Background(), connect.NewRequest(&trendbirdv1.ListPostAnalyticsRequest{
			SortBy: &sortBy,
			Limit:  &limit,
		}))
		if err != nil {
			t.Fatalf("ListPostAnalytics: %v", err)
		}

		posts := resp.Msg.GetPosts()
		if len(posts) != 3 {
			t.Fatalf("count: want 3, got %d", len(posts))
		}
		if posts[0].GetImpressions() != 500 {
			t.Errorf("first post impressions: want 500, got %d", posts[0].GetImpressions())
		}
		if posts[0].GetPostText() != "高い" {
			t.Errorf("first post text: want '高い', got %q", posts[0].GetPostText())
		}
	})

	t.Run("cross_user_isolation", func(t *testing.T) {
		env := setupTest(t)
		user1 := seedUser(t, env.db)
		user2 := seedUser(t, env.db)
		client1 := connectClient(t, env, user1.ID, trendbirdv1connect.NewAnalyticsServiceClient)
		client2 := connectClient(t, env, user2.ID, trendbirdv1connect.NewAnalyticsServiceClient)

		_, err := client1.ImportPostAnalytics(context.Background(), connect.NewRequest(&trendbirdv1.ImportPostAnalyticsRequest{
			Records: []*trendbirdv1.PostAnalytics{
				{PostId: "1", PostedAt: "2026-03-25T10:00:00Z", PostText: "user1", Impressions: 100},
			},
		}))
		if err != nil {
			t.Fatalf("import user1: %v", err)
		}

		limit := int32(10)
		resp, err := client2.ListPostAnalytics(context.Background(), connect.NewRequest(&trendbirdv1.ListPostAnalyticsRequest{
			Limit: &limit,
		}))
		if err != nil {
			t.Fatalf("ListPostAnalytics user2: %v", err)
		}

		if len(resp.Msg.GetPosts()) != 0 {
			t.Errorf("user2 should see 0 posts, got %d", len(resp.Msg.GetPosts()))
		}
	})
}

func TestAnalyticsService_GetGrowthInsights(t *testing.T) {
	t.Run("with_data", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		client := connectClient(t, env, user.ID, trendbirdv1connect.NewAnalyticsServiceClient)

		// Import both daily and post data
		_, err := client.ImportDailyAnalytics(context.Background(), connect.NewRequest(&trendbirdv1.ImportDailyAnalyticsRequest{
			Records: []*trendbirdv1.DailyAnalytics{
				{Date: "2026-03-25", Impressions: 1000, Likes: 20, Engagements: 50, NewFollows: 10, Unfollows: 2, PostsCreated: 3},
			},
		}))
		if err != nil {
			t.Fatalf("import daily: %v", err)
		}

		_, err = client.ImportPostAnalytics(context.Background(), connect.NewRequest(&trendbirdv1.ImportPostAnalyticsRequest{
			Records: []*trendbirdv1.PostAnalytics{
				{PostId: "1", PostedAt: "2026-03-25T10:00:00Z", PostText: "ベストポスト", Impressions: 500},
			},
		}))
		if err != nil {
			t.Fatalf("import posts: %v", err)
		}

		start := "2026-03-25"
		end := "2026-03-25"
		resp, err := client.GetGrowthInsights(context.Background(), connect.NewRequest(&trendbirdv1.GetGrowthInsightsRequest{
			StartDate: &start,
			EndDate:   &end,
		}))
		if err != nil {
			t.Fatalf("GetGrowthInsights: %v", err)
		}

		if len(resp.Msg.GetInsights()) == 0 {
			t.Error("expected at least one insight")
		}
		if resp.Msg.GetSummary() == nil {
			t.Error("expected summary in response")
		}
	})

	t.Run("no_data", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		client := connectClient(t, env, user.ID, trendbirdv1connect.NewAnalyticsServiceClient)

		resp, err := client.GetGrowthInsights(context.Background(), connect.NewRequest(&trendbirdv1.GetGrowthInsightsRequest{}))
		if err != nil {
			t.Fatalf("GetGrowthInsights: %v", err)
		}

		insights := resp.Msg.GetInsights()
		if len(insights) != 1 {
			t.Fatalf("want 1 'no data' insight, got %d", len(insights))
		}
		if insights[0].GetCategory() != "data" {
			t.Errorf("category: want 'data', got %q", insights[0].GetCategory())
		}
	})
}

// Suppress unused import warning for time package
var _ = time.Now
