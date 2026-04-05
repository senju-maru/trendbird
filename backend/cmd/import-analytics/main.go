package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/infrastructure/config"
	"github.com/trendbird/backend/internal/infrastructure/persistence"
	"github.com/trendbird/backend/internal/infrastructure/persistence/repository"
)

func main() {
	if len(os.Args) < 4 {
		log.Fatal("Usage: import-analytics <daily|post> <csv-file> <user-id>")
	}

	mode := os.Args[1]
	csvFile := os.Args[2]
	userID := os.Args[3]

	cfg, err := config.LoadMCP()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := persistence.NewMCPDB(cfg)
	if err != nil {
		log.Fatalf("db: %v", err)
	}

	content, err := os.ReadFile(csvFile)
	if err != nil {
		log.Fatalf("read file: %v", err)
	}

	ctx := context.Background()

	switch mode {
	case "daily":
		repo := repository.NewXAnalyticsDailyRepository(db)
		records, err := parseDailyCSV(string(content))
		if err != nil {
			log.Fatalf("parse: %v", err)
		}
		for _, r := range records {
			r.UserID = userID
		}
		inserted, updated, err := repo.UpsertBatch(ctx, records)
		if err != nil {
			log.Fatalf("upsert: %v", err)
		}
		fmt.Printf("Daily analytics: inserted=%d, updated=%d\n", inserted, updated)

	case "post":
		repo := repository.NewXAnalyticsPostRepository(db)
		records, err := parsePostCSV(string(content))
		if err != nil {
			log.Fatalf("parse: %v", err)
		}
		for _, r := range records {
			r.UserID = userID
		}
		const batchSize = 50
		var totalInserted, totalUpdated int32
		for i := 0; i < len(records); i += batchSize {
			end := i + batchSize
			if end > len(records) {
				end = len(records)
			}
			ins, upd, err := repo.UpsertBatch(ctx, records[i:end])
			if err != nil {
				log.Fatalf("upsert batch %d-%d: %v", i, end, err)
			}
			totalInserted += ins
			totalUpdated += upd
		}
		fmt.Printf("Post analytics: inserted=%d, updated=%d (total %d records)\n", totalInserted, totalUpdated, len(records))

	default:
		log.Fatalf("unknown mode: %s (use 'daily' or 'post')", mode)
	}
}

func stripBOM(s string) string { return strings.TrimPrefix(s, "\xef\xbb\xbf") }

func buildHeaderMap(header []string) map[string]int {
	m := make(map[string]int, len(header))
	for i, h := range header {
		m[strings.TrimSpace(strings.TrimRight(h, "\\"))] = i
	}
	return m
}

func getField(row []string, hm map[string]int, key string) string {
	if idx, ok := hm[key]; ok && idx < len(row) {
		return strings.TrimSpace(row[idx])
	}
	return ""
}

func getInt(row []string, hm map[string]int, key string) int32 {
	v, _ := strconv.Atoi(getField(row, hm, key))
	return int32(v)
}

func parseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if t, err := time.Parse("Mon, Jan 02, 2006", s); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("bad date: %q", s)
}

func parseDailyCSV(content string) ([]*entity.XAnalyticsDaily, error) {
	r := csv.NewReader(strings.NewReader(stripBOM(content)))
	r.LazyQuotes = true
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("no data")
	}
	hm := buildHeaderMap(rows[0])
	out := make([]*entity.XAnalyticsDaily, 0, len(rows)-1)
	for _, row := range rows[1:] {
		d, err := parseDate(getField(row, hm, "Date"))
		if err != nil {
			continue
		}
		out = append(out, &entity.XAnalyticsDaily{
			Date: d, Impressions: getInt(row, hm, "インプレッション数"), Likes: getInt(row, hm, "いいね"),
			Engagements: getInt(row, hm, "エンゲージメント"), Bookmarks: getInt(row, hm, "ブックマーク"),
			Shares: getInt(row, hm, "共有された回数"), NewFollows: getInt(row, hm, "新しいフォロー"),
			Unfollows: getInt(row, hm, "フォロー解除"), Replies: getInt(row, hm, "返信"),
			Reposts: getInt(row, hm, "リポスト"), ProfileVisits: getInt(row, hm, "プロフィールへのアクセス数"),
			PostsCreated: getInt(row, hm, "ポストを作成"), VideoViews: getInt(row, hm, "動画再生数"),
			MediaViews: getInt(row, hm, "メディアの再生数"),
		})
	}
	return out, nil
}

func parsePostCSV(content string) ([]*entity.XAnalyticsPost, error) {
	r := csv.NewReader(strings.NewReader(stripBOM(content)))
	r.LazyQuotes = true
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("no data")
	}
	hm := buildHeaderMap(rows[0])
	out := make([]*entity.XAnalyticsPost, 0, len(rows)-1)
	for _, row := range rows[1:] {
		pid := getField(row, hm, "ポストID")
		if pid == "" {
			continue
		}
		d, err := parseDate(getField(row, hm, "日付"))
		if err != nil {
			continue
		}
		out = append(out, &entity.XAnalyticsPost{
			PostID: pid, PostedAt: d, PostText: getField(row, hm, "ポスト本文"),
			PostURL: getField(row, hm, "ポストのリンク"), Impressions: getInt(row, hm, "インプレッション数"),
			Likes: getInt(row, hm, "いいね"), Engagements: getInt(row, hm, "エンゲージメント"),
			Bookmarks: getInt(row, hm, "ブックマーク"), Shares: getInt(row, hm, "共有された回数"),
			NewFollows: getInt(row, hm, "新しいフォロー"), Replies: getInt(row, hm, "返信"),
			Reposts: getInt(row, hm, "リポスト"), ProfileVisits: getInt(row, hm, "プロフィールへのアクセス数"),
			DetailClicks: getInt(row, hm, "詳細のクリック数"), URLClicks: getInt(row, hm, "URLのクリック数"),
			HashtagClicks: getInt(row, hm, "ハッシュタグのクリック数"), PermalinkClicks: getInt(row, hm, "パーマリンクのクリック数"),
		})
	}
	return out, nil
}
