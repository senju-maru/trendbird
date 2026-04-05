package mcp

import (
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/trendbird/backend/internal/di"
	"github.com/trendbird/backend/internal/domain/entity"
)

func registerAnalyticsTools(s *server.MCPServer, c *di.MCPContainer) {
	s.AddTool(mcp.NewTool("import_daily_analytics",
		mcp.WithDescription("X アナリティクスの日次オーバービュー CSV データを取り込みます"),
		mcp.WithString("csv_content", mcp.Required(), mcp.Description("CSV ファイルの内容（ヘッダー行を含む）")),
	), makeImportDailyAnalytics(c))

	s.AddTool(mcp.NewTool("import_post_analytics",
		mcp.WithDescription("X アナリティクスの投稿別コンテンツ CSV データを取り込みます"),
		mcp.WithString("csv_content", mcp.Required(), mcp.Description("CSV ファイルの内容（ヘッダー行を含む）")),
	), makeImportPostAnalytics(c))

	s.AddTool(mcp.NewTool("get_analytics_summary",
		mcp.WithDescription("X アナリティクスの期間サマリーを取得します"),
		mcp.WithString("start_date", mcp.Description("開始日（YYYY-MM-DD、省略時は30日前）")),
		mcp.WithString("end_date", mcp.Description("終了日（YYYY-MM-DD、省略時は今日）")),
	), makeGetAnalyticsSummary(c))

	s.AddTool(mcp.NewTool("list_top_posts",
		mcp.WithDescription("パフォーマンスの高い投稿を一覧表示します"),
		mcp.WithString("sort_by", mcp.Description("ソート基準: impressions, likes, engagements, new_follows（デフォルト: impressions）")),
		mcp.WithNumber("limit", mcp.Description("取得件数（デフォルト: 10、最大: 100）")),
		mcp.WithString("start_date", mcp.Description("開始日（YYYY-MM-DD）")),
		mcp.WithString("end_date", mcp.Description("終了日（YYYY-MM-DD）")),
	), makeListTopPosts(c))

	s.AddTool(mcp.NewTool("get_growth_insights",
		mcp.WithDescription("アナリティクスデータに基づく成長インサイトと推奨アクションを取得します"),
		mcp.WithString("start_date", mcp.Description("開始日（YYYY-MM-DD、省略時は30日前）")),
		mcp.WithString("end_date", mcp.Description("終了日（YYYY-MM-DD、省略時は今日）")),
	), makeGetGrowthInsights(c))
}

func makeImportDailyAnalytics(c *di.MCPContainer) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		csvContent, _ := req.GetArguments()["csv_content"].(string)
		if csvContent == "" {
			return mcp.NewToolResultError("csv_content は必須です"), nil
		}

		records, err := parseDailyCSV(csvContent)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("CSV パースエラー: %v", err)), nil
		}

		inserted, updated, err := c.AnalyticsUC.ImportDailyAnalytics(ctx, c.UserID, records)
		if err != nil {
			return errorResult(err), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("日次アナリティクスを取り込みました。\n新規: %d件 / 更新: %d件", inserted, updated)), nil
	}
}

func makeImportPostAnalytics(c *di.MCPContainer) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		csvContent, _ := req.GetArguments()["csv_content"].(string)
		if csvContent == "" {
			return mcp.NewToolResultError("csv_content は必須です"), nil
		}

		records, err := parsePostCSV(csvContent)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("CSV パースエラー: %v", err)), nil
		}

		inserted, updated, err := c.AnalyticsUC.ImportPostAnalytics(ctx, c.UserID, records)
		if err != nil {
			return errorResult(err), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("投稿別アナリティクスを取り込みました。\n新規: %d件 / 更新: %d件", inserted, updated)), nil
	}
}

