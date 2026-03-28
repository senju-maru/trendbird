package mcp

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/trendbird/backend/internal/di"
)

// NewServer は MCP サーバーを作成し、全ツールを登録して返す。
func NewServer(c *di.MCPContainer) *server.MCPServer {
	s := server.NewMCPServer(
		"TrendBird",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	registerPostTools(s, c)
	registerTopicTools(s, c)
	registerNotificationTools(s, c)
	registerAutoReplyTools(s, c)
	registerAutoDMTools(s, c)

	return s
}
