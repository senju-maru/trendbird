package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/trendbird/backend/internal/di"
)

func registerAutoDMTools(s *server.MCPServer, c *di.MCPContainer) {
	s.AddTool(mcp.NewTool("list_auto_dm_rules",
		mcp.WithDescription("自動DMルールの一覧を取得します"),
	), makeListAutoDMRules(c))

	s.AddTool(mcp.NewTool("create_auto_dm_rule",
		mcp.WithDescription("自動DMルールを作成します。指定キーワードを含むリプライに対して自動でDMを送信します"),
		mcp.WithString("keywords", mcp.Required(), mcp.Description("トリガーキーワード（カンマ区切り）")),
		mcp.WithString("template_message", mcp.Required(), mcp.Description("DM送信テンプレート文")),
	), makeCreateAutoDMRule(c))

	s.AddTool(mcp.NewTool("delete_auto_dm_rule",
		mcp.WithDescription("自動DMルールを削除します"),
		mcp.WithString("rule_id", mcp.Required(), mcp.Description("ルールID")),
	), makeDeleteAutoDMRule(c))
}

func makeListAutoDMRules(c *di.MCPContainer) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		rules, err := c.AutoDMUC.ListRules(ctx, c.UserID)
		if err != nil {
			return errorResult(err), nil
		}

		if len(rules) == 0 {
			return mcp.NewToolResultText("自動DMルールはありません。create_auto_dm_rule で作成できます。"), nil
		}

		var sb strings.Builder
		fmt.Fprintf(&sb, "自動DMルール: %d件\n\n", len(rules))
		for _, r := range rules {
			enabled := "有効"
			if !r.Enabled {
				enabled = "無効"
			}
			fmt.Fprintf(&sb, "- [%s] キーワード: %s\n", enabled, strings.Join(r.TriggerKeywords, ", "))
			fmt.Fprintf(&sb, "  テンプレート: %s\n", truncate(r.TemplateMessage, 60))
			fmt.Fprintf(&sb, "  ID: %s\n\n", r.ID)
		}
		return mcp.NewToolResultText(sb.String()), nil
	}
}

func makeCreateAutoDMRule(c *di.MCPContainer) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		keywordsStr, _ := args["keywords"].(string)
		templateMessage, _ := args["template_message"].(string)

		keywords := splitAndTrim(keywordsStr)

		rule, err := c.AutoDMUC.CreateRule(ctx, c.UserID, keywords, templateMessage)
		if err != nil {
			return errorResult(err), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("自動DMルールを作成しました。\nキーワード: %s\nID: %s",
			strings.Join(rule.TriggerKeywords, ", "), rule.ID)), nil
	}
}

func makeDeleteAutoDMRule(c *di.MCPContainer) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		ruleID, _ := args["rule_id"].(string)

		if err := c.AutoDMUC.DeleteRule(ctx, c.UserID, ruleID); err != nil {
			return errorResult(err), nil
		}

		return mcp.NewToolResultText("自動DMルールを削除しました。"), nil
	}
}
