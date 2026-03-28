package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const migrationsDir = "/migrations"

// runMigrate applies all pending migrations from /migrations/*.up.sql.
func runMigrate(dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	// schema_migrations テーブルを作成（存在しなければ）
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	// マイグレーションファイルを取得してソート
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	if err != nil {
		return fmt.Errorf("glob migration files: %w", err)
	}
	sort.Strings(files)

	if len(files) == 0 {
		slog.Info("no migration files found")
		return nil
	}

	// 適用済みバージョンを取得
	applied, err := getAppliedVersions(db)
	if err != nil {
		return fmt.Errorf("get applied versions: %w", err)
	}

	for _, f := range files {
		version := extractVersion(f)
		if applied[version] {
			slog.Info("migration already applied, skipping", "version", version)
			continue
		}

		slog.Info("applying migration", "version", version, "file", filepath.Base(f))

		content, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", version, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			// 既に適用済みのDDLエラーは無視して記録
			if isAlreadyAppliedError(err) {
				slog.Warn("migration DDL already applied, recording", "version", version, "error", err)
			} else {
				return fmt.Errorf("execute migration %s: %w", version, err)
			}
		}

		if _, err := db.Exec(`INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
			return fmt.Errorf("record migration %s: %w", version, err)
		}

		slog.Info("migration applied", "version", version)
	}

	return nil
}

// getAppliedVersions returns a set of already-applied migration versions.
func getAppliedVersions(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		applied[v] = true
	}
	return applied, rows.Err()
}

// extractVersion extracts the version prefix from a migration filename.
// e.g. "000001_create_initial_tables.up.sql" -> "000001"
func extractVersion(path string) string {
	base := filepath.Base(path)
	if idx := strings.IndexByte(base, '_'); idx > 0 {
		return base[:idx]
	}
	return base
}

// isAlreadyAppliedError checks if the error indicates the DDL was already applied.
// PostgreSQL SQLSTATE codes:
//   - 42P07: duplicate_table (CREATE TABLE)
//   - 42710: duplicate_object (CREATE INDEX, ADD CONSTRAINT)
//   - 42701: duplicate_column (ADD COLUMN)
//   - 42703: undefined_column (DROP COLUMN already executed)
//   - 42P01: undefined_table (DROP TABLE already executed)
func isAlreadyAppliedError(err error) bool {
	msg := err.Error()
	for _, code := range []string{"42P07", "42710", "42701", "42703", "42P01"} {
		if strings.Contains(msg, code) {
			return true
		}
	}
	return false
}
