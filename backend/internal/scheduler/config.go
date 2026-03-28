package scheduler

import (
	"github.com/caarlos0/env/v11"

	"github.com/trendbird/backend/internal/infrastructure/config"
)

// SchedulerConfig はローカルスケジューラの設定を保持する。
// BatchConfig を埋め込み、ジョブ別のスケジュール設定を追加する。
type SchedulerConfig struct {
	config.BatchConfig

	// ジョブ別 cron 式（環境変数で上書き可能）
	SpikeNotificationSchedule  string `env:"SCHEDULE_SPIKE_NOTIFICATION"  envDefault:"*/5 * * * *"`
	RisingNotificationSchedule string `env:"SCHEDULE_RISING_NOTIFICATION" envDefault:"*/10 * * * *"`
	ScheduledPublishSchedule   string `env:"SCHEDULE_SCHEDULED_PUBLISH"   envDefault:"0 * * * *"`
	ReplyDMBatchSchedule       string `env:"SCHEDULE_REPLY_DM_BATCH"      envDefault:"0 * * * *"`
	AutoReplyBatchSchedule     string `env:"SCHEDULE_AUTO_REPLY_BATCH"    envDefault:"0 * * * *"`
	TrendFetchSchedule         string `env:"SCHEDULE_TREND_FETCH"         envDefault:"0 */12 * * *"`
	// 空 = trend-fetch 完了後にチェーン実行。cron 式を設定すると独立スケジュールで実行。
	TopicResearchSchedule string `env:"SCHEDULE_TOPIC_RESEARCH" envDefault:""`

	// cron 式を解釈するタイムゾーン（デフォルト: Asia/Tokyo）
	Timezone string `env:"SCHEDULER_TIMEZONE" envDefault:"Asia/Tokyo"`
	// カンマ区切りで無効化するジョブ名（例: "spike-notification,reply-dm-batch"）
	DisabledJobs string `env:"SCHEDULER_DISABLED_JOBS" envDefault:""`
}

// LoadConfig は環境変数を読み込み SchedulerConfig を返す。
func LoadConfig() (*SchedulerConfig, error) {
	cfg := &SchedulerConfig{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
