package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/trendbird/backend/internal/di"
	"github.com/trendbird/backend/internal/infrastructure/config"
	tbmcp "github.com/trendbird/backend/internal/mcp"
)

func main() {
	// MCP プロトコルは stdout を使うため、ログは stderr に出力する
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg, err := config.LoadMCP()
	if err != nil {
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "=== TrendBird MCP サーバー起動エラー ===")
		fmt.Fprintln(os.Stderr, "環境変数が設定されていません。以下の手順でセットアップしてください:")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "  1. make setup")
		fmt.Fprintln(os.Stderr, "  2. backend/.env を編集して X API キーを設定")
		fmt.Fprintln(os.Stderr, "  3. make start でサーバー起動")
		fmt.Fprintln(os.Stderr, "  4. http://localhost:3000 でXログイン")
		fmt.Fprintln(os.Stderr, "  5. Claude Code を再起動")
		fmt.Fprintln(os.Stderr, "")
		os.Exit(1)
	}

	container, err := di.NewMCPContainer(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "=== TrendBird MCP サーバー起動エラー ===")
		fmt.Fprintf(os.Stderr, "初期化に失敗: %v\n", err)
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "よくある原因:")
		fmt.Fprintln(os.Stderr, "  - PostgreSQL が起動していない → make db-up")
		fmt.Fprintln(os.Stderr, "  - マイグレーション未実行 → make migrate")
		fmt.Fprintln(os.Stderr, "  - Xログイン未完了 → make start → http://localhost:3000 でログイン")
		fmt.Fprintln(os.Stderr, "")
		os.Exit(1)
	}
	defer container.Close()

	slog.Info("MCP server starting", "user_id", container.UserID)

	mcpServer := tbmcp.NewServer(container)

	if err := server.ServeStdio(mcpServer); err != nil {
		slog.Error("MCP server error", "error", err)
		os.Exit(1)
	}
}
