package e2etest

import (
	"context"
	"testing"

	"connectrpc.com/connect"

	trendbirdv1 "github.com/trendbird/backend/gen/trendbird/v1"
	"github.com/trendbird/backend/gen/trendbird/v1/trendbirdv1connect"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

func TestTopicService_AddGenre(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		genreID := ensureGenre(t, env.db, "technology")

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		_, err := client.AddGenre(context.Background(), connect.NewRequest(&trendbirdv1.AddGenreRequest{
			Genre: "technology",
		}))
		if err != nil {
			t.Fatalf("AddGenre: %v", err)
		}

		// DB に保存されていることを確認
		var count int64
		env.db.Model(&model.UserGenre{}).Where("user_id = ? AND genre_id = ?", user.ID, genreID).Count(&count)
		if count != 1 {
			t.Errorf("expected 1 user_genre, got %d", count)
		}
	})

	t.Run("idempotent", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		genreID := ensureGenre(t, env.db, "marketing")

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)

		// 1回目
		_, err := client.AddGenre(context.Background(), connect.NewRequest(&trendbirdv1.AddGenreRequest{
			Genre: "marketing",
		}))
		if err != nil {
			t.Fatalf("AddGenre (1st): %v", err)
		}

		// 2回目 — 同じジャンルでも成功すること
		_, err = client.AddGenre(context.Background(), connect.NewRequest(&trendbirdv1.AddGenreRequest{
			Genre: "marketing",
		}))
		if err != nil {
			t.Fatalf("AddGenre (2nd): %v", err)
		}

		// レコードが1件のみであること
		var count int64
		env.db.Model(&model.UserGenre{}).Where("user_id = ? AND genre_id = ?", user.ID, genreID).Count(&count)
		if count != 1 {
			t.Errorf("expected 1 user_genre after idempotent add, got %d", count)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.topicClient.AddGenre(context.Background(), connect.NewRequest(&trendbirdv1.AddGenreRequest{
			Genre: "technology",
		}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})

	t.Run("nonexistent_genre_slug", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		_, err := client.AddGenre(context.Background(), connect.NewRequest(&trendbirdv1.AddGenreRequest{
			Genre: "nonexistent-slug",
		}))
		assertConnectCode(t, err, connect.CodeNotFound)
	})
}

