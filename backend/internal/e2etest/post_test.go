package e2etest

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	trendbirdv1 "github.com/trendbird/backend/gen/trendbird/v1"
	"github.com/trendbird/backend/gen/trendbird/v1/trendbirdv1connect"
	"github.com/trendbird/backend/internal/domain/gateway"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

// ---------------------------------------------------------------------------
// TestPostService_GeneratePosts
// ---------------------------------------------------------------------------

func TestPostService_GeneratePosts(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		topic := seedTopic(t, env.db)
		seedUserTopic(t, env.db, user.ID, topic.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		resp, err := client.GeneratePosts(context.Background(), connect.NewRequest(&trendbirdv1.GeneratePostsRequest{
			TopicId: topic.ID,
		}))
		if err != nil {
			t.Fatalf("GeneratePosts: %v", err)
		}

		posts := resp.Msg.GetPosts()
		if got := len(posts); got != 3 {
			t.Fatalf("expected 3 generated posts, got %d", got)
		}
		for _, p := range posts {
			if p.GetContent() == "" {
				t.Error("generated post content should not be empty")
			}
			if p.GetTopicId() != topic.ID {
				t.Errorf("topic_id: want %s, got %s", topic.ID, p.GetTopicId())
			}
		}

		// DB: generated_posts 3件
		var genCount int64
		env.db.Model(&model.GeneratedPost{}).Where("user_id = ?", user.ID).Count(&genCount)
		if genCount != 3 {
			t.Errorf("generated_posts count: want 3, got %d", genCount)
		}

		// DB: ai_generation_logs 1件
		var logCount int64
		env.db.Model(&model.AIGenerationLog{}).Where("user_id = ?", user.ID).Count(&logCount)
		if logCount != 1 {
			t.Errorf("ai_generation_logs count: want 1, got %d", logCount)
		}
	})

	t.Run("monthly_limit_reset_on_new_month", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		topic := seedTopic(t, env.db)
		seedUserTopic(t, env.db, user.ID, topic.ID)

		// 前月に10回生成（上限到達）
		for i := 0; i < 10; i++ {
			seedAIGenerationLog(t, env.db, user.ID)
		}
		lastMonth := time.Now().AddDate(0, -1, 0)
		env.db.Exec("UPDATE ai_generation_logs SET created_at = ? WHERE user_id = ?", lastMonth, user.ID)

		// 今月は0回 → 生成成功（前月のログはカウントされない）
		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		_, err := client.GeneratePosts(context.Background(), connect.NewRequest(&trendbirdv1.GeneratePostsRequest{
			TopicId: topic.ID,
		}))
		if err != nil {
			t.Fatalf("GeneratePosts should succeed after month reset: %v", err)
		}
	})

	t.Run("topic_permission_denied", func(t *testing.T) {
		env := setupTest(t)
		owner := seedUser(t, env.db)
		other := seedUser(t, env.db)
		topic := seedTopic(t, env.db)
		seedUserTopic(t, env.db, owner.ID, topic.ID)

		// other ユーザーが owner のトピック ID で GeneratePosts → NotFound
		client := connectClient(t, env, other.ID, trendbirdv1connect.NewPostServiceClient)
		_, err := client.GeneratePosts(context.Background(), connect.NewRequest(&trendbirdv1.GeneratePostsRequest{
			TopicId: topic.ID,
		}))
		assertConnectCode(t, err, connect.CodeNotFound)
	})

	t.Run("ai_gateway_failure", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		topic := seedTopic(t, env.db)
		seedUserTopic(t, env.db, user.ID, topic.ID)

		// GeneratePostsFn をエラー返却に差し替え
		env.mockAI.GeneratePostsFn = func(_ context.Context, _ gateway.GeneratePostsInput) (*gateway.GeneratePostsOutput, error) {
			return nil, fmt.Errorf("OpenAI API rate limit exceeded")
		}

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		_, err := client.GeneratePosts(context.Background(), connect.NewRequest(&trendbirdv1.GeneratePostsRequest{
			TopicId: topic.ID,
		}))
		assertConnectCode(t, err, connect.CodeInternal)
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.postClient.GeneratePosts(context.Background(), connect.NewRequest(&trendbirdv1.GeneratePostsRequest{
			TopicId: "some-id",
		}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

// ---------------------------------------------------------------------------
// TestPostService_CreateDraft
// ---------------------------------------------------------------------------

func TestPostService_CreateDraft(t *testing.T) {
	t.Run("empty_content", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		_, err := client.CreateDraft(context.Background(), connect.NewRequest(&trendbirdv1.CreateDraftRequest{
			Content: "",
		}))
		assertConnectCode(t, err, connect.CodeInvalidArgument)
	})

	t.Run("normal", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		topic := seedTopic(t, env.db)
		seedUserTopic(t, env.db, user.ID, topic.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		resp, err := client.CreateDraft(context.Background(), connect.NewRequest(&trendbirdv1.CreateDraftRequest{
			Content: "Test draft content",
			TopicId: &topic.ID,
		}))
		if err != nil {
			t.Fatalf("CreateDraft: %v", err)
		}

		draft := resp.Msg.GetDraft()
		if draft.GetContent() != "Test draft content" {
			t.Errorf("content: want %q, got %q", "Test draft content", draft.GetContent())
		}
		if draft.GetStatus() != trendbirdv1.PostStatus_POST_STATUS_DRAFT {
			t.Errorf("status: want POST_STATUS_DRAFT, got %v", draft.GetStatus())
		}
		if draft.GetTopicId() != topic.ID {
			t.Errorf("topic_id: want %s, got %s", topic.ID, draft.GetTopicId())
		}

		// DB: status=1 (DRAFT) の投稿が作成されていること
		var dbPost model.Post
		if err := env.db.First(&dbPost, "id = ?", draft.GetId()).Error; err != nil {
			t.Fatalf("failed to fetch post from DB: %v", err)
		}
		if dbPost.Status != 1 {
			t.Errorf("DB status: want 1 (DRAFT), got %d", dbPost.Status)
		}
		if dbPost.Content != "Test draft content" {
			t.Errorf("DB content: want %q, got %q", "Test draft content", dbPost.Content)
		}
	})

	t.Run("topic_permission_denied", func(t *testing.T) {
		env := setupTest(t)
		owner := seedUser(t, env.db)
		other := seedUser(t, env.db)
		topic := seedTopic(t, env.db)
		seedUserTopic(t, env.db, owner.ID, topic.ID)

		// other ユーザーが owner のトピック ID で CreateDraft → NotFound
		client := connectClient(t, env, other.ID, trendbirdv1connect.NewPostServiceClient)
		_, err := client.CreateDraft(context.Background(), connect.NewRequest(&trendbirdv1.CreateDraftRequest{
			Content: "draft with other's topic",
			TopicId: &topic.ID,
		}))
		assertConnectCode(t, err, connect.CodeNotFound)
	})

	t.Run("without_topic", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		resp, err := client.CreateDraft(context.Background(), connect.NewRequest(&trendbirdv1.CreateDraftRequest{
			Content: "Draft without topic",
		}))
		if err != nil {
			t.Fatalf("CreateDraft without topic: %v", err)
		}

		draft := resp.Msg.GetDraft()
		if draft.GetContent() != "Draft without topic" {
			t.Errorf("content: want %q, got %q", "Draft without topic", draft.GetContent())
		}
		if draft.GetTopicId() != "" {
			t.Errorf("topic_id: want empty, got %q", draft.GetTopicId())
		}

		// DB: topic_id が NULL であることを検証
		var dbPost model.Post
		if err := env.db.First(&dbPost, "id = ?", draft.GetId()).Error; err != nil {
			t.Fatalf("failed to fetch post from DB: %v", err)
		}
		if dbPost.TopicID != nil {
			t.Errorf("DB topic_id: want nil, got %q", *dbPost.TopicID)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.postClient.CreateDraft(context.Background(), connect.NewRequest(&trendbirdv1.CreateDraftRequest{
			Content: "test",
		}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

// ---------------------------------------------------------------------------
// TestPostService_UpdateDraft
// ---------------------------------------------------------------------------

func TestPostService_UpdateDraft(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		post := seedPost(t, env.db, user.ID, withPostContent("original content"))

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		resp, err := client.UpdateDraft(context.Background(), connect.NewRequest(&trendbirdv1.UpdateDraftRequest{
			Id:      post.ID,
			Content: "updated content",
		}))
		if err != nil {
			t.Fatalf("UpdateDraft: %v", err)
		}

		draft := resp.Msg.GetDraft()
		if draft.GetContent() != "updated content" {
			t.Errorf("content: want %q, got %q", "updated content", draft.GetContent())
		}
		if draft.GetStatus() != trendbirdv1.PostStatus_POST_STATUS_DRAFT {
			t.Errorf("status: want POST_STATUS_DRAFT, got %v", draft.GetStatus())
		}

		// DB 確認
		var dbPost model.Post
		if err := env.db.First(&dbPost, "id = ?", post.ID).Error; err != nil {
			t.Fatalf("failed to fetch post from DB: %v", err)
		}
		if dbPost.Content != "updated content" {
			t.Errorf("DB content: want %q, got %q", "updated content", dbPost.Content)
		}
	})

	t.Run("not_found", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		_, err := client.UpdateDraft(context.Background(), connect.NewRequest(&trendbirdv1.UpdateDraftRequest{
			Id:      "00000000-0000-0000-0000-000000000000",
			Content: "should fail",
		}))
		assertConnectCode(t, err, connect.CodeNotFound)
	})

	t.Run("permission_denied", func(t *testing.T) {
		env := setupTest(t)
		owner := seedUser(t, env.db)
		other := seedUser(t, env.db)
		post := seedPost(t, env.db, owner.ID)

		client := connectClient(t, env, other.ID, trendbirdv1connect.NewPostServiceClient)
		_, err := client.UpdateDraft(context.Background(), connect.NewRequest(&trendbirdv1.UpdateDraftRequest{
			Id:      post.ID,
			Content: "hacked",
		}))
		assertConnectCode(t, err, connect.CodePermissionDenied)
	})

	t.Run("status_check_published", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		post := seedPost(t, env.db, user.ID, withPostStatus(3)) // Published

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		_, err := client.UpdateDraft(context.Background(), connect.NewRequest(&trendbirdv1.UpdateDraftRequest{
			Id:      post.ID,
			Content: "should fail",
		}))
		assertConnectCode(t, err, connect.CodeInvalidArgument)
	})

	t.Run("scheduled_post_can_be_edited", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		post := seedPost(t, env.db, user.ID, withPostStatus(2)) // Scheduled

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		resp, err := client.UpdateDraft(context.Background(), connect.NewRequest(&trendbirdv1.UpdateDraftRequest{
			Id:      post.ID,
			Content: "updated scheduled content",
		}))
		if err != nil {
			t.Fatalf("UpdateDraft on scheduled post: %v", err)
		}

		draft := resp.Msg.GetDraft()
		if draft.GetContent() != "updated scheduled content" {
			t.Errorf("content: want %q, got %q", "updated scheduled content", draft.GetContent())
		}

		// DB 確認
		var dbPost model.Post
		if err := env.db.First(&dbPost, "id = ?", post.ID).Error; err != nil {
			t.Fatalf("failed to fetch post from DB: %v", err)
		}
		if dbPost.Content != "updated scheduled content" {
			t.Errorf("DB content: want %q, got %q", "updated scheduled content", dbPost.Content)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.postClient.UpdateDraft(context.Background(), connect.NewRequest(&trendbirdv1.UpdateDraftRequest{
			Id:      "some-id",
			Content: "test",
		}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

// ---------------------------------------------------------------------------
// TestPostService_DeleteDraft
// ---------------------------------------------------------------------------

func TestPostService_DeleteDraft(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		post := seedPost(t, env.db, user.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		_, err := client.DeleteDraft(context.Background(), connect.NewRequest(&trendbirdv1.DeleteDraftRequest{
			Id: post.ID,
		}))
		if err != nil {
			t.Fatalf("DeleteDraft: %v", err)
		}

		// DB から削除されていることを検証
		var count int64
		env.db.Model(&model.Post{}).Where("id = ?", post.ID).Count(&count)
		if count != 0 {
			t.Errorf("expected post to be deleted, got count=%d", count)
		}
	})

	t.Run("not_found", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		_, err := client.DeleteDraft(context.Background(), connect.NewRequest(&trendbirdv1.DeleteDraftRequest{
			Id: "00000000-0000-0000-0000-000000000000",
		}))
		assertConnectCode(t, err, connect.CodeNotFound)
	})

	t.Run("permission_denied", func(t *testing.T) {
		env := setupTest(t)
		owner := seedUser(t, env.db)
		other := seedUser(t, env.db)
		post := seedPost(t, env.db, owner.ID)

		client := connectClient(t, env, other.ID, trendbirdv1connect.NewPostServiceClient)
		_, err := client.DeleteDraft(context.Background(), connect.NewRequest(&trendbirdv1.DeleteDraftRequest{
			Id: post.ID,
		}))
		assertConnectCode(t, err, connect.CodePermissionDenied)
	})

	t.Run("delete_scheduled", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		scheduledAt := time.Now().Add(24 * time.Hour)
		post := seedPost(t, env.db, user.ID, withPostStatus(2), withPostScheduledAt(scheduledAt))

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		_, err := client.DeleteDraft(context.Background(), connect.NewRequest(&trendbirdv1.DeleteDraftRequest{
			Id: post.ID,
		}))
		if err != nil {
			t.Fatalf("expected success but got error: %v", err)
		}

		// DB から削除されていることを確認
		var count int64
		env.db.Model(&model.Post{}).Where("id = ?", post.ID).Count(&count)
		if count != 0 {
			t.Errorf("post should be deleted from DB, but found %d records", count)
		}
	})

	t.Run("status_check", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		post := seedPost(t, env.db, user.ID, withPostStatus(3)) // Published

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		_, err := client.DeleteDraft(context.Background(), connect.NewRequest(&trendbirdv1.DeleteDraftRequest{
			Id: post.ID,
		}))
		assertConnectCode(t, err, connect.CodeInvalidArgument)
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.postClient.DeleteDraft(context.Background(), connect.NewRequest(&trendbirdv1.DeleteDraftRequest{
			Id: "some-id",
		}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

// ---------------------------------------------------------------------------
// TestPostService_SchedulePost
// ---------------------------------------------------------------------------

func TestPostService_SchedulePost(t *testing.T) {
	t.Run("normal_pro", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		post := seedPost(t, env.db, user.ID)

		futureTime := time.Now().Truncate(time.Hour).Add(25 * time.Hour).Format(time.RFC3339)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		resp, err := client.SchedulePost(context.Background(), connect.NewRequest(&trendbirdv1.SchedulePostRequest{
			Id:          post.ID,
			ScheduledAt: futureTime,
		}))
		if err != nil {
			t.Fatalf("SchedulePost: %v", err)
		}

		draft := resp.Msg.GetDraft()
		if draft.GetStatus() != trendbirdv1.PostStatus_POST_STATUS_SCHEDULED {
			t.Errorf("status: want POST_STATUS_SCHEDULED, got %v", draft.GetStatus())
		}
		if draft.GetScheduledAt() == "" {
			t.Error("expected scheduled_at to be set")
		}

		// DB 確認
		var dbPost model.Post
		if err := env.db.First(&dbPost, "id = ?", post.ID).Error; err != nil {
			t.Fatalf("failed to fetch post from DB: %v", err)
		}
		if dbPost.Status != 2 {
			t.Errorf("DB status: want 2 (SCHEDULED), got %d", dbPost.Status)
		}
		if dbPost.ScheduledAt == nil {
			t.Error("DB scheduled_at should be set")
		}
	})

	t.Run("permission_denied", func(t *testing.T) {
		env := setupTest(t)
		owner := seedUser(t, env.db)
		other := seedUser(t, env.db)
		post := seedPost(t, env.db, owner.ID)

		futureTime := time.Now().Truncate(time.Hour).Add(25 * time.Hour).Format(time.RFC3339)

		// other ユーザーが owner の投稿を SchedulePost → PermissionDenied
		client := connectClient(t, env, other.ID, trendbirdv1connect.NewPostServiceClient)
		_, err := client.SchedulePost(context.Background(), connect.NewRequest(&trendbirdv1.SchedulePostRequest{
			Id:          post.ID,
			ScheduledAt: futureTime,
		}))
		assertConnectCode(t, err, connect.CodePermissionDenied)
	})

	t.Run("status_check_published", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		post := seedPost(t, env.db, user.ID, withPostStatus(3)) // Published

		futureTime := time.Now().Truncate(time.Hour).Add(25 * time.Hour).Format(time.RFC3339)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		_, err := client.SchedulePost(context.Background(), connect.NewRequest(&trendbirdv1.SchedulePostRequest{
			Id:          post.ID,
			ScheduledAt: futureTime,
		}))
		assertConnectCode(t, err, connect.CodeInvalidArgument)
	})

	t.Run("reschedule_already_scheduled", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		scheduledAt := time.Now().Truncate(time.Hour).Add(25 * time.Hour)
		post := seedPost(t, env.db, user.ID, withPostStatus(2), withPostScheduledAt(scheduledAt)) // Scheduled

		newTime := time.Now().Truncate(time.Hour).Add(49 * time.Hour).Format(time.RFC3339)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		resp, err := client.SchedulePost(context.Background(), connect.NewRequest(&trendbirdv1.SchedulePostRequest{
			Id:          post.ID,
			ScheduledAt: newTime,
		}))
		if err != nil {
			t.Fatalf("expected success but got error: %v", err)
		}
		if resp.Msg.Draft == nil {
			t.Fatal("expected draft in response")
		}

		// DB 確認
		var dbPost model.Post
		if err := env.db.First(&dbPost, "id = ?", post.ID).Error; err != nil {
			t.Fatalf("failed to fetch post from DB: %v", err)
		}
		if dbPost.Status != 2 {
			t.Errorf("DB status: want 2 (SCHEDULED), got %d", dbPost.Status)
		}
		if dbPost.ScheduledAt == nil {
			t.Fatal("DB scheduled_at should be set")
		}
		newTimeParsed, _ := time.Parse(time.RFC3339, newTime)
		if !dbPost.ScheduledAt.Round(time.Second).Equal(newTimeParsed.Round(time.Second)) {
			t.Errorf("DB scheduled_at: want %v, got %v", newTimeParsed, *dbPost.ScheduledAt)
		}
	})

	t.Run("not_found", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		futureTime := time.Now().Truncate(time.Hour).Add(25 * time.Hour).Format(time.RFC3339)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		_, err := client.SchedulePost(context.Background(), connect.NewRequest(&trendbirdv1.SchedulePostRequest{
			Id:          "00000000-0000-0000-0000-000000000000",
			ScheduledAt: futureTime,
		}))
		assertConnectCode(t, err, connect.CodeNotFound)
	})

	t.Run("past_time", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		post := seedPost(t, env.db, user.ID)

		pastTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		_, err := client.SchedulePost(context.Background(), connect.NewRequest(&trendbirdv1.SchedulePostRequest{
			Id:          post.ID,
			ScheduledAt: pastTime,
		}))
		assertConnectCode(t, err, connect.CodeInvalidArgument)
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.postClient.SchedulePost(context.Background(), connect.NewRequest(&trendbirdv1.SchedulePostRequest{
			Id:          "some-id",
			ScheduledAt: "2099-01-01T00:00:00Z",
		}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

// ---------------------------------------------------------------------------
// TestPostService_PublishPost
// ---------------------------------------------------------------------------

func TestPostService_PublishPost(t *testing.T) {
	t.Run("publish", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		post := seedPost(t, env.db, user.ID, withPostContent("Lite user post"))
		seedTwitterConnection(t, env.db, user.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		resp, err := client.PublishPost(context.Background(), connect.NewRequest(&trendbirdv1.PublishPostRequest{
			Id: post.ID,
		}))
		if err != nil {
			t.Fatalf("PublishPost: %v", err)
		}

		published := resp.Msg.GetPost()
		if published.GetContent() != "Lite user post" {
			t.Errorf("content: want %q, got %q", "Lite user post", published.GetContent())
		}
		if published.GetPublishedAt() == "" {
			t.Error("expected published_at to be set")
		}
	})

	t.Run("normal", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		post := seedPost(t, env.db, user.ID, withPostContent("Hello from TrendBird!"))
		seedTwitterConnection(t, env.db, user.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		resp, err := client.PublishPost(context.Background(), connect.NewRequest(&trendbirdv1.PublishPostRequest{
			Id: post.ID,
		}))
		if err != nil {
			t.Fatalf("PublishPost: %v", err)
		}

		published := resp.Msg.GetPost()
		if published.GetContent() != "Hello from TrendBird!" {
			t.Errorf("content: want %q, got %q", "Hello from TrendBird!", published.GetContent())
		}
		if published.GetPublishedAt() == "" {
			t.Error("expected published_at to be set")
		}
		if published.GetTweetUrl() != "tweet-id-123" {
			t.Errorf("tweet_url: want %q, got %q", "tweet-id-123", published.GetTweetUrl())
		}

		// DB 確認
		var dbPost model.Post
		if err := env.db.First(&dbPost, "id = ?", post.ID).Error; err != nil {
			t.Fatalf("failed to fetch post from DB: %v", err)
		}
		if dbPost.Status != 3 {
			t.Errorf("DB status: want 3 (PUBLISHED), got %d", dbPost.Status)
		}
		if dbPost.TweetURL == nil || *dbPost.TweetURL != "tweet-id-123" {
			t.Errorf("DB tweet_url: want %q, got %v", "tweet-id-123", dbPost.TweetURL)
		}
		if dbPost.PublishedAt == nil {
			t.Error("DB published_at should be set")
		}
	})

	t.Run("from_scheduled", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		scheduledAt := time.Now().Add(24 * time.Hour)
		post := seedPost(t, env.db, user.ID,
			withPostStatus(2), // Scheduled
			withPostContent("Scheduled post content"),
			withPostScheduledAt(scheduledAt),
		)
		seedTwitterConnection(t, env.db, user.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		resp, err := client.PublishPost(context.Background(), connect.NewRequest(&trendbirdv1.PublishPostRequest{
			Id: post.ID,
		}))
		if err != nil {
			t.Fatalf("PublishPost from scheduled: %v", err)
		}

		published := resp.Msg.GetPost()
		if published.GetContent() != "Scheduled post content" {
			t.Errorf("content: want %q, got %q", "Scheduled post content", published.GetContent())
		}
		if published.GetPublishedAt() == "" {
			t.Error("expected published_at to be set")
		}

		// DB 確認
		var dbPost model.Post
		if err := env.db.First(&dbPost, "id = ?", post.ID).Error; err != nil {
			t.Fatalf("failed to fetch post from DB: %v", err)
		}
		if dbPost.Status != 3 {
			t.Errorf("DB status: want 3 (PUBLISHED), got %d", dbPost.Status)
		}
	})

	t.Run("x_api_failure", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		post := seedPost(t, env.db, user.ID)
		seedTwitterConnection(t, env.db, user.ID)

		// PostTweetFn をエラー返却に差し替え
		env.mockTwitter.PostTweetFn = func(_ context.Context, _ string, _ string) (string, error) {
			return "", fmt.Errorf("X API rate limit exceeded")
		}

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		_, err := client.PublishPost(context.Background(), connect.NewRequest(&trendbirdv1.PublishPostRequest{
			Id: post.ID,
		}))
		if err == nil {
			t.Fatal("expected error for X API failure, got nil")
		}

		// DB 上で status=FAILED + error_message が設定されていることを検証
		var dbPost model.Post
		if err := env.db.First(&dbPost, "id = ?", post.ID).Error; err != nil {
			t.Fatalf("failed to fetch post from DB: %v", err)
		}
		if dbPost.Status != 4 {
			t.Errorf("DB status: want 4 (FAILED), got %d", dbPost.Status)
		}
		if dbPost.ErrorMessage == nil || *dbPost.ErrorMessage == "" {
			t.Error("DB error_message should be set")
		}
		if dbPost.FailedAt == nil {
			t.Error("DB failed_at should be set")
		}
	})

	t.Run("x_api_error_types", func(t *testing.T) {
		cases := []struct {
			name    string
			errMsg  string
			wantSub string // error_message に含まれるべき部分文字列
		}{
			{"timeout", "context deadline exceeded: X API request timed out", "context deadline exceeded"},
			{"rate_limit_429", "X API error: 429 Too Many Requests", "429 Too Many Requests"},
			{"auth_error_401", "X API error: 401 Unauthorized", "401 Unauthorized"},
			{"server_error_500", "X API error: 500 Internal Server Error", "500 Internal Server Error"},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				env := setupTest(t)
				user := seedUser(t, env.db)
				post := seedPost(t, env.db, user.ID)
				seedTwitterConnection(t, env.db, user.ID)

				env.mockTwitter.PostTweetFn = func(_ context.Context, _ string, _ string) (string, error) {
					return "", fmt.Errorf("%s", tc.errMsg)
				}

				client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
				_, err := client.PublishPost(context.Background(), connect.NewRequest(&trendbirdv1.PublishPostRequest{
					Id: post.ID,
				}))
				if err == nil {
					t.Fatal("expected error for X API failure, got nil")
				}

				// DB: status=FAILED, failed_at 設定, error_message に固有文字列が含まれる
				var dbPost model.Post
				if err := env.db.First(&dbPost, "id = ?", post.ID).Error; err != nil {
					t.Fatalf("failed to fetch post from DB: %v", err)
				}
				if dbPost.Status != 4 {
					t.Errorf("DB status: want 4 (FAILED), got %d", dbPost.Status)
				}
				if dbPost.FailedAt == nil {
					t.Error("DB failed_at should be set")
				}
				if dbPost.ErrorMessage == nil {
					t.Fatal("DB error_message should be set")
				}
				if !strings.Contains(*dbPost.ErrorMessage, tc.wantSub) {
					t.Errorf("DB error_message: want substring %q, got %q", tc.wantSub, *dbPost.ErrorMessage)
				}
			})
		}
	})

	t.Run("failed_post_error_message_in_api_response", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		post := seedPost(t, env.db, user.ID)
		seedTwitterConnection(t, env.db, user.ID)

		env.mockTwitter.PostTweetFn = func(_ context.Context, _ string, _ string) (string, error) {
			return "", fmt.Errorf("X API error: 429 Too Many Requests")
		}

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		_, _ = client.PublishPost(context.Background(), connect.NewRequest(&trendbirdv1.PublishPostRequest{
			Id: post.ID,
		}))

		// ListDrafts で FAILED 投稿を取得し、API レスポンスの error_message/failed_at を検証
		resp, err := client.ListDrafts(context.Background(), connect.NewRequest(&trendbirdv1.ListDraftsRequest{}))
		if err != nil {
			t.Fatalf("ListDrafts: %v", err)
		}

		drafts := resp.Msg.GetDrafts()
		if got := len(drafts); got != 1 {
			t.Fatalf("expected 1 draft (FAILED), got %d", got)
		}

		d := drafts[0]
		if d.GetStatus() != trendbirdv1.PostStatus_POST_STATUS_FAILED {
			t.Errorf("status: want FAILED, got %v", d.GetStatus())
		}
		if d.ErrorMessage == nil || !strings.Contains(*d.ErrorMessage, "429 Too Many Requests") {
			t.Errorf("error_message: want substring '429 Too Many Requests', got %v", d.ErrorMessage)
		}
		if d.FailedAt == nil || *d.FailedAt == "" {
			t.Error("failed_at should be set in API response")
		}

		stats := resp.Msg.GetStats()
		if stats.GetTotalFailed() != 1 {
			t.Errorf("total_failed: want 1, got %d", stats.GetTotalFailed())
		}
	})

	t.Run("token_refresh", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		post := seedPost(t, env.db, user.ID)

		// token_expires_at が過去の Twitter 接続を seed
		seedTwitterConnection(t, env.db, user.ID,
			withTokenExpiresAt(time.Now().Add(-1*time.Hour)),
		)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		_, err := client.PublishPost(context.Background(), connect.NewRequest(&trendbirdv1.PublishPostRequest{
			Id: post.ID,
		}))
		if err != nil {
			t.Fatalf("PublishPost with token refresh: %v", err)
		}

		// DB 上のアクセストークンが更新されていることを検証
		var dbConn model.TwitterConnection
		if err := env.db.First(&dbConn, "user_id = ?", user.ID).Error; err != nil {
			t.Fatalf("failed to fetch twitter_connection from DB: %v", err)
		}
		if dbConn.AccessToken != "refreshed-access-token" {
			t.Errorf("access_token: want %q, got %q", "refreshed-access-token", dbConn.AccessToken)
		}
		if dbConn.RefreshToken != "refreshed-refresh-token" {
			t.Errorf("refresh_token: want %q, got %q", "refreshed-refresh-token", dbConn.RefreshToken)
		}
	})

	t.Run("no_twitter_connection", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		post := seedPost(t, env.db, user.ID)
		// Twitter 接続は seed しない

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		_, err := client.PublishPost(context.Background(), connect.NewRequest(&trendbirdv1.PublishPostRequest{
			Id: post.ID,
		}))
		assertConnectCode(t, err, connect.CodePermissionDenied)
	})

	t.Run("status_check_published", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		now := time.Now()
		post := seedPost(t, env.db, user.ID, withPostStatus(3), withPostPublishedAt(now)) // Published
		seedTwitterConnection(t, env.db, user.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		_, err := client.PublishPost(context.Background(), connect.NewRequest(&trendbirdv1.PublishPostRequest{
			Id: post.ID,
		}))
		assertConnectCode(t, err, connect.CodeInvalidArgument)
	})

	t.Run("status_check_failed", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		post := seedPost(t, env.db, user.ID, withPostStatus(4)) // Failed
		seedTwitterConnection(t, env.db, user.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		_, err := client.PublishPost(context.Background(), connect.NewRequest(&trendbirdv1.PublishPostRequest{
			Id: post.ID,
		}))
		assertConnectCode(t, err, connect.CodeInvalidArgument)
	})

	t.Run("not_found", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		seedTwitterConnection(t, env.db, user.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		_, err := client.PublishPost(context.Background(), connect.NewRequest(&trendbirdv1.PublishPostRequest{
			Id: "00000000-0000-0000-0000-000000000000",
		}))
		assertConnectCode(t, err, connect.CodeNotFound)
	})

	t.Run("cross_user_isolation", func(t *testing.T) {
		env := setupTest(t)
		owner := seedUser(t, env.db)
		other := seedUser(t, env.db)
		post := seedPost(t, env.db, owner.ID, withPostContent("Owner's post"))
		seedTwitterConnection(t, env.db, other.ID)

		// other が owner の投稿を PublishPost → PermissionDenied
		client := connectClient(t, env, other.ID, trendbirdv1connect.NewPostServiceClient)
		_, err := client.PublishPost(context.Background(), connect.NewRequest(&trendbirdv1.PublishPostRequest{
			Id: post.ID,
		}))
		assertConnectCode(t, err, connect.CodePermissionDenied)
	})

	t.Run("token_refresh_failure", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		post := seedPost(t, env.db, user.ID)

		// token_expires_at が過去の Twitter 接続を seed
		seedTwitterConnection(t, env.db, user.ID,
			withTokenExpiresAt(time.Now().Add(-1*time.Hour)),
		)

		// RefreshTokenFn をエラー返却に差し替え
		env.mockTwitter.RefreshTokenFn = func(_ context.Context, _ string) (*gateway.OAuthTokenResponse, error) {
			return nil, fmt.Errorf("refresh token expired")
		}

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		_, err := client.PublishPost(context.Background(), connect.NewRequest(&trendbirdv1.PublishPostRequest{
			Id: post.ID,
		}))
		assertConnectCode(t, err, connect.CodeInternal)

		// DB: twitter_connection の status が Error になっていることを検証
		var dbConn model.TwitterConnection
		if err := env.db.First(&dbConn, "user_id = ?", user.ID).Error; err != nil {
			t.Fatalf("failed to fetch twitter_connection from DB: %v", err)
		}
		if dbConn.Status != 4 { // Error
			t.Errorf("DB status: want 4 (Error), got %d", dbConn.Status)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.postClient.PublishPost(context.Background(), connect.NewRequest(&trendbirdv1.PublishPostRequest{
			Id: "some-id",
		}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

// ---------------------------------------------------------------------------
// TestPostService_ListDrafts
// ---------------------------------------------------------------------------

func TestPostService_ListDrafts(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		// Draft 3件 + Published 1件
		seedPost(t, env.db, user.ID, withPostContent("Draft 1"))
		seedPost(t, env.db, user.ID, withPostContent("Draft 2"))
		seedPost(t, env.db, user.ID, withPostContent("Draft 3"))
		now := time.Now()
		seedPost(t, env.db, user.ID, withPostStatus(3), withPostContent("Published"), withPostPublishedAt(now))

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		resp, err := client.ListDrafts(context.Background(), connect.NewRequest(&trendbirdv1.ListDraftsRequest{}))
		if err != nil {
			t.Fatalf("ListDrafts: %v", err)
		}

		drafts := resp.Msg.GetDrafts()
		if got := len(drafts); got != 3 {
			t.Fatalf("expected 3 drafts, got %d", got)
		}

		for _, d := range drafts {
			if d.GetStatus() != trendbirdv1.PostStatus_POST_STATUS_DRAFT {
				t.Errorf("draft status: want POST_STATUS_DRAFT, got %v", d.GetStatus())
			}
		}

		// stats の検証
		stats := resp.Msg.GetStats()
		if stats == nil {
			t.Fatal("expected stats to be present")
		}
		if stats.GetTotalDrafts() != 3 {
			t.Errorf("total_drafts: want 3, got %d", stats.GetTotalDrafts())
		}
		if stats.GetTotalPublished() != 1 {
			t.Errorf("total_published: want 1, got %d", stats.GetTotalPublished())
		}
	})

	t.Run("empty", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		resp, err := client.ListDrafts(context.Background(), connect.NewRequest(&trendbirdv1.ListDraftsRequest{}))
		if err != nil {
			t.Fatalf("ListDrafts: %v", err)
		}

		if got := len(resp.Msg.GetDrafts()); got != 0 {
			t.Errorf("expected 0 drafts, got %d", got)
		}
	})

	t.Run("includes_scheduled", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		seedPost(t, env.db, user.ID, withPostContent("Draft 1"))
		scheduledAt := time.Now().Add(24 * time.Hour)
		seedPost(t, env.db, user.ID, withPostStatus(2), withPostContent("Scheduled 1"), withPostScheduledAt(scheduledAt))

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		resp, err := client.ListDrafts(context.Background(), connect.NewRequest(&trendbirdv1.ListDraftsRequest{}))
		if err != nil {
			t.Fatalf("ListDrafts: %v", err)
		}

		drafts := resp.Msg.GetDrafts()
		// ListDrafts は Draft + Scheduled を返すことを検証
		if got := len(drafts); got != 2 {
			t.Fatalf("expected 2 drafts (draft+scheduled), got %d", got)
		}

		statusMap := make(map[trendbirdv1.PostStatus]int)
		for _, d := range drafts {
			statusMap[d.GetStatus()]++
		}
		if statusMap[trendbirdv1.PostStatus_POST_STATUS_DRAFT] != 1 {
			t.Errorf("expected 1 draft, got %d", statusMap[trendbirdv1.PostStatus_POST_STATUS_DRAFT])
		}
		if statusMap[trendbirdv1.PostStatus_POST_STATUS_SCHEDULED] != 1 {
			t.Errorf("expected 1 scheduled, got %d", statusMap[trendbirdv1.PostStatus_POST_STATUS_SCHEDULED])
		}
	})

	t.Run("cross_user_isolation", func(t *testing.T) {
		env := setupTest(t)
		userA := seedUser(t, env.db)
		userB := seedUser(t, env.db)

		// userA に2件、userB に1件
		seedPost(t, env.db, userA.ID, withPostContent("A Draft 1"))
		seedPost(t, env.db, userA.ID, withPostContent("A Draft 2"))
		seedPost(t, env.db, userB.ID, withPostContent("B Draft 1"))

		// userB が ListDrafts → 自分の1件のみ
		client := connectClient(t, env, userB.ID, trendbirdv1connect.NewPostServiceClient)
		resp, err := client.ListDrafts(context.Background(), connect.NewRequest(&trendbirdv1.ListDraftsRequest{}))
		if err != nil {
			t.Fatalf("ListDrafts: %v", err)
		}

		drafts := resp.Msg.GetDrafts()
		if got := len(drafts); got != 1 {
			t.Fatalf("expected 1 draft for userB, got %d", got)
		}

		stats := resp.Msg.GetStats()
		if stats.GetTotalDrafts() != 1 {
			t.Errorf("total_drafts: want 1, got %d", stats.GetTotalDrafts())
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.postClient.ListDrafts(context.Background(), connect.NewRequest(&trendbirdv1.ListDraftsRequest{}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

// ---------------------------------------------------------------------------
// TestPostService_ListPostHistory
// ---------------------------------------------------------------------------

func TestPostService_ListPostHistory(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		now := time.Now()
		// Published 2件 + Draft 1件
		seedPost(t, env.db, user.ID, withPostStatus(3), withPostPublishedAt(now), withPostContent("Published 1"))
		seedPost(t, env.db, user.ID, withPostStatus(3), withPostPublishedAt(now.Add(-1*time.Hour)), withPostContent("Published 2"))
		seedPost(t, env.db, user.ID, withPostContent("Draft"))

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		resp, err := client.ListPostHistory(context.Background(), connect.NewRequest(&trendbirdv1.ListPostHistoryRequest{}))
		if err != nil {
			t.Fatalf("ListPostHistory: %v", err)
		}

		posts := resp.Msg.GetPosts()
		if got := len(posts); got != 2 {
			t.Fatalf("expected 2 published posts, got %d", got)
		}

		for _, p := range posts {
			if p.GetPublishedAt() == "" {
				t.Error("published post should have published_at set")
			}
		}
	})

	t.Run("empty", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		resp, err := client.ListPostHistory(context.Background(), connect.NewRequest(&trendbirdv1.ListPostHistoryRequest{}))
		if err != nil {
			t.Fatalf("ListPostHistory: %v", err)
		}

		if got := len(resp.Msg.GetPosts()); got != 0 {
			t.Errorf("expected 0 posts, got %d", got)
		}
	})

	t.Run("only_published_posts", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		now := time.Now()
		// Published 1件 + Draft 1件 + Scheduled 1件 + Failed 1件
		seedPost(t, env.db, user.ID, withPostStatus(3), withPostPublishedAt(now), withPostContent("Published"))
		seedPost(t, env.db, user.ID, withPostContent("Draft"))
		seedPost(t, env.db, user.ID, withPostStatus(2), withPostScheduledAt(now.Add(24*time.Hour)))
		seedPost(t, env.db, user.ID, withPostStatus(4))

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		resp, err := client.ListPostHistory(context.Background(), connect.NewRequest(&trendbirdv1.ListPostHistoryRequest{}))
		if err != nil {
			t.Fatalf("ListPostHistory: %v", err)
		}

		posts := resp.Msg.GetPosts()
		if got := len(posts); got != 1 {
			t.Fatalf("expected 1 published post only, got %d", got)
		}
		if posts[0].GetContent() != "Published" {
			t.Errorf("content: want Published, got %s", posts[0].GetContent())
		}
	})

	t.Run("cross_user_isolation", func(t *testing.T) {
		env := setupTest(t)
		userA := seedUser(t, env.db)
		userB := seedUser(t, env.db)

		now := time.Now()
		// userA に Published 2件、userB に Published 1件
		seedPost(t, env.db, userA.ID, withPostStatus(3), withPostPublishedAt(now), withPostContent("A Published 1"))
		seedPost(t, env.db, userA.ID, withPostStatus(3), withPostPublishedAt(now), withPostContent("A Published 2"))
		seedPost(t, env.db, userB.ID, withPostStatus(3), withPostPublishedAt(now), withPostContent("B Published 1"))

		// userB が ListPostHistory → 自分の1件のみ
		client := connectClient(t, env, userB.ID, trendbirdv1connect.NewPostServiceClient)
		resp, err := client.ListPostHistory(context.Background(), connect.NewRequest(&trendbirdv1.ListPostHistoryRequest{}))
		if err != nil {
			t.Fatalf("ListPostHistory: %v", err)
		}

		posts := resp.Msg.GetPosts()
		if got := len(posts); got != 1 {
			t.Fatalf("expected 1 post for userB, got %d", got)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.postClient.ListPostHistory(context.Background(), connect.NewRequest(&trendbirdv1.ListPostHistoryRequest{}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

// ---------------------------------------------------------------------------
// TestPostService_GetPostStats
// ---------------------------------------------------------------------------

func TestPostService_GetPostStats(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		now := time.Now()

		// Draft 2件
		seedPost(t, env.db, user.ID, withPostContent("Draft 1"))
		seedPost(t, env.db, user.ID, withPostContent("Draft 2"))

		// Scheduled 1件
		seedPost(t, env.db, user.ID, withPostStatus(2), withPostScheduledAt(now.Add(24*time.Hour)))

		// Published 3件（うち今月2件）
		seedPost(t, env.db, user.ID, withPostStatus(3), withPostPublishedAt(now))
		seedPost(t, env.db, user.ID, withPostStatus(3), withPostPublishedAt(now.Add(-1*time.Hour)))
		// 先月分
		lastMonth := time.Date(now.Year(), now.Month()-1, 15, 12, 0, 0, 0, time.UTC)
		seedPost(t, env.db, user.ID, withPostStatus(3), withPostPublishedAt(lastMonth))

		// Failed 1件
		seedPost(t, env.db, user.ID, withPostStatus(4))

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		resp, err := client.GetPostStats(context.Background(), connect.NewRequest(&trendbirdv1.GetPostStatsRequest{}))
		if err != nil {
			t.Fatalf("GetPostStats: %v", err)
		}

		stats := resp.Msg.GetStats()
		if stats.GetTotalDrafts() != 2 {
			t.Errorf("total_drafts: want 2, got %d", stats.GetTotalDrafts())
		}
		if stats.GetTotalScheduled() != 1 {
			t.Errorf("total_scheduled: want 1, got %d", stats.GetTotalScheduled())
		}
		if stats.GetTotalPublished() != 3 {
			t.Errorf("total_published: want 3, got %d", stats.GetTotalPublished())
		}
		if stats.GetTotalFailed() != 1 {
			t.Errorf("total_failed: want 1, got %d", stats.GetTotalFailed())
		}
		if stats.GetThisMonthPublished() != 2 {
			t.Errorf("this_month_published: want 2, got %d", stats.GetThisMonthPublished())
		}
	})

	t.Run("empty", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewPostServiceClient)
		resp, err := client.GetPostStats(context.Background(), connect.NewRequest(&trendbirdv1.GetPostStatsRequest{}))
		if err != nil {
			t.Fatalf("GetPostStats: %v", err)
		}

		stats := resp.Msg.GetStats()
		if stats.GetTotalDrafts() != 0 {
			t.Errorf("total_drafts: want 0, got %d", stats.GetTotalDrafts())
		}
		if stats.GetTotalScheduled() != 0 {
			t.Errorf("total_scheduled: want 0, got %d", stats.GetTotalScheduled())
		}
		if stats.GetTotalPublished() != 0 {
			t.Errorf("total_published: want 0, got %d", stats.GetTotalPublished())
		}
		if stats.GetTotalFailed() != 0 {
			t.Errorf("total_failed: want 0, got %d", stats.GetTotalFailed())
		}
		if stats.GetThisMonthPublished() != 0 {
			t.Errorf("this_month_published: want 0, got %d", stats.GetThisMonthPublished())
		}
	})

	t.Run("cross_user_isolation", func(t *testing.T) {
		env := setupTest(t)
		userA := seedUser(t, env.db)
		userB := seedUser(t, env.db)

		now := time.Now()
		// userA: Draft 3件 + Published 2件
		seedPost(t, env.db, userA.ID, withPostContent("A Draft 1"))
		seedPost(t, env.db, userA.ID, withPostContent("A Draft 2"))
		seedPost(t, env.db, userA.ID, withPostContent("A Draft 3"))
		seedPost(t, env.db, userA.ID, withPostStatus(3), withPostPublishedAt(now))
		seedPost(t, env.db, userA.ID, withPostStatus(3), withPostPublishedAt(now))

		// userB: Draft 1件 + Published 1件
		seedPost(t, env.db, userB.ID, withPostContent("B Draft 1"))
		seedPost(t, env.db, userB.ID, withPostStatus(3), withPostPublishedAt(now))

		// userB が GetPostStats → 自分の分のみ
		client := connectClient(t, env, userB.ID, trendbirdv1connect.NewPostServiceClient)
		resp, err := client.GetPostStats(context.Background(), connect.NewRequest(&trendbirdv1.GetPostStatsRequest{}))
		if err != nil {
			t.Fatalf("GetPostStats: %v", err)
		}

		stats := resp.Msg.GetStats()
		if stats.GetTotalDrafts() != 1 {
			t.Errorf("total_drafts: want 1, got %d", stats.GetTotalDrafts())
		}
		if stats.GetTotalPublished() != 1 {
			t.Errorf("total_published: want 1, got %d", stats.GetTotalPublished())
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.postClient.GetPostStats(context.Background(), connect.NewRequest(&trendbirdv1.GetPostStatsRequest{}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

// ---------------------------------------------------------------------------
// TestPostService_Unauthenticated (D2)
// ---------------------------------------------------------------------------

func TestPostService_Unauthenticated(t *testing.T) {
	env := setupTest(t)

	t.Run("GeneratePosts", func(t *testing.T) {
		_, err := env.postClient.GeneratePosts(context.Background(), connect.NewRequest(&trendbirdv1.GeneratePostsRequest{
			TopicId: "some-id",
		}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})

	t.Run("CreateDraft", func(t *testing.T) {
		_, err := env.postClient.CreateDraft(context.Background(), connect.NewRequest(&trendbirdv1.CreateDraftRequest{
			Content: "test",
		}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})

	t.Run("UpdateDraft", func(t *testing.T) {
		_, err := env.postClient.UpdateDraft(context.Background(), connect.NewRequest(&trendbirdv1.UpdateDraftRequest{
			Id:      "some-id",
			Content: "test",
		}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})

	t.Run("DeleteDraft", func(t *testing.T) {
		_, err := env.postClient.DeleteDraft(context.Background(), connect.NewRequest(&trendbirdv1.DeleteDraftRequest{
			Id: "some-id",
		}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})

	t.Run("SchedulePost", func(t *testing.T) {
		_, err := env.postClient.SchedulePost(context.Background(), connect.NewRequest(&trendbirdv1.SchedulePostRequest{
			Id:          "some-id",
			ScheduledAt: "2099-01-01T00:00:00Z",
		}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})

	t.Run("PublishPost", func(t *testing.T) {
		_, err := env.postClient.PublishPost(context.Background(), connect.NewRequest(&trendbirdv1.PublishPostRequest{
			Id: "some-id",
		}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})

	t.Run("ListDrafts", func(t *testing.T) {
		_, err := env.postClient.ListDrafts(context.Background(), connect.NewRequest(&trendbirdv1.ListDraftsRequest{}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})

	t.Run("ListPostHistory", func(t *testing.T) {
		_, err := env.postClient.ListPostHistory(context.Background(), connect.NewRequest(&trendbirdv1.ListPostHistoryRequest{}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})

	t.Run("GetPostStats", func(t *testing.T) {
		_, err := env.postClient.GetPostStats(context.Background(), connect.NewRequest(&trendbirdv1.GetPostStatsRequest{}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}
