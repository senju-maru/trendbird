package mcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/trendbird/backend/internal/di"
	"github.com/trendbird/backend/internal/domain/entity"
)

func registerPostTools(s *server.MCPServer, c *di.MCPContainer) {
	s.AddTool(mcp.NewTool("list_drafts",
		mcp.WithDescription("下書き・予約投稿の一覧を取得します"),
	), makeListDrafts(c))

	s.AddTool(mcp.NewTool("create_draft",
		mcp.WithDescription("投稿の下書きを作成します"),
		mcp.WithString("content", mcp.Required(), mcp.Description("投稿内容")),
		mcp.WithString("topic_id", mcp.Description("紐付けるトピックID（省略可）")),
	), makeCreateDraft(c))

	s.AddTool(mcp.NewTool("schedule_post",
		mcp.WithDescription("下書きを指定日時に予約投稿します。時刻は毎時00分のみ指定可能です"),
		mcp.WithString("post_id", mcp.Required(), mcp.Description("下書きのID")),
		mcp.WithString("scheduled_at", mcp.Required(), mcp.Description("予約日時（ISO 8601形式、例: 2026-03-29T12:00:00+09:00）")),
	), makeSchedulePost(c))

	s.AddTool(mcp.NewTool("create_and_schedule_post",
		mcp.WithDescription("投稿を作成して、指定日時に予約します。時刻は毎時00分のみ指定可能です"),
		mcp.WithString("content", mcp.Required(), mcp.Description("投稿内容")),
		mcp.WithString("scheduled_at", mcp.Required(), mcp.Description("予約日時（ISO 8601形式、例: 2026-03-29T12:00:00+09:00）")),
		mcp.WithString("topic_id", mcp.Description("紐付けるトピックID（省略可）")),
	), makeCreateAndSchedulePost(c))

	s.AddTool(mcp.NewTool("publish_post",
		mcp.WithDescription("下書きまたは予約投稿を今すぐXに投稿します"),
		mcp.WithString("post_id", mcp.Required(), mcp.Description("投稿のID")),
	), makePublishPost(c))

	s.AddTool(mcp.NewTool("list_post_history",
		mcp.WithDescription("投稿済みの履歴を取得します（いいね・リポスト数つき）"),
	), makeListPostHistory(c))

	s.AddTool(mcp.NewTool("generate_posts",
		mcp.WithDescription("AIでトピックに合った投稿文を3パターン自動生成します"),
		mcp.WithString("topic_id", mcp.Required(), mcp.Description("トピックID")),
		mcp.WithString("style", mcp.Description("生成スタイル: casual（カジュアル）, breaking（速報）, analysis（分析）。省略時はcasual")),
	), makeGeneratePosts(c))
}

func makeListDrafts(c *di.MCPContainer) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		drafts, stats, _, err := c.PostUC.ListDrafts(ctx, c.UserID, 50, 0)
		if err != nil {
			return errorResult(err), nil
		}

		if len(drafts) == 0 {
			return mcp.NewToolResultText("下書き・予約投稿はありません。"), nil
		}

		var sb strings.Builder
		fmt.Fprintf(&sb, "下書き: %d件 / 予約: %d件 / 投稿済み: %d件 / 失敗: %d件\n\n",
			stats.TotalDrafts, stats.TotalScheduled, stats.TotalPublished, stats.TotalFailed)

		for _, d := range drafts {
			fmt.Fprintf(&sb, "- [%s] %s (ID: %s)\n", postStatusLabel(d.Status), truncate(d.Content, 60), d.ID)
			if d.ScheduledAt != nil {
				fmt.Fprintf(&sb, "  予約日時: %s\n", d.ScheduledAt.In(time.FixedZone("JST", 9*3600)).Format("2006-01-02 15:04"))
			}
		}
		return mcp.NewToolResultText(sb.String()), nil
	}
}

func makeCreateDraft(c *di.MCPContainer) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		content, _ := req.GetArguments()["content"].(string)
		topicID := stringPtr(req.GetArguments()["topic_id"])

		post, err := c.PostUC.CreateDraft(ctx, c.UserID, content, topicID)
		if err != nil {
			return errorResult(err), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("下書きを作成しました。\nID: %s\n内容: %s", post.ID, truncate(post.Content, 100))), nil
	}
}

func makeSchedulePost(c *di.MCPContainer) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		postID, _ := req.GetArguments()["post_id"].(string)
		scheduledAtStr, _ := req.GetArguments()["scheduled_at"].(string)

		scheduledAt, err := time.Parse(time.RFC3339, scheduledAtStr)
		if err != nil {
			return mcp.NewToolResultError("日時の形式が正しくありません。ISO 8601形式で指定してください（例: 2026-03-29T12:00:00+09:00）"), nil
		}

		post, err := c.PostUC.SchedulePost(ctx, c.UserID, postID, scheduledAt)
		if err != nil {
			return errorResult(err), nil
		}

		jst := post.ScheduledAt.In(time.FixedZone("JST", 9*3600))
		return mcp.NewToolResultText(fmt.Sprintf("投稿を予約しました。\n予約日時: %s\nID: %s\n内容: %s",
			jst.Format("2006年1月2日 15時04分"), post.ID, truncate(post.Content, 100))), nil
	}
}