func TestTopicService_RemoveGenre(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		seedUserGenre(t, env.db, user.ID, "technology")
		genreID := ensureGenre(t, env.db, "technology")

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		_, err := client.RemoveGenre(context.Background(), connect.NewRequest(&trendbirdv1.RemoveGenreRequest{
			Genre: "technology",
		}))
		if err != nil {
			t.Fatalf("RemoveGenre: %v", err)
		}

		// DB から削除されていること
		var count int64
		env.db.Model(&model.UserGenre{}).Where("user_id = ? AND genre_id = ?", user.ID, genreID).Count(&count)
		if count != 0 {
			t.Errorf("expected user_genre to be deleted, got count=%d", count)
		}
	})

	t.Run("cascade_delete_topics", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		seedUserGenre(t, env.db, user.ID, "technology")
		tpA := seedTopic(t, env.db, withTopicGenre("technology"), withTopicName("Topic A"))
		seedUserTopic(t, env.db, user.ID, tpA.ID)
		tpB := seedTopic(t, env.db, withTopicGenre("technology"), withTopicName("Topic B"))
		seedUserTopic(t, env.db, user.ID, tpB.ID)

		// 別ジャンルのトピックは残る
		seedUserGenre(t, env.db, user.ID, "marketing")
		tpE := seedTopic(t, env.db, withTopicGenre("marketing"), withTopicName("Marketing Topic"))
		seedUserTopic(t, env.db, user.ID, tpE.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		_, err := client.RemoveGenre(context.Background(), connect.NewRequest(&trendbirdv1.RemoveGenreRequest{
			Genre: "technology",
		}))
		if err != nil {
			t.Fatalf("RemoveGenre: %v", err)
		}

		// technology の user_topics リンクが全削除されていること
		aiGenreID := ensureGenre(t, env.db, "technology")
		var aiCount int64
		env.db.Model(&model.UserTopic{}).
			Joins("JOIN topics ON user_topics.topic_id = topics.id").
			Where("user_topics.user_id = ? AND topics.genre_id = ?", user.ID, aiGenreID).
			Count(&aiCount)
		if aiCount != 0 {
			t.Errorf("expected technology user_topics deleted, got %d", aiCount)
		}

		// marketing の user_topics リンクは残っていること
		engGenreID := ensureGenre(t, env.db, "marketing")
		var engCount int64
		env.db.Model(&model.UserTopic{}).
			Joins("JOIN topics ON user_topics.topic_id = topics.id").
			Where("user_topics.user_id = ? AND topics.genre_id = ?", user.ID, engGenreID).
			Count(&engCount)
		if engCount != 1 {
			t.Errorf("expected 1 marketing user_topic, got %d", engCount)
		}
	})

	t.Run("nonexistent_genre", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		ensureGenre(t, env.db, "technology")

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		// 存在しないジャンルの削除は成功する（冪等）
		_, err := client.RemoveGenre(context.Background(), connect.NewRequest(&trendbirdv1.RemoveGenreRequest{
			Genre: "technology",
		}))
		if err != nil {
			t.Fatalf("RemoveGenre (nonexistent): %v", err)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.topicClient.RemoveGenre(context.Background(), connect.NewRequest(&trendbirdv1.RemoveGenreRequest{
			Genre: "technology",
		}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

func TestTopicService_ListUserGenres(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		seedUserGenre(t, env.db, user.ID, "technology")
		seedUserGenre(t, env.db, user.ID, "marketing")

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		resp, err := client.ListUserGenres(context.Background(), connect.NewRequest(&trendbirdv1.ListUserGenresRequest{}))
		if err != nil {
			t.Fatalf("ListUserGenres: %v", err)
		}

		genres := resp.Msg.GetGenres()
		if len(genres) != 2 {
			t.Fatalf("expected 2 genres, got %d", len(genres))
		}
		// created_at ASC order
		if genres[0] != "technology" {
			t.Errorf("genres[0]: want technology, got %s", genres[0])
		}
		if genres[1] != "marketing" {
			t.Errorf("genres[1]: want marketing, got %s", genres[1])
		}
	})

	t.Run("empty", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
		resp, err := client.ListUserGenres(context.Background(), connect.NewRequest(&trendbirdv1.ListUserGenresRequest{}))
		if err != nil {
			t.Fatalf("ListUserGenres: %v", err)
		}

		if got := len(resp.Msg.GetGenres()); got != 0 {
			t.Errorf("expected 0 genres, got %d", got)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.topicClient.ListUserGenres(context.Background(), connect.NewRequest(&trendbirdv1.ListUserGenresRequest{}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

// ---------------------------------------------------------------------------
// Cascade Advanced Tests (7a-7c)
// ---------------------------------------------------------------------------

func TestTopicService_RemoveGenre_CascadeAdvanced(t *testing.T) {
	t.Run("cascade_multi_topics_preserves_other_genre", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		ensureGenre(t, env.db, "technology")
		ensureGenre(t, env.db, "marketing")

		seedUserGenre(t, env.db, user.ID, "technology")
		seedUserGenre(t, env.db, user.ID, "marketing")

		// 2 topics in technology
		techA := seedTopic(t, env.db, withTopicGenre("technology"), withTopicName("Tech A"))
		seedUserTopic(t, env.db, user.ID, techA.ID)
		techB := seedTopic(t, env.db, withTopicGenre("technology"), withTopicName("Tech B"))
		seedUserTopic(t, env.db, user.ID, techB.ID)

		// 1 topic in marketing
		mktA := seedTopic(t, env.db, withTopicGenre("marketing"), withTopicName("Mkt A"))
		seedUserTopic(t, env.db, user.ID, mktA.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)

		// Remove technology genre (cascade)
		_, err := client.RemoveGenre(context.Background(), connect.NewRequest(&trendbirdv1.RemoveGenreRequest{
			Genre: "technology",
		}))
		if err != nil {
			t.Fatalf("RemoveGenre technology: %v", err)
		}

		// ListTopics → only marketing topic remains
		listResp, err := client.ListTopics(context.Background(), connect.NewRequest(&trendbirdv1.ListTopicsRequest{}))
		if err != nil {
			t.Fatalf("ListTopics: %v", err)
		}
		topics := listResp.Msg.GetTopics()
		if len(topics) != 1 {
			t.Fatalf("expected 1 topic, got %d", len(topics))
		}
		if topics[0].GetName() != "Mkt A" {
			t.Errorf("expected Mkt A, got %s", topics[0].GetName())
		}

		// ListUserGenres → only marketing remains
		genreResp, err := client.ListUserGenres(context.Background(), connect.NewRequest(&trendbirdv1.ListUserGenresRequest{}))
		if err != nil {
			t.Fatalf("ListUserGenres: %v", err)
		}
		genres := genreResp.Msg.GetGenres()
		if len(genres) != 1 || genres[0] != "marketing" {
			t.Errorf("expected [marketing], got %v", genres)
		}
	})

	t.Run("cascade_genre_with_no_topics", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		seedUserGenre(t, env.db, user.ID, "technology")

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)

		// Remove genre that has no topics
		_, err := client.RemoveGenre(context.Background(), connect.NewRequest(&trendbirdv1.RemoveGenreRequest{
			Genre: "technology",
		}))
		if err != nil {
			t.Fatalf("RemoveGenre: %v", err)
		}

		// ListUserGenres → empty
		genreResp, err := client.ListUserGenres(context.Background(), connect.NewRequest(&trendbirdv1.ListUserGenresRequest{}))
		if err != nil {
			t.Fatalf("ListUserGenres: %v", err)
		}
		if got := len(genreResp.Msg.GetGenres()); got != 0 {
			t.Errorf("expected 0 genres, got %d", got)
		}
	})

	t.Run("cascade_shared_topic_other_user_unaffected", func(t *testing.T) {
		env := setupTest(t)
		ensureGenre(t, env.db, "technology")

		userA := seedUser(t, env.db)
		userB := seedUser(t, env.db)

		seedUserGenre(t, env.db, userA.ID, "technology")
		seedUserGenre(t, env.db, userB.ID, "technology")

		// Shared topic
		tp := seedTopic(t, env.db, withTopicGenre("technology"), withTopicName("Shared Tech"))
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

		// userA: 0 topics
		listA, err := clientA.ListTopics(context.Background(), connect.NewRequest(&trendbirdv1.ListTopicsRequest{}))
		if err != nil {
			t.Fatalf("ListTopics userA: %v", err)
		}
		if got := len(listA.Msg.GetTopics()); got != 0 {
			t.Errorf("userA: expected 0 topics, got %d", got)
		}

		// userB: 1 topic
		listB, err := clientB.ListTopics(context.Background(), connect.NewRequest(&trendbirdv1.ListTopicsRequest{}))
		if err != nil {
			t.Fatalf("ListTopics userB: %v", err)
		}
		if got := len(listB.Msg.GetTopics()); got != 1 {
			t.Errorf("userB: expected 1 topic, got %d", got)
		}

		// topics record remains
		var topicCount int64
		env.db.Model(&model.Topic{}).Where("id = ?", tp.ID).Count(&topicCount)
		if topicCount != 1 {
			t.Errorf("expected topic record to remain, got count=%d", topicCount)
		}
	})
}

// ---------------------------------------------------------------------------
// TestTopicService_AddGenre_Concurrent
// ---------------------------------------------------------------------------

func TestTopicService_AddGenre_Concurrent(t *testing.T) {
	env := setupTest(t)
	user := seedUser(t, env.db)
	genreID := ensureGenre(t, env.db, "technology")

	type result struct {
		err error
	}
	ch := make(chan result, 2)

	for range 2 {
		go func() {
			client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)
			_, err := client.AddGenre(context.Background(), connect.NewRequest(&trendbirdv1.AddGenreRequest{
				Genre: "technology",
			}))
			ch <- result{err: err}
		}()
	}

	r1 := <-ch
	r2 := <-ch

	// 両方成功（冪等）
	if r1.err != nil {
		t.Errorf("goroutine 1 failed: %v", r1.err)
	}
	if r2.err != nil {
		t.Errorf("goroutine 2 failed: %v", r2.err)
	}

	// user_genres=1行（冪等）
	var count int64
	env.db.Model(&model.UserGenre{}).Where("user_id = ? AND genre_id = ?", user.ID, genreID).Count(&count)
	if count != 1 {
		t.Errorf("user_genres count: want 1, got %d", count)
	}
}

func TestTopicService_CreateTopic_AutoAddGenre(t *testing.T) {
	t.Run("auto_adds_genre", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		genreID := ensureGenre(t, env.db, "technology")

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTopicServiceClient)

		// user_genres にジャンルがない状態でトピック作成 → ジャンルが自動追加される
		_, err := client.CreateTopic(context.Background(), connect.NewRequest(&trendbirdv1.CreateTopicRequest{
			Name:     "AI Topic",
			Keywords: []string{"openai"},
			Genre:    "technology",
		}))
		if err != nil {
			t.Fatalf("CreateTopic: %v", err)
		}

		// user_genres に自動追加されていること
		var count int64
		env.db.Model(&model.UserGenre{}).Where("user_id = ? AND genre_id = ?", user.ID, genreID).Count(&count)
		if count != 1 {
			t.Errorf("expected genre auto-added, got count=%d", count)
		}
	})
}
