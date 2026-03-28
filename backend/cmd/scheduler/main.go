package main

import (
	"log/slog"
	"os"

	"github.com/trendbird/backend/internal/di"
	"github.com/trendbird/backend/internal/infrastructure/persistence"
	"github.com/trendbird/backend/internal/scheduler"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg, err := scheduler.LoadConfig()
	if err != nil {
		slog.Error("failed to load scheduler config", "error", err)
		os.Exit(1)
	}

	db, err := persistence.NewSchedulerDB(&cfg.BatchConfig)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	container, err := di.NewBatchContainerWithDB(&cfg.BatchConfig, db)
	if err != nil {
		slog.Error("failed to initialize batch container", "error", err)
		os.Exit(1)
	}
	defer container.Close()

	s, err := scheduler.New(cfg, container)
	if err != nil {
		slog.Error("failed to initialize scheduler", "error", err)
		os.Exit(1)
	}

	if err := s.Run(); err != nil {
		slog.Error("scheduler error", "error", err)
		os.Exit(1)
	}
}