func makeCreateAndSchedulePost(c *di.MCPContainer) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		content, _ := req.GetArguments()["content"].(string)
		scheduledAtStr, _ := req.GetArguments()["scheduled_at"].(string)
		topicID := stringPtr(req.GetArguments()["topic_id"])

		scheduledAt, err := time.Parse(time.RFC3339, scheduledAtStr)
		if err != nil {
			return mcp.NewToolResultError("日時の形式が正しくありません。ISO 8601形式で指定してください（例: 2026-03-29T12:00:00+09:00）"), nil
		}

		post, err := c.PostUC.CreateDraft(ctx, c.UserID, content, topicID)
		if err != nil {
			return errorResult(err), nil
		}

		post, err = c.PostUC.SchedulePost(ctx, c.UserID, post.ID, scheduledAt)
		if err != nil {
			return errorResult(err), nil
		}

		jst := post.ScheduledAt.In(time.FixedZone("JST", 9*3600))
		return mcp.NewToolResultText(fmt.Sprintf("投稿を作成して予約しました。\n予約日時: %s\nID: %s\n内容: %s",
			jst.Format("2006年1月2日 15時04分"), post.ID, truncate(post.Content, 100))), nil
	}
}

func makePublishPost(c *di.MCPContainer) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		postID, _ := req.GetArguments()["post_id"].(string)

		post, err := c.PostUC.PublishPost(ctx, c.UserID, postID)
		if err != nil {
			return errorResult(err), nil
		}

		msg := fmt.Sprintf("Xに投稿しました。\nID: %s", post.ID)
		if post.TweetURL != nil {
			msg += fmt.Sprintf("\nURL: %s", *post.TweetURL)
		}
		return mcp.NewToolResultText(msg), nil
	}
}

func makeListPostHistory(c *di.MCPContainer) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		posts, _, err := c.PostUC.ListPostHistory(ctx, c.UserID, 20, 0)
		if err != nil {
			return errorResult(err), nil
		}

		if len(posts) == 0 {
			return mcp.NewToolResultText("投稿履歴はまだありません。"), nil
		}

		var sb strings.Builder
		sb.WriteString("投稿履歴:\n\n")
		for _, p := range posts {
			fmt.Fprintf(&sb, "- %s\n", truncate(p.Content, 60))
			if p.PublishedAt != nil {
				jst := p.PublishedAt.In(time.FixedZone("JST", 9*3600))
				fmt.Fprintf(&sb, "  投稿日: %s\n", jst.Format("2006-01-02 15:04"))
			}
			fmt.Fprintf(&sb, "  いいね: %d / リポスト: %d / リプライ: %d / 閲覧: %d\n", p.Likes, p.Retweets, p.Replies, p.Views)
			if p.TweetURL != nil {
				fmt.Fprintf(&sb, "  URL: %s\n", *p.TweetURL)
			}
			sb.WriteString("\n")
		}
		return mcp.NewToolResultText(sb.String()), nil
	}
}

func makeGeneratePosts(c *di.MCPContainer) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		topicID, _ := req.GetArguments()["topic_id"].(string)
		styleStr, _ := req.GetArguments()["style"].(string)

		var style *entity.PostStyle
		if styleStr != "" {
			s := parsePostStyle(styleStr)
			style = &s
		}

		posts, err := c.PostUC.GeneratePosts(ctx, c.UserID, topicID, style)
		if err != nil {
			return errorResult(err), nil
		}

		var sb strings.Builder
		sb.WriteString("AI投稿文を生成しました:\n\n")
		for i, p := range posts {
			fmt.Fprintf(&sb, "--- %d ---\n%s\n\n", i+1, p.Content)
		}
		sb.WriteString("気に入った投稿を下書き保存するには create_draft ツールを使ってください。")
		return mcp.NewToolResultText(sb.String()), nil
	}
}

// --- helpers ---

func postStatusLabel(s entity.PostStatus) string {
	switch s {
	case entity.PostDraft:
		return "下書き"
	case entity.PostScheduled:
		return "予約済"
	case entity.PostPublished:
		return "投稿済"
	case entity.PostFailed:
		return "失敗"
	default:
		return "不明"
	}
}

func parsePostStyle(s string) entity.PostStyle {
	switch strings.ToLower(s) {
	case "breaking":
		return entity.PostStyleBreaking
	case "analysis":
		return entity.PostStyleAnalysis
	default:
		return entity.PostStyleCasual
	}
}

func truncate(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes]) + "..."
}

func stringPtr(v interface{}) *string {
	s, ok := v.(string)
	if !ok || s == "" {
		return nil
	}
	return &s
}
