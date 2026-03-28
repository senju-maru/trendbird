package scheduler

import (
	"testing"
)

func TestParseDisabledJobs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: map[string]bool{},
		},
		{
			name:  "single job",
			input: "spike-notification",
			expected: map[string]bool{
				"spike-notification": true,
			},
		},
		{
			name:  "multiple jobs",
			input: "spike-notification,rising-notification,reply-dm-batch",
			expected: map[string]bool{
				"spike-notification":  true,
				"rising-notification": true,
				"reply-dm-batch":      true,
			},
		},
		{
			name:  "with spaces",
			input: " spike-notification , rising-notification ",
			expected: map[string]bool{
				"spike-notification":  true,
				"rising-notification": true,
			},
		},
		{
			name:  "trailing comma",
			input: "spike-notification,",
			expected: map[string]bool{
				"spike-notification": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDisabledJobs(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d disabled jobs, got %d", len(tt.expected), len(result))
			}
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("expected %s=%v, got %v", k, v, result[k])
				}
			}
		})
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	// DATABASE_URL は BatchConfig の required フィールド
	// テスト環境ではダミー値を設定
	t.Setenv("DATABASE_URL", "postgres://test@localhost:5432/test?sslmode=disable")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// デフォルト値の検証
	if cfg.SpikeNotificationSchedule != "*/5 * * * *" {
		t.Errorf("expected spike schedule '*/5 * * * *', got '%s'", cfg.SpikeNotificationSchedule)
	}
	if cfg.RisingNotificationSchedule != "*/10 * * * *" {
		t.Errorf("expected rising schedule '*/10 * * * *', got '%s'", cfg.RisingNotificationSchedule)
	}
	if cfg.TrendFetchSchedule != "0 */12 * * *" {
		t.Errorf("expected trend-fetch schedule '0 */12 * * *', got '%s'", cfg.TrendFetchSchedule)
	}
	if cfg.TopicResearchSchedule != "" {
		t.Errorf("expected empty topic-research schedule, got '%s'", cfg.TopicResearchSchedule)
	}
	if cfg.Timezone != "Asia/Tokyo" {
		t.Errorf("expected timezone 'Asia/Tokyo', got '%s'", cfg.Timezone)
	}
	if cfg.DisabledJobs != "" {
		t.Errorf("expected empty disabled jobs, got '%s'", cfg.DisabledJobs)
	}
}

func TestLoadConfig_EnvOverride(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test@localhost:5432/test?sslmode=disable")
	t.Setenv("SCHEDULE_TREND_FETCH", "0 */6 * * *")
	t.Setenv("SCHEDULER_TIMEZONE", "America/New_York")
	t.Setenv("SCHEDULER_DISABLED_JOBS", "spike-notification,reply-dm-batch")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.TrendFetchSchedule != "0 */6 * * *" {
		t.Errorf("expected trend-fetch schedule '0 */6 * * *', got '%s'", cfg.TrendFetchSchedule)
	}
	if cfg.Timezone != "America/New_York" {
		t.Errorf("expected timezone 'America/New_York', got '%s'", cfg.Timezone)
	}
	if cfg.DisabledJobs != "spike-notification,reply-dm-batch" {
		t.Errorf("expected disabled 'spike-notification,reply-dm-batch', got '%s'", cfg.DisabledJobs)
	}
}
