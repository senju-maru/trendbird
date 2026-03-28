package seed

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

// Run executes the full seed process inside a single transaction.
func Run(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		slog.Info("truncating all tables...")
		if err := truncateAll(tx); err != nil {
			return fmt.Errorf("truncateAll: %w", err)
		}

		// genres テーブルはマイグレーションで seed 済みなので、slug → id マップを構築
		slog.Info("looking up genres...")
		if err := seedGenreLookup(tx); err != nil {
			return fmt.Errorf("genre lookup: %w", err)
		}

		steps := []struct {
			name string
			fn   func(*gorm.DB) error
		}{
			// ユーザーが存在しない場合のみテストユーザーを作成
			{"users (if empty)", seedUsersIfEmpty},
			{"notification_settings (if missing)", seedNotificationSettingsIfMissing},
			// 共有データ
			{"user_genres", seedUserGenresForAllUsers},
			{"topics", seedTopics},
			{"user_topics", seedUserTopicsForAllUsers},
			{"topic_volumes", seedTopicVolumes},
			{"spike_histories", seedSpikeHistories},
			{"posting_tips", seedPostingTips},
		}

		for _, s := range steps {
			slog.Info("seeding", "table", s.name)
			if err := s.fn(tx); err != nil {
				return fmt.Errorf("seed %s: %w", s.name, err)
			}
		}

		return nil
	})
}

func truncateAll(tx *gorm.DB) error {
	// users / twitter_connections / notification_settings は除外し、
	// ログイン済みの実ユーザーを保持したまま共有データだけリセットする
	const query = `TRUNCATE TABLE
		user_topics,
		user_notifications, notifications, activities, ai_generation_logs, generated_posts, posts,
		posting_tips, spike_histories, topic_volumes, topics,
		user_genres
		CASCADE`
	return tx.Exec(query).Error
}
