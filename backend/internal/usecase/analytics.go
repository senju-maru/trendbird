package usecase

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/trendbird/backend/internal/domain/apperror"
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/domain/repository"
)

// AnalyticsUsecase handles analytics-related operations.
type AnalyticsUsecase struct {
	dailyRepo repository.XAnalyticsDailyRepository
	postRepo  repository.XAnalyticsPostRepository
}

// NewAnalyticsUsecase creates a new AnalyticsUsecase.
func NewAnalyticsUsecase(
	dailyRepo repository.XAnalyticsDailyRepository,
	postRepo repository.XAnalyticsPostRepository,
) *AnalyticsUsecase {
	return &AnalyticsUsecase{
		dailyRepo: dailyRepo,
		postRepo:  postRepo,
	}
}

// ImportDailyAnalytics imports daily analytics records (upsert).
func (u *AnalyticsUsecase) ImportDailyAnalytics(ctx context.Context, userID string, records []*entity.XAnalyticsDaily) (inserted, updated int32, err error) {
	if len(records) == 0 {
		return 0, 0, apperror.InvalidArgument("records must not be empty")
	}
	for _, rec := range records {
		rec.UserID = userID
	}
	return u.dailyRepo.UpsertBatch(ctx, records)
}

// ImportPostAnalytics imports per-post analytics records (upsert).
func (u *AnalyticsUsecase) ImportPostAnalytics(ctx context.Context, userID string, records []*entity.XAnalyticsPost) (inserted, updated int32, err error) {
	if len(records) == 0 {
		return 0, 0, apperror.InvalidArgument("records must not be empty")
	}
	for _, rec := range records {
		rec.UserID = userID
	}
	return u.postRepo.UpsertBatch(ctx, records)
}

// GetAnalyticsSummary returns aggregated summary for a date range.
func (u *AnalyticsUsecase) GetAnalyticsSummary(ctx context.Context, userID string, startDate, endDate *string) (*entity.AnalyticsSummary, error) {
	start, end := defaultDateRange(startDate, endDate)
	return u.dailyRepo.GetSummary(ctx, userID, start, end)
}

// ListPostAnalytics returns per-post analytics sorted by the given column.
func (u *AnalyticsUsecase) ListPostAnalytics(ctx context.Context, userID string, sortBy *string, limit *int32, startDate, endDate *string) ([]*entity.XAnalyticsPost, int64, error) {
	sb := "impressions"
	if sortBy != nil && *sortBy != "" {
		sb = *sortBy
	}
	lim := 20
	if limit != nil && *limit > 0 {
		lim = int(*limit)
	}
	if lim > 100 {
		lim = 100
	}

	var start, end *time.Time
	if startDate != nil && *startDate != "" {
		t, err := time.Parse("2006-01-02", *startDate)
		if err != nil {
			return nil, 0, apperror.InvalidArgument("invalid start_date format")
		}
		start = &t
	}
	if endDate != nil && *endDate != "" {
		t, err := time.Parse("2006-01-02", *endDate)
		if err != nil {
			return nil, 0, apperror.InvalidArgument("invalid end_date format")
		}
		eod := t.Add(24*time.Hour - time.Second)
		end = &eod
	}

	return u.postRepo.ListByUserID(ctx, userID, sb, lim, start, end)
}

// GetGrowthInsights generates data-driven insights from analytics data.
func (u *AnalyticsUsecase) GetGrowthInsights(ctx context.Context, userID string, startDate, endDate *string) ([]*entity.GrowthInsight, *entity.AnalyticsSummary, error) {
	start, end := defaultDateRange(startDate, endDate)
	summary, err := u.dailyRepo.GetSummary(ctx, userID, start, end)
	if err != nil {
		return nil, nil, err
	}

	topPosts, _, err := u.postRepo.ListByUserID(ctx, userID, "impressions", 50, &start, &end)
	if err != nil {
		return nil, nil, err
	}

	topBM, _, err := u.postRepo.ListByUserID(ctx, userID, "bookmarks", 10, &start, &end)
	if err != nil {
		return nil, nil, err
	}

	insights := generateInsights(summary, topPosts, topBM)
	return insights, summary, nil
}

