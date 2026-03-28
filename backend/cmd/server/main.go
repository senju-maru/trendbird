package main

import (
	"log/slog"
	"os"

	"github.com/trendbird/backend/internal/adapter/router"
	"github.com/trendbird/backend/internal/di"
	"github.com/trendbird/backend/internal/infrastructure/config"
	"github.com/trendbird/backend/internal/infrastructure/server"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	container, err := di.NewContainer(cfg)
	if err != nil {
		slog.Error("failed to initialize container", "error", err)
		os.Exit(1)
	}

	sqlDB, err := container.DB.DB()
	if err != nil {
		slog.Error("failed to get sql.DB", "error", err)
		os.Exit(1)
	}
	defer sqlDB.Close()

	handler := router.New(container)
	srv := server.NewServer(cfg, handler)

	if err := server.Run(srv); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
