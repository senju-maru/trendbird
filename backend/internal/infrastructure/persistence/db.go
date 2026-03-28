package persistence

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/trendbird/backend/internal/infrastructure/config"
)

// NewDB は PostgreSQL への接続を確立し *gorm.DB を返す。
func NewDB(cfg *config.Config) (*gorm.DB, error) {
	return newDB(cfg.DatabaseURL, 25, 5)
}

// NewBatchDB はバッチジョブ用の PostgreSQL 接続を確立し *gorm.DB を返す。
// サーバーより小さい接続プール設定を使用する。
func NewBatchDB(cfg *config.BatchConfig) (*gorm.DB, error) {
	return newDB(cfg.DatabaseURL, 5, 2)
}

// NewSchedulerDB はローカルスケジューラ用の PostgreSQL 接続を確立し *gorm.DB を返す。
// 複数ジョブが並行実行されるため、バッチ用より大きい接続プールを使用する。
func NewSchedulerDB(cfg *config.BatchConfig) (*gorm.DB, error) {
	return newDB(cfg.DatabaseURL, 10, 5)
}

func newDB(dsn string, maxOpen, maxIdle int) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}