func makeGetAnalyticsSummary(c *di.MCPContainer) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		startDate := stringPtr(req.GetArguments()["start_date"])
		endDate := stringPtr(req.GetArguments()["end_date"])

		summary, err := c.AnalyticsUC.GetAnalyticsSummary(ctx, c.UserID, startDate, endDate)
		if err != nil {
			return errorResult(err), nil
		}

		var sb strings.Builder
		fmt.Fprintf(&sb, "📊 アナリティクスサマリー（%s 〜 %s）\n\n", summary.StartDate, summary.EndDate)
		fmt.Fprintf(&sb, "日数: %d日\n", summary.DaysCount)
		fmt.Fprintf(&sb, "合計インプレッション: %d\n", summary.TotalImpressions)
		fmt.Fprintf(&sb, "合計いいね: %d\n", summary.TotalLikes)
		fmt.Fprintf(&sb, "合計エンゲージメント: %d\n", summary.TotalEngagements)
		fmt.Fprintf(&sb, "新規フォロー: %d\n", summary.TotalNewFollows)
		fmt.Fprintf(&sb, "フォロー解除: %d\n", summary.TotalUnfollows)
		fmt.Fprintf(&sb, "純フォロー増: %d\n", summary.TotalNewFollows-summary.TotalUnfollows)

		if summary.TotalImpressions > 0 {
			engRate := float64(summary.TotalEngagements) / float64(summary.TotalImpressions) * 100
			fmt.Fprintf(&sb, "エンゲージメント率: %.2f%%\n", engRate)
		}

		if len(summary.DailyData) > 0 {
			fmt.Fprintf(&sb, "\n日別データ:\n")
			for _, d := range summary.DailyData {
				fmt.Fprintf(&sb, "  %s: %d imp / %d いいね / %d eng / +%d follow\n",
					d.Date.Format("01/02"), d.Impressions, d.Likes, d.Engagements, d.NewFollows)
			}
		}

		return mcp.NewToolResultText(sb.String()), nil
	}
}

func makeListTopPosts(c *di.MCPContainer) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sortBy := stringPtr(req.GetArguments()["sort_by"])
		startDate := stringPtr(req.GetArguments()["start_date"])
		endDate := stringPtr(req.GetArguments()["end_date"])

		var limit *int32
		if v, ok := req.GetArguments()["limit"].(float64); ok {
			l := int32(v)
			limit = &l
		}

		posts, total, err := c.AnalyticsUC.ListPostAnalytics(ctx, c.UserID, sortBy, limit, startDate, endDate)
		if err != nil {
			return errorResult(err), nil
		}

		if len(posts) == 0 {
			return mcp.NewToolResultText("投稿アナリティクスデータがありません。"), nil
		}

		var sb strings.Builder
		fmt.Fprintf(&sb, "投稿パフォーマンス（全%d件中 上位%d件）\n\n", total, len(posts))

		for i, p := range posts {
			text := truncate(p.PostText, 50)
			fmt.Fprintf(&sb, "%d. 「%s」\n", i+1, text)
			fmt.Fprintf(&sb, "   %s | %d imp | %d いいね | %d eng | +%d follow | %d BM\n",
				p.PostedAt.Format("01/02"), p.Impressions, p.Likes, p.Engagements, p.NewFollows, p.Bookmarks)
			if p.PostURL != "" {
				fmt.Fprintf(&sb, "   %s\n", p.PostURL)
			}
			fmt.Fprintln(&sb)
		}

		return mcp.NewToolResultText(sb.String()), nil
	}
}

func makeGetGrowthInsights(c *di.MCPContainer) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		startDate := stringPtr(req.GetArguments()["start_date"])
		endDate := stringPtr(req.GetArguments()["end_date"])

		insights, summary, err := c.AnalyticsUC.GetGrowthInsights(ctx, c.UserID, startDate, endDate)
		if err != nil {
			return errorResult(err), nil
		}

		var sb strings.Builder
		fmt.Fprintf(&sb, "📈 成長インサイト（%s 〜 %s）\n\n", summary.StartDate, summary.EndDate)

		for i, ins := range insights {
			fmt.Fprintf(&sb, "%d. [%s] %s\n", i+1, ins.Category, ins.Insight)
			fmt.Fprintf(&sb, "   → %s\n\n", ins.Action)
		}

		return mcp.NewToolResultText(sb.String()), nil
	}
}

// --- CSV Parsing ---

