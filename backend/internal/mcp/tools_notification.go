package mcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/trendbird/backend/internal/di"
)

func registerNotificationTools(s *server.MCPServer, c *di.MCPContainer) {
	s.AddTool(mcp.NewTool("list_notifications",
		mcp.WithDescription("通知の一覧を取得します"),
	), makeListNotifications(c))

	s.AddTool(mcp.NewTool("mark_all_notifications_read",
		mcp.WithDescription("すべての通知を既読にします"),
	), makeMarkAllNotificationsRead(c))
}

func makeListNotifications(c *di.MCPContainer) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		notifications, _, err := c.NotificationUC.ListNotifications(ctx, c.UserID, 20, 0)
		if err != nil {
			return errorResult(err), nil
		}

		if len(notifications) == 0 {
			return mcp.NewToolResultText("通知はありません。"), nil
		}

		var sb strings.Builder
		fmt.Fprintf(&sb, "通知: %d件\n\n", len(notifications))
		for _, n := range notifications {
			read := "未読"
			if n.IsRead {
				read = "既読"
			}
			jst := n.CreatedAt.In(time.FixedZone("JST", 9*3600))
			fmt.Fprintf(&sb, "- [%s] %s\n  %s (%s)\n\n", read, n.Title, n.Message, jst.Format("01/02 15:04"))
		}
		return mcp.NewToolResultText(sb.String()), nil
	}
}

func makeMarkAllNotificationsRead(c *di.MCPContainer) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.NotificationUC.MarkAllAsRead(ctx, c.UserID); err != nil {
			return errorResult(err), nil
		}
		return mcp.NewToolResultText("すべての通知を既読にしました。"), nil
	}
}
