package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/trendbird/backend/internal/di"
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/usecase"
)

func registerTopicTools(s *server.MCPServer, c *di.MCPContainer) {
	s.AddTool(mcp.NewTool("list_topics",
		mcp.WithDescription("監視中のトピック一覧を取得します（トレンド状態・Z-score付き）"),
	), makeListTopics(c))

	s.AddTool(mcp.NewTool("get_topic",
		mcp.WithDescription("トピックの詳細情報を取得します"),
		mcp.WithString("topic_id", mcp.Required(), mcp.Description("トピックID")),
	), makeGetTopic(c))

	s.AddTool(mcp.NewTool("create_topic",
		mcp.WithDescription("新しいトピックを作成して監視を開始します"),
		mcp.WithString("name", mcp.Required(), mcp.Description("トピック名")),
		mcp.WithString("keywords", mcp.Required(), mcp.Description("監視キーワード（カンマ区切り）")),
		mcp.WithString("genre", mcp.Required(), mcp.Description("ジャンル（technology, business, marketing, entertainment, sports, health, politics, science, education, lifestyle）")),
	), makeCreateTopic(c))

	s.AddTool(mcp.NewTool("delete_topic",
		mcp.WithDescription("トピックを削除して監視を停止します"),
		mcp.WithString("topic_id", mcp.Required(), mcp.Description("トピックID")),
	), makeDeleteTopic(c))
}

func makeListTopics(c *di.MCPContainer) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		topics, err := c.TopicUC.ListTopics(ctx, c.UserID)
		if err != nil {
			return errorResult(err), nil
		}

		if len(topics) == 0 {
			return mcp.NewToolResultText("監視中のトピックはありません。create_topic で追加できます。"), nil
		}

		var sb strings.Builder
		fmt.Fprintf(&sb, "監視中のトピック: %d件\n\n", len(topics))
		for _, t := range topics {
			status := topicStatusLabel(t.Status)
			zScore := "—"
			if t.ZScore != nil {
				zScore = fmt.Sprintf("%.1f", *t.ZScore)
			}
			fmt.Fprintf(&sb, "- [%s] %s (Z-score: %s, ID: %s)\n", status, t.Name, zScore, t.ID)
			fmt.Fprintf(&sb, "  キーワード: %s\n", strings.Join(t.Keywords, ", "))
		}
		return mcp.NewToolResultText(sb.String()), nil
	}
}

func makeGetTopic(c *di.MCPContainer) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		topicID, _ := args["topic_id"].(string)

		output, err := c.TopicUC.GetTopic(ctx, c.UserID, topicID)
		if err != nil {
			return errorResult(err), nil
		}

		t := output.Topic
		var sb strings.Builder
		status := topicStatusLabel(t.Status)
		zScore := "—"
		if t.ZScore != nil {
			zScore = fmt.Sprintf("%.1f", *t.ZScore)
		}
		fmt.Fprintf(&sb, "トピック: %s\n", t.Name)
		fmt.Fprintf(&sb, "状態: %s / Z-score: %s\n", status, zScore)
		fmt.Fprintf(&sb, "キーワード: %s\n", strings.Join(t.Keywords, ", "))
		fmt.Fprintf(&sb, "変動率: %.1f%% / 現在量: %d / 基準量: %d\n", t.ChangePercent*100, t.CurrentVolume, t.BaselineVolume)
		if t.ContextSummary != nil {
			fmt.Fprintf(&sb, "背景: %s\n", *t.ContextSummary)
		}
		fmt.Fprintf(&sb, "ID: %s\n", t.ID)

		return mcp.NewToolResultText(sb.String()), nil
	}
}

func makeCreateTopic(c *di.MCPContainer) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		name, _ := args["name"].(string)
		keywordsStr, _ := args["keywords"].(string)
		genre, _ := args["genre"].(string)

		keywords := splitAndTrim(keywordsStr)

		topic, err := c.TopicUC.CreateTopic(ctx, c.UserID, usecase.CreateTopicInput{
			Name:     name,
			Keywords: keywords,
			Genre:    genre,
		})
		if err != nil {
			return errorResult(err), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("トピックを作成しました。\n名前: %s\nキーワード: %s\nID: %s",
			topic.Name, strings.Join(topic.Keywords, ", "), topic.ID)), nil
	}
}

func makeDeleteTopic(c *di.MCPContainer) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		topicID, _ := args["topic_id"].(string)

		if err := c.TopicUC.DeleteTopic(ctx, c.UserID, topicID); err != nil {
			return errorResult(err), nil
		}

		return mcp.NewToolResultText("トピックを削除しました。"), nil
	}
}

// --- helpers ---

func topicStatusLabel(s entity.TopicStatus) string {
	switch s {
	case entity.TopicStable:
		return "安定"
	case entity.TopicRising:
		return "上昇中"
	case entity.TopicSpike:
		return "スパイク"
	default:
		return "不明"
	}
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
