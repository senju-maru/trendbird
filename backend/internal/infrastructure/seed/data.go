package seed

import (
	"encoding/json"
	"log/slog"
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Fixed UUIDs (00000000-0000-0000-XXXX-YYYYYYYYYYYY)
// ---------------------------------------------------------------------------

const (
	// テストユーザーの固定 UUID（users テーブルが空の場合のみ使用）
	userID = "00000000-0000-0000-0001-000000000001"

	// Topics (shared) — truncate 後も ID を固定することで topic_volumes 等と結合可能
	topicID1 = "00000000-0000-0000-0010-000000000001"
	topicID2 = "00000000-0000-0000-0010-000000000002"
	topicID3 = "00000000-0000-0000-0010-000000000003"
	topicID4 = "00000000-0000-0000-0010-000000000004"
	topicID5 = "00000000-0000-0000-0010-000000000005"
	topicID6 = "00000000-0000-0000-0010-000000000006"
	topicID7 = "00000000-0000-0000-0010-000000000007"
)

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// genreIDMap は seed 内で使う genre slug → id マップ。
// seedGenreLookup で DB から取得して埋める。
var genreIDMap map[string]string

func seedGenreLookup(tx *gorm.DB) error {
	var genres []model.Genre
	if err := tx.Find(&genres).Error; err != nil {
		return err
	}
	genreIDMap = make(map[string]string, len(genres))
	for _, g := range genres {
		genreIDMap[g.Slug] = g.ID
	}
	return nil
}

func genreID(slug string) string {
	return genreIDMap[slug]
}

// ---------------------------------------------------------------------------
// 1. Users (users が空の場合のみテストユーザーを作成)
// ---------------------------------------------------------------------------

// seedUsersIfEmpty は users テーブルが空の場合のみテストユーザーを INSERT する。
// 実ユーザーが既にログインしている場合はスキップする。
func seedUsersIfEmpty(tx *gorm.DB) error {
	var count int64
	if err := tx.Model(&model.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		slog.Info("users already exist, skipping test user creation")
		return nil
	}
	return seedUsers(tx)
}

func seedUsers(tx *gorm.DB) error {
	users := []model.User{
		{
			ID:            userID,
			TwitterID:     "tw-seed-001",
			Name:          "田中 太郎",
			Email:         "tanaka@example.com",
			Image:         "https://pbs.twimg.com/profile_images/example/tanaka.jpg",
			TwitterHandle: "tanaka_dev",
			CreatedAt:     time.Date(2025, 11, 15, 9, 0, 0, 0, time.UTC),
			UpdatedAt:     time.Now(),
		},
	}
	return tx.Create(&users).Error
}

// ---------------------------------------------------------------------------
// 3. NotificationSettings
// ---------------------------------------------------------------------------

// seedNotificationSettingsIfMissing は各ユーザーに notification_settings がなければ作成する。
func seedNotificationSettingsIfMissing(tx *gorm.DB) error {
	var users []model.User
	if err := tx.Find(&users).Error; err != nil {
		return err
	}
	for _, u := range users {
		var count int64
		if err := tx.Model(&model.NotificationSetting{}).Where("user_id = ?", u.ID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		record := model.NotificationSetting{
			ID:            uuid.NewString(),
			UserID:        u.ID,
			SpikeEnabled:  true,
			RisingEnabled: true,
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// 4. UserGenres
// ---------------------------------------------------------------------------

// seedUserGenresForAllUsers は DB 上の全ユーザーに technology / business ジャンルを関連付ける。
// TRUNCATE 後に実行されるため、既存エントリは存在しない想定だが念のため FirstOrCreate で重複を回避する。
func seedUserGenresForAllUsers(tx *gorm.DB) error {
	var users []model.User
	if err := tx.Find(&users).Error; err != nil {
		return err
	}
	genreSlugs := []string{"technology", "business"}
	for _, u := range users {
		for _, slug := range genreSlugs {
			gid := genreID(slug)
			var record model.UserGenre
			result := tx.Where(model.UserGenre{UserID: u.ID, GenreID: gid}).FirstOrCreate(&record, model.UserGenre{
				ID:      uuid.NewString(),
				UserID:  u.ID,
				GenreID: gid,
			})
			if result.Error != nil {
				return result.Error
			}
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// 5. Topics (7 shared records)
// ---------------------------------------------------------------------------

func seedTopics(tx *gorm.DB) error {
	now := time.Now()
	spikeStarted := now.Add(-2 * time.Hour)
	z4 := 4.2
	z3 := 3.1
	z1 := 1.5
	z08 := 0.8
	z45 := 4.5
	z05 := 0.5
	z03 := 0.3

	tech := genreID("technology")
	biz := genreID("business")

	records := []model.Topic{
		{
			ID: topicID1, Name: "ChatGPT 動向",
			Keywords: mustJSON([]string{"ChatGPT", "OpenAI", "GPT-5"}),
			GenreID: tech, Status: 1, // Spike
			ChangePercent: 285.0, ZScore: &z4,
			CurrentVolume: 12500, BaselineVolume: 3200,
			ContextSummary: strPtr("ChatGPT の新バージョン発表に伴い関連ワードが急増"),
			SpikeStartedAt: &spikeStarted,
		},
		{
			ID: topicID2, Name: "Claude API",
			Keywords: mustJSON([]string{"Claude", "Anthropic", "Claude API"}),
			GenreID: tech, Status: 2, // Rising
			ChangePercent: 145.0, ZScore: &z3,
			CurrentVolume: 8200, BaselineVolume: 3400,
			ContextSummary: strPtr("Claude の新モデルリリースで API 利用者が増加"),
		},
		{
			ID: topicID3, Name: "LLM 最新研究",
			Keywords: mustJSON([]string{"LLM", "大規模言語モデル", "Transformer"}),
			GenreID: tech, Status: 3, // Stable
			ChangePercent: 12.0, ZScore: &z08,
			CurrentVolume: 4500, BaselineVolume: 4000,
			ContextSummary: strPtr("LLM 関連の学術研究が安定的に注目を集めている"),
		},
		{
			ID: topicID4, Name: "Go 言語トレンド",
			Keywords: mustJSON([]string{"Go", "Golang", "Go言語"}),
			GenreID: tech, Status: 1, // Spike
			ChangePercent: 210.0, ZScore: &z3,
			CurrentVolume: 9800, BaselineVolume: 3100,
			ContextSummary: strPtr("Go の新バージョンリリースでコミュニティが活性化"),
			SpikeStartedAt: &spikeStarted,
		},
		{
			ID: topicID5, Name: "Rust エコシステム",
			Keywords: mustJSON([]string{"Rust", "Rustlang", "Cargo"}),
			GenreID: tech, Status: 2, // Rising
			ChangePercent: 95.0, ZScore: &z1,
			CurrentVolume: 6200, BaselineVolume: 3200,
			ContextSummary: strPtr("Rust の採用企業が増加し、エコシステムが拡大中"),
		},
		{
			ID: topicID6, Name: "スタートアップ資金調達",
			Keywords: mustJSON([]string{"資金調達", "スタートアップ", "VC"}),
			GenreID: biz, Status: 3, // Stable
			ChangePercent: 8.0, ZScore: &z05,
			CurrentVolume: 3200, BaselineVolume: 2900,
			ContextSummary: strPtr("スタートアップの資金調達ニュースが安定的に注目を集めている"),
		},
		{
			ID: topicID7, Name: "Cursor",
			Keywords: mustJSON([]string{"Cursor", "cursor AI", "cursor editor"}),
			GenreID: tech, Status: 1, // Spike
			ChangePercent: 320.0, ZScore: &z45,
			CurrentVolume: 14200, BaselineVolume: 3400,
			ContextSummary: strPtr("Cursor 1.0 正式リリースで AI コードエディタへの関心が急増"),
			SpikeStartedAt: &spikeStarted,
		},
	}
	_ = z03
	return tx.Create(&records).Error
}

// ---------------------------------------------------------------------------
// 6. UserTopics
// ---------------------------------------------------------------------------

// seedUserTopicsForAllUsers は DB 上の全ユーザーに 7 トピックを関連付ける。
// TRUNCATE 後に実行されるため既存エントリは存在しないが、念のため FirstOrCreate で重複を回避する。
func seedUserTopicsForAllUsers(tx *gorm.DB) error {
	var users []model.User
	if err := tx.Find(&users).Error; err != nil {
		return err
	}
	topicIDs := []string{topicID1, topicID2, topicID3, topicID4, topicID5, topicID6, topicID7}
	for _, u := range users {
		for _, tid := range topicIDs {
			var record model.UserTopic
			result := tx.Where(model.UserTopic{UserID: u.ID, TopicID: tid}).FirstOrCreate(&record, model.UserTopic{
				ID:                  uuid.NewString(),
				UserID:              u.ID,
				TopicID:             tid,
				NotificationEnabled: true,
			})
			if result.Error != nil {
				return result.Error
			}
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// 7. TopicVolumes (24 data points per topic = 168 total)
// ---------------------------------------------------------------------------

func seedTopicVolumes(tx *gorm.DB) error {
	now := time.Now().Truncate(time.Hour)
	type topicPattern struct {
		id       string
		baseline int32
		spike    bool
	}
	patterns := []topicPattern{
		{topicID1, 3200, true},
		{topicID2, 3400, true},
		{topicID3, 4000, false},
		{topicID4, 3100, true},
		{topicID5, 3200, true},
		{topicID6, 2900, false},
		{topicID7, 3400, true},
	}

	rng := rand.New(rand.NewSource(42))
	var records []model.TopicVolume

	for _, p := range patterns {
		for i := 23; i >= 0; i-- {
			ts := now.Add(-time.Duration(i) * time.Hour)
			var value int32
			if p.spike {
				// Gradual increase peaking at recent hours
				factor := 1.0 + 2.0*math.Pow(float64(24-i)/24.0, 2)
				value = int32(float64(p.baseline) * factor)
			} else {
				// Stable with slight random variation
				value = p.baseline + int32(rng.Intn(200)-100)
			}
			records = append(records, model.TopicVolume{
				TopicID:   p.id,
				Timestamp: ts,
				Value:     value,
			})
		}
	}

	return tx.Create(&records).Error
}

// ---------------------------------------------------------------------------
// 8. SpikeHistories (Spike/Rising topics: 1-2 records each = 8 total)
// ---------------------------------------------------------------------------

func seedSpikeHistories(tx *gorm.DB) error {
	now := time.Now()
	records := []model.SpikeHistory{
		{TopicID: topicID1, Timestamp: now.Add(-2 * time.Hour), PeakZScore: 4.2, Status: 1, Summary: "ChatGPT 新バージョンの発表で関連ワードが急増", DurationMinutes: 120},
		{TopicID: topicID1, Timestamp: now.Add(-26 * time.Hour), PeakZScore: 3.5, Status: 1, Summary: "OpenAI の記者会見に関連するワードが一時的に急増", DurationMinutes: 45},
		{TopicID: topicID2, Timestamp: now.Add(-6 * time.Hour), PeakZScore: 3.1, Status: 2, Summary: "Claude の新モデル発表で API 関連の言及が増加", DurationMinutes: 90},
		{TopicID: topicID4, Timestamp: now.Add(-3 * time.Hour), PeakZScore: 3.8, Status: 1, Summary: "Go 新バージョンリリースで Gopher コミュニティが盛り上がり", DurationMinutes: 150},
		{TopicID: topicID4, Timestamp: now.Add(-50 * time.Hour), PeakZScore: 2.9, Status: 2, Summary: "Go カンファレンスの発表内容がトレンド入り", DurationMinutes: 60},
		{TopicID: topicID5, Timestamp: now.Add(-12 * time.Hour), PeakZScore: 2.5, Status: 2, Summary: "Rust Foundation の新プロジェクト発表", DurationMinutes: 75},
		{TopicID: topicID7, Timestamp: now.Add(-1 * time.Hour), PeakZScore: 4.5, Status: 1, Summary: "Cursor 1.0 正式リリースで AI コードエディタへの関心が急増", DurationMinutes: 60},
		{TopicID: topicID7, Timestamp: now.Add(-30 * time.Hour), PeakZScore: 3.0, Status: 2, Summary: "Cursor の新機能プレビューが話題に", DurationMinutes: 45},
	}
	return tx.Create(&records).Error
}

// ---------------------------------------------------------------------------
// 9. PostingTips (1 per topic = 7 total)
// ---------------------------------------------------------------------------

func seedPostingTips(tx *gorm.DB) error {
	now := time.Now()
	tomorrow9am := time.Date(now.Year(), now.Month(), now.Day()+1, 9, 0, 0, 0, now.Location())

	records := []model.PostingTip{
		{TopicID: topicID1, PeakDays: mustJSON([]string{"月", "水", "金"}), PeakHoursStart: 9, PeakHoursEnd: 12, NextSuggestedTime: tomorrow9am},
		{TopicID: topicID2, PeakDays: mustJSON([]string{"火", "木"}), PeakHoursStart: 10, PeakHoursEnd: 14, NextSuggestedTime: tomorrow9am.Add(time.Hour)},
		{TopicID: topicID3, PeakDays: mustJSON([]string{"月", "火", "水", "木", "金"}), PeakHoursStart: 8, PeakHoursEnd: 11, NextSuggestedTime: tomorrow9am.Add(-time.Hour)},
		{TopicID: topicID4, PeakDays: mustJSON([]string{"月", "水", "金"}), PeakHoursStart: 11, PeakHoursEnd: 15, NextSuggestedTime: tomorrow9am.Add(2 * time.Hour)},
		{TopicID: topicID5, PeakDays: mustJSON([]string{"火", "木", "土"}), PeakHoursStart: 10, PeakHoursEnd: 13, NextSuggestedTime: tomorrow9am.Add(time.Hour)},
		{TopicID: topicID6, PeakDays: mustJSON([]string{"水", "金"}), PeakHoursStart: 9, PeakHoursEnd: 12, NextSuggestedTime: tomorrow9am},
		{TopicID: topicID7, PeakDays: mustJSON([]string{"月", "火", "水", "木", "金"}), PeakHoursStart: 10, PeakHoursEnd: 14, NextSuggestedTime: tomorrow9am.Add(time.Hour)},
	}
	return tx.Create(&records).Error
}

func strPtr(s string) *string {
	return &s
}
