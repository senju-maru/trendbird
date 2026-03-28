package e2etest

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var testDB *gorm.DB

const (
	testJWTSecret = "test-secret-for-e2e"
	testJWTExpiry = 1 * time.Hour

	defaultTestDSN = "postgres://localhost:5432/trendbird_test?sslmode=disable"
)

func TestMain(m *testing.M) {
	// TEST_DATABASE_URL が設定されていればそれを使い、なければデフォルトを使用
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = defaultTestDSN
	}

	// Connect with GORM
	var err error
	testDB, err = gorm.Open(pgdriver.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to test database: %v\n", err)
		os.Exit(1)
	}

	// Drop all objects and re-run migrations for a clean schema.
	// Using DROP SCHEMA CASCADE is more reliable than running down migrations
	// in reverse order, which can fail and leave the database in a dirty state.
	testDB.Exec("DROP SCHEMA public CASCADE")
	testDB.Exec("CREATE SCHEMA public")

	// Apply up migrations in order
	upFiles, _ := filepath.Glob("../../migrations/*.up.sql")
	sort.Strings(upFiles)
	for _, f := range upFiles {
		migration, err := os.ReadFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read migration file %s: %v\n", f, err)
			os.Exit(1)
		}
		if err := testDB.Exec(string(migration)).Error; err != nil {
			fmt.Fprintf(os.Stderr, "failed to execute migration %s: %v\n", f, err)
			os.Exit(1)
		}
	}

	// Run tests
	code := m.Run()

	// DB 接続を閉じてリソースリークを防ぐ
	if sqlDB, err := testDB.DB(); err == nil {
		sqlDB.Close()
	}

	os.Exit(code)
}