// parseDailyCSV parses the daily overview CSV content into entities.
func parseDailyCSV(content string) ([]*entity.XAnalyticsDaily, error) {
	content = stripBOM(content)
	reader := csv.NewReader(strings.NewReader(content))
	reader.LazyQuotes = true

	allRows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("CSV 読み取りエラー: %w", err)
	}
	if len(allRows) < 2 {
		return nil, fmt.Errorf("データ行がありません")
	}

	// Map header columns by Japanese name
	headerMap := buildHeaderMap(allRows[0])

	records := make([]*entity.XAnalyticsDaily, 0, len(allRows)-1)
	for i, row := range allRows[1:] {
		date, err := parseDateField(getField(row, headerMap, "Date"))
		if err != nil {
			return nil, fmt.Errorf("行 %d: 日付パースエラー: %w", i+2, err)
		}

		records = append(records, &entity.XAnalyticsDaily{
			Date:          date,
			Impressions:   int32(getIntField(row, headerMap, "インプレッション数")),
			Likes:         int32(getIntField(row, headerMap, "いいね")),
			Engagements:   int32(getIntField(row, headerMap, "エンゲージメント")),
			Bookmarks:     int32(getIntField(row, headerMap, "ブックマーク")),
			Shares:        int32(getIntField(row, headerMap, "共有された回数")),
			NewFollows:    int32(getIntField(row, headerMap, "新しいフォロー")),
			Unfollows:     int32(getIntField(row, headerMap, "フォロー解除")),
			Replies:       int32(getIntField(row, headerMap, "返信")),
			Reposts:       int32(getIntField(row, headerMap, "リポスト")),
			ProfileVisits: int32(getIntField(row, headerMap, "プロフィールへのアクセス数")),
			PostsCreated:  int32(getIntField(row, headerMap, "ポストを作成")),
			VideoViews:    int32(getIntField(row, headerMap, "動画再生数")),
			MediaViews:    int32(getIntField(row, headerMap, "メディアの再生数")),
		})
	}

	return records, nil
}

// parsePostCSV parses the per-post content CSV into entities.
func parsePostCSV(content string) ([]*entity.XAnalyticsPost, error) {
	content = stripBOM(content)
	reader := csv.NewReader(strings.NewReader(content))
	reader.LazyQuotes = true

	allRows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("CSV 読み取りエラー: %w", err)
	}
	if len(allRows) < 2 {
		return nil, fmt.Errorf("データ行がありません")
	}

	headerMap := buildHeaderMap(allRows[0])

	records := make([]*entity.XAnalyticsPost, 0, len(allRows)-1)
	for i, row := range allRows[1:] {
		postID := getField(row, headerMap, "ポストID")
		if postID == "" {
			continue
		}

		postedAt, err := parseDateField(getField(row, headerMap, "日付"))
		if err != nil {
			return nil, fmt.Errorf("行 %d: 日付パースエラー: %w", i+2, err)
		}

		records = append(records, &entity.XAnalyticsPost{
			PostID:          postID,
			PostedAt:        postedAt,
			PostText:        getField(row, headerMap, "ポスト本文"),
			PostURL:         getField(row, headerMap, "ポストのリンク"),
			Impressions:     int32(getIntField(row, headerMap, "インプレッション数")),
			Likes:           int32(getIntField(row, headerMap, "いいね")),
			Engagements:     int32(getIntField(row, headerMap, "エンゲージメント")),
			Bookmarks:       int32(getIntField(row, headerMap, "ブックマーク")),
			Shares:          int32(getIntField(row, headerMap, "共有された回数")),
			NewFollows:      int32(getIntField(row, headerMap, "新しいフォロー")),
			Replies:         int32(getIntField(row, headerMap, "返信")),
			Reposts:         int32(getIntField(row, headerMap, "リポスト")),
			ProfileVisits:   int32(getIntField(row, headerMap, "プロフィールへのアクセス数")),
			DetailClicks:    int32(getIntField(row, headerMap, "詳細のクリック数")),
			URLClicks:       int32(getIntField(row, headerMap, "URLのクリック数")),
			HashtagClicks:   int32(getIntField(row, headerMap, "ハッシュタグのクリック数")),
			PermalinkClicks: int32(getIntField(row, headerMap, "パーマリンクのクリック数")),
		})
	}

	return records, nil
}

// --- CSV Helpers ---

func stripBOM(s string) string {
	return strings.TrimPrefix(s, "\xef\xbb\xbf")
}

func buildHeaderMap(header []string) map[string]int {
	m := make(map[string]int, len(header))
	for i, h := range header {
		// Remove backslash artifacts from CSV header (e.g. "共有された回数\")
		h = strings.TrimRight(h, "\\")
		h = strings.TrimSpace(h)
		m[h] = i
	}
	return m
}

func getField(row []string, headerMap map[string]int, key string) string {
	idx, ok := headerMap[key]
	if !ok || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

func getIntField(row []string, headerMap map[string]int, key string) int {
	s := getField(row, headerMap, key)
	if s == "" {
		return 0
	}
	v, _ := strconv.Atoi(s)
	return v
}

func parseDateField(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}

	// Try "Mon, Jan 02, 2006" format (e.g. "Tue, Mar 31, 2026")
	if t, err := time.Parse("Mon, Jan 02, 2006", s); err == nil {
		return t, nil
	}

	// Try ISO date "2006-01-02"
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}

	// Try RFC3339
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unsupported date format: %q", s)
}
