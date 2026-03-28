package scheduler

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/trendbird/backend/internal/di"
)

// Job はスケジュール実行されるバッチジョブを表す。
type Job struct {
	Name     string                             // ジョブ名（例: "trend-fetch"）
	Fn       func(ctx context.Context) error    // Execute メソッドへの参照
	Schedule string                             // cron 式。空 = チェーン実行のみ
	Timeout  time.Duration                      // 実行タイムアウト
	After    string                             // 依存先ジョブ名。空 = 独立実行
}

// Scheduler はローカル環境で定期バッチジョブを実行するスケジューラ。
type Scheduler struct {
	cron     *cron.Cron
	jobs     map[string]*Job
	disabled map[string]bool
	logger   *slog.Logger
}

// New は SchedulerConfig と BatchContainer からスケジューラを構築する。
func New(cfg *SchedulerConfig, container *di.BatchContainer) (*Scheduler, error) {
	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		return nil, err
	}

	c := cron.New(
		cron.WithLocation(loc),
		cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)),
	)

	disabled := parseDisabledJobs(cfg.DisabledJobs)

	s := &Scheduler{
		cron:     c,
		jobs:     make(map[string]*Job),
		disabled: disabled,
		logger:   slog.Default(),
	}

	jobs := buildJobs(cfg, container)
	for _, job := range jobs {
		s.jobs[job.Name] = job
	}

	// cron に登録（チェーン実行のみのジョブは登録しない）
	for _, job := range jobs {
		if disabled[job.Name] {
			s.logger.Info("job disabled", "job", job.Name)
			continue
		}
		if job.Schedule == "" {
			// チェーン実行のみ（親ジョブから呼ばれる）
			s.logger.Info("job chained", "job", job.Name, "after", job.After)
			continue
		}
		if _, err := c.AddFunc(job.Schedule, s.makeRunner(job)); err != nil {
			return nil, err
		}
		s.logger.Info("job registered", "job", job.Name, "schedule", job.Schedule)
	}

	return s, nil
}

// Run はスケジューラを起動し、SIGINT/SIGTERM で graceful shutdown する。
func (s *Scheduler) Run() error {
	s.cron.Start()
	s.logger.Info("scheduler started", "job_count", len(s.cron.Entries()))

	for _, entry := range s.cron.Entries() {
		s.logger.Info("next run", "next", entry.Next.Format(time.RFC3339))
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	s.logger.Info("shutdown signal received", "signal", sig)

	ctx := s.cron.Stop()
	select {
	case <-ctx.Done():
		s.logger.Info("all jobs completed, scheduler stopped")
	case <-time.After(30 * time.Minute):
		s.logger.Warn("shutdown timeout exceeded, forcing exit")
	}

	return nil
}

// makeRunner はジョブ実行関数を返す。タイムアウト・ログ・チェーン実行を含む。
func (s *Scheduler) makeRunner(job *Job) func() {
	return func() {
		s.executeJob(job)
	}
}

// executeJob はジョブを実行し、完了後にチェーン対象のジョブを順次実行する。
func (s *Scheduler) executeJob(job *Job) {
	ctx, cancel := context.WithTimeout(context.Background(), job.Timeout)
	defer cancel()

	s.logger.Info("job started", "job", job.Name)
	start := time.Now()

	if err := job.Fn(ctx); err != nil {
		s.logger.Error("job failed", "job", job.Name, "duration", time.Since(start), "error", err)
		return
	}

	s.logger.Info("job completed", "job", job.Name, "duration", time.Since(start))

	// チェーン実行: このジョブに依存しているジョブを探して順次実行
	for _, child := range s.jobs {
		if child.After == job.Name && child.Schedule == "" && !s.disabled[child.Name] {
			s.logger.Info("running chained job", "job", child.Name, "after", job.Name)
			s.executeJob(child)
		}
	}
}

// buildJobs は BatchContainer から全ジョブ定義を生成する。
func buildJobs(cfg *SchedulerConfig, c *di.BatchContainer) []*Job {
	return []*Job{
		{Name: "trend-fetch", Fn: c.TrendFetchUC.Execute, Schedule: cfg.TrendFetchSchedule, Timeout: 25 * time.Minute},
		{Name: "topic-research", Fn: c.TopicResearchCollectionUC.Execute, Schedule: cfg.TopicResearchSchedule, Timeout: 25 * time.Minute, After: "trend-fetch"},
		{Name: "spike-notification", Fn: c.SpikeNotificationUC.Execute, Schedule: cfg.SpikeNotificationSchedule, Timeout: 10 * time.Minute},
		{Name: "rising-notification", Fn: c.RisingNotificationUC.Execute, Schedule: cfg.RisingNotificationSchedule, Timeout: 10 * time.Minute},
		{Name: "scheduled-publish", Fn: c.ScheduledPublishUC.Execute, Schedule: cfg.ScheduledPublishSchedule, Timeout: 10 * time.Minute},
		{Name: "reply-dm-batch", Fn: c.AutoDMBatchUC.Execute, Schedule: cfg.ReplyDMBatchSchedule, Timeout: 10 * time.Minute},
		{Name: "auto-reply-batch", Fn: c.AutoReplyBatchUC.Execute, Schedule: cfg.AutoReplyBatchSchedule, Timeout: 10 * time.Minute},
	}
}

// parseDisabledJobs はカンマ区切りの文字列を set に変換する。
func parseDisabledJobs(s string) map[string]bool {
	disabled := make(map[string]bool)
	if s == "" {
		return disabled
	}
	for _, name := range strings.Split(s, ",") {
		name = strings.TrimSpace(name)
		if name != "" {
			disabled[name] = true
		}
	}
	return disabled
}
