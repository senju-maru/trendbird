package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/trendbird/backend/internal/di"
	"github.com/trendbird/backend/internal/infrastructure/config"
)

// jobTimeouts defines per-job timeouts.
// trend-fetch calls the X API for many topics and can exceed 10 minutes under normal load.
var jobTimeouts = map[string]time.Duration{
	"trend-fetch":     25 * time.Minute,
	"topic-research":  25 * time.Minute,
}

const defaultBatchTimeout = 10 * time.Minute

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	jobType := os.Getenv("BATCH_JOB_TYPE")
	if jobType == "" {
		slog.Error("BATCH_JOB_TYPE environment variable is required")
		os.Exit(1)
	}

	// "migrate" は BatchContainer 不要 — DATABASE_URL のみで動作
	if jobType == "migrate" {
		dsn := os.Getenv("DATABASE_URL")
		if dsn == "" {
			slog.Error("DATABASE_URL environment variable is required for migrate")
			os.Exit(1)
		}
		if err := runMigrate(dsn); err != nil {
			slog.Error("migration failed", "error", err)
			os.Exit(1)
		}
		slog.Info("migration completed successfully")
		return
	}

	cfg, err := config.LoadBatch()
	if err != nil {
		slog.Error("failed to load batch config", "error", err)
		os.Exit(1)
	}

	container, err := di.NewBatchContainer(cfg)
	if err != nil {
		slog.Error("failed to initialize batch container", "error", err)
		os.Exit(1)
	}
	defer container.Close()

	timeout := defaultBatchTimeout
	if t, ok := jobTimeouts[jobType]; ok {
		timeout = t
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	slog.Info("starting batch job", "type", jobType)

	if err := run(ctx, container, jobType); err != nil {
		slog.Error("batch job failed", "type", jobType, "error", err)
		os.Exit(1)
	}

	slog.Info("batch job completed successfully", "type", jobType)
}

func run(ctx context.Context, container *di.BatchContainer, jobType string) error {
	switch jobType {
	case "spike-notification":
		return container.SpikeNotificationUC.Execute(ctx)
	case "rising-notification":
		return container.RisingNotificationUC.Execute(ctx)
	case "trend-fetch":
		return container.TrendFetchUC.Execute(ctx)
	case "scheduled-publish":
		return container.ScheduledPublishUC.Execute(ctx)
	case "reply-dm-batch":
		return container.AutoDMBatchUC.Execute(ctx)
	case "topic-research":
		return container.TopicResearchCollectionUC.Execute(ctx)
	default:
		return fmt.Errorf("unknown batch job type: %s", jobType)
	}
}