func defaultDateRange(startDate, endDate *string) (time.Time, time.Time) {
	now := time.Now()
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	start := end.AddDate(0, 0, -30)

	if startDate != nil && *startDate != "" {
		if t, err := time.Parse("2006-01-02", *startDate); err == nil {
			start = t
		}
	}
	if endDate != nil && *endDate != "" {
		if t, err := time.Parse("2006-01-02", *endDate); err == nil {
			end = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
		}
	}
	return start, end
}

func generateInsights(summary *entity.AnalyticsSummary, topPosts []*entity.XAnalyticsPost, topBM []*entity.XAnalyticsPost) []*entity.GrowthInsight {
	var insights []*entity.GrowthInsight

	if summary.DaysCount == 0 {
		return []*entity.GrowthInsight{{
			Category: "data",
			Insight:  "アナリティクスデータがまだありません。",
			Action:   "X アナリティクスの CSV をインポートしてください。",
		}}
	}

	// --- 1. リプライ vs オリジナル分析 ---
	var replyCount, origCount int
	var replyImp, origImp int64
	var replyBM, origBM int32
	for _, p := range topPosts {
		isReply := len(p.PostText) > 0 && p.PostText[0] == '@'
		if isReply {
			replyCount++
			replyImp += int64(p.Impressions)
			replyBM += p.Bookmarks
		} else {
			origCount++
			origImp += int64(p.Impressions)
			origBM += p.Bookmarks
		}
	}
	if replyCount > 0 && origCount > 0 {
		replyAvg := replyImp / int64(replyCount)
		origAvg := origImp / int64(origCount)
		if replyAvg > origAvg {
			ratio := float64(replyAvg) / float64(origAvg)
			insights = append(insights, &entity.GrowthInsight{
				Category: "content",
				Insight:  fmt.Sprintf("リプライの平均インプレッション（%d）はオリジナル投稿（%d）の %.1f 倍。リプライ %d件 vs オリジナル %d件。", replyAvg, origAvg, ratio, replyCount, origCount),
				Action:   "影響力のあるアカウントへの知見を加えたリプライが最大のリーチ源。社交的な挨拶リプライではなく、具体的な経験・データ・代替案を含むリプライに集中しましょう。",
			})
		} else {
			insights = append(insights, &entity.GrowthInsight{
				Category: "content",
				Insight:  fmt.Sprintf("オリジナル投稿の平均インプレッション（%d）がリプライ（%d）を上回っています。", origAvg, replyAvg),
				Action:   "自発的な投稿がリーチを稼いでいます。テーマを深堀りした投稿を増やしましょう。",
			})
		}
	}

	// --- 2. ブックマーク（ストック価値）分析 ---
	var bmPosts []*entity.XAnalyticsPost
	for _, p := range topBM {
		if p.Bookmarks > 0 {
			bmPosts = append(bmPosts, p)
		}
	}
	if len(bmPosts) > 0 {
		best := bmPosts[0]
		text := best.PostText
		if len(text) > 60 {
			text = text[:60] + "..."
		}
		insights = append(insights, &entity.GrowthInsight{
			Category: "content",
			Insight:  fmt.Sprintf("ブックマーク最多: 「%s」（%d BM / %d imp）。ブックマーク付き投稿 %d件。", text, best.Bookmarks, best.Impressions, len(bmPosts)),
			Action:   "ブックマーク = 保存する価値がある投稿。ツール名+具体手順、リスク指摘、比較情報など「あとで見返したい」内容を週2本以上投稿しましょう。",
		})
	}

	// --- 3. エンゲージメント率 ---
	if summary.TotalImpressions > 0 {
		engRate := float64(summary.TotalEngagements) / float64(summary.TotalImpressions) * 100
		insights = append(insights, &entity.GrowthInsight{
			Category: "engagement",
			Insight:  fmt.Sprintf("エンゲージメント率 %.2f%%（%d eng / %d imp）。", engRate, summary.TotalEngagements, summary.TotalImpressions),
			Action: func() string {
				if engRate < 2.0 {
					return "質問形式の投稿や具体的な体験談を増やして、リプライとリアクションを促しましょう。"
				}
				return "良好なエンゲージメント率です。現在のコンテンツ戦略を維持しましょう。"
			}(),
		})
	}

	// --- 4. フォロー成長分析 ---
	netFollows := summary.TotalNewFollows - summary.TotalUnfollows
	avgDaily := float64(netFollows) / float64(summary.DaysCount)
	insights = append(insights, &entity.GrowthInsight{
		Category: "growth",
		Insight:  fmt.Sprintf("フォロワー純増 +%d（+%.1f/日）。新規 %d / 解除 %d。", netFollows, avgDaily, summary.TotalNewFollows, summary.TotalUnfollows),
		Action: func() string {
			if avgDaily >= 10 {
				return "良好なフォロー成長です。フォロワーが増えた日の投稿パターンを分析して再現しましょう。"
			}
			return "プロフィールへの訪問→フォローの導線を強化。自己紹介文の最適化と、固定ツイートの見直しを検討しましょう。"
		}(),
	})

	// --- 5. 曜日別分析 ---
	if len(summary.DailyData) >= 7 {
		dowImp := make(map[int][]int32)
		for _, d := range summary.DailyData {
			dow := int(d.Date.Weekday())
			dowImp[dow] = append(dowImp[dow], d.Impressions)
		}
		type dowAvg struct {
			dow int
			avg int64
		}
		var avgs []dowAvg
		for dow, imps := range dowImp {
			var sum int64
			for _, v := range imps {
				sum += int64(v)
			}
			avgs = append(avgs, dowAvg{dow, sum / int64(len(imps))})
		}
		sort.Slice(avgs, func(i, j int) bool { return avgs[i].avg > avgs[j].avg })
		dowNames := []string{"日", "月", "火", "水", "木", "金", "土"}
		if len(avgs) >= 2 {
			best := avgs[0]
			worst := avgs[len(avgs)-1]
			insights = append(insights, &entity.GrowthInsight{
				Category: "timing",
				Insight:  fmt.Sprintf("最もリーチが高い曜日: %s曜日（平均 %d imp）。最低: %s曜日（平均 %d imp）。", dowNames[best.dow], best.avg, dowNames[worst.dow], worst.avg),
				Action:   fmt.Sprintf("重要な投稿は%s曜日に集中させましょう。%s曜日はリプライ巡回や予約投稿で補いましょう。", dowNames[best.dow], dowNames[worst.dow]),
			})
		}
	}

	// --- 6. ベストデイ ---
	if len(summary.DailyData) > 0 {
		sorted := make([]*entity.XAnalyticsDaily, len(summary.DailyData))
		copy(sorted, summary.DailyData)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Impressions > sorted[j].Impressions
		})
		best := sorted[0]
		insights = append(insights, &entity.GrowthInsight{
			Category: "timing",
			Insight:  fmt.Sprintf("最高パフォーマンス日: %s（%d imp / %d likes / %d eng / +%d follow）。", best.Date.Format("2006-01-02"), best.Impressions, best.Likes, best.Engagements, best.NewFollows),
			Action:   "この日の投稿内容を確認し、バズの要因（テーマ・フック・投稿時間）を次の投稿に活かしましょう。",
		})
	}

	// --- 7. 次の投稿アクション提案 ---
	insights = append(insights, &entity.GrowthInsight{
		Category: "next_action",
		Insight:  "データに基づく次のアクション提案",
		Action: func() string {
			var actions []string
			// ブックマーク投稿が少ない場合
			if len(bmPosts) < 3 {
				actions = append(actions, "ストック型コンテンツ（ツール比較・手順解説・リスク指摘）を作成して保存価値のある投稿を増やす")
			}
			// リプライ依存度が高い場合
			if replyCount > 0 && origCount > 0 && replyImp > origImp*3 {
				actions = append(actions, "オリジナル投稿のフック（冒頭の引き）を改善してリプライ依存から脱却する")
			}
			// エンゲージメントが低い場合
			if summary.TotalImpressions > 0 {
				rate := float64(summary.TotalEngagements) / float64(summary.TotalImpressions) * 100
				if rate < 2.0 {
					actions = append(actions, "投稿の末尾に質問や意見を求めるCTAを追加してエンゲージメントを促す")
				}
			}
			if len(actions) == 0 {
				actions = append(actions, "現在の戦略を継続しつつ、週1回のデータ振り返りで微調整を続ける")
			}
			return fmt.Sprintf("1. %s", joinActions(actions))
		}(),
	})

	return insights
}

func joinActions(actions []string) string {
	result := actions[0]
	for i := 1; i < len(actions); i++ {
		result += fmt.Sprintf("\n%d. %s", i+1, actions[i])
	}
	return result
}
