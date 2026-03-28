package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/trendbird/backend/internal/di"
)

func registerAutoReplyTools(s *server.MCPServer, c *di.MCPContainer) {
	s.AddTool(mcp.NewTool("list_auto_reply_rules",
		mcp.WithDescription("自動リプライルールの一覧を取得します"),
	), makeListAutoReplyRules(c))

	s.AddTool(mcp.NewTool("create_auto_reply_rule",
		mcp.WithDescription("自動リプライルールを作成します。指定したポストへのリプライに含まれるキーワードに反応して自動返信します"),
		mcp.WithString("target_tweet_id", mcp.Required(), mcp.Description("監視対象のポストID（URLでも可）")),
		mcp.WithString("target_tweet_text", mcp.Required(), mcp.Description("監視対象ポストの内容メモ")),
		mcp.WithString("keywords", mcp.Required(), mcp.Description("トリガーキーワード（カンマ区切り）")),
		mcp.WithString("reply_template", mcp.Required(), mcp.Description("自動返信のテンプレート文")),
	), makeCreateAutoReplyRule(c))

	s.AddTool(mcp.NewTool("delete_auto_reply_rule",
		mcp.WithDescription("自動リプライルールを削除します"),
		mcp.WithString("rule_id", mcp.Required(), mcp.Description("ルールID")),
	), makeDeleteAutoReplyRule(c))
}

func makeListAutoReplyRules(c *di.MCPContainer) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		rules, err := c.AutoReplyUC.ListRules(ctx, c.UserID)
		if err != nil {
			return errorResult(err), nil
		}

		if len(rules) == 0 {
			return mcp.NewToolResultText("自動リプライルールはありません。create_auto_reply_rule で作成できます。"), nil
		}

		var sb strings.Builder
		fmt.Fprintf(&sb, "自動リプライルール: %d件\n\n", len(rules))
		for _, r := range rules {
			enabled := "有効"
			if !r.Enabled {
				enabled = "無効"
			}
			fmt.Fprintf(&sb, "- [%s] 対象ポスト: %s\n", enabled, truncate(r.TargetTweetText, 40))
			fmt.Fprintf(&sb, "  キーワード: %s\n", strings.Join(r.TriggerKeywords, ", "))
			fmt.Fprintf(&sb, "  返信テンプレート: %s\n", truncate(r.ReplyTemplate, 60))
			fmt.Fprintf(&sb, "  ID: %s\n\n", r.ID)
		}
		return mcp.NewToolResultText(sb.String()), nil
	}
}

func makeCreateAutoReplyRule(c *di.MCPContainer) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		targetTweetID, _ := args["target_tweet_id"].(string)
		targetTweetText, _ := args["target_tweet_text"].(string)
		keywordsStr, _ := args["keywords"].(string)
		replyTemplate, _ := args["reply_template"].(string)

		keywords := splitAndTrim(keywordsStr)

		rule, err := c.AutoReplyUC.CreateRule(ctx, c.UserID, targetTweetID, targetTweetText, keywords, replyTemplate)
		if err != nil {
			return errorResult(err), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("自動リプライルールを作成しました。\n対象ポスト: %s\nキーワード: %s\nID: %s",
			truncate(rule.TargetTweetText, 40), strings.Join(rule.TriggerKeywords, ", "), rule.ID)), nil
	}
}

func makeDeleteAutoReplyRule(c *di.MCPContainer) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		ruleID, _ := args["rule_id"].(string)

		if err := c.AutoReplyUC.DeleteRule(ctx, c.UserID, ruleID); err != nil {
			return errorResult(err), nil
		}

		return mcp.NewToolResultText("自動リプライルールを削除しました。"), nil
	}
}
