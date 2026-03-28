package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/trendbird/backend/internal/infrastructure/seed"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		slog.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("failed to get sql.DB", "error", err)
		os.Exit(1)
	}
	defer sqlDB.Close()

	if err := seed.Run(db); err != nil {
		slog.Error("seed failed", "error", err)
		os.Exit(1)
	}

	fmt.Println("Seed completed successfully!")
	fmt.Println("ブラウザをリロード (Cmd+R) してデータを確認してください。")
}
