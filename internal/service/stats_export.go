package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"time"
)

// ExportCampaignStatsCSV возвращает CSV-отчёт по кампании (только для Pro).
func (s *Service) ExportCampaignStatsCSV(ctx context.Context, advertiserID, campaignID int, from, to time.Time) ([]byte, error) {
	if _, err := s.RequireProActive(ctx, advertiserID); err != nil {
		return nil, err
	}

	stats, err := s.GetCampaignStats(ctx, advertiserID, campaignID, from, to)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	_ = w.Write([]string{"period_from", stats.Period.From.Format("2006-01-02")})
	_ = w.Write([]string{"period_to", stats.Period.To.Format("2006-01-02")})
	_ = w.Write([]string{
		"impressions", strconv.FormatInt(stats.Totals.Impressions, 10),
	})
	_ = w.Write([]string{"clicks", strconv.FormatInt(stats.Totals.Clicks, 10)})
	_ = w.Write([]string{"ctr", fmt.Sprintf("%.4f", stats.Totals.CTR)})
	_ = w.Write([]string{"spend", strconv.FormatInt(stats.Totals.Spend, 10)})
	_ = w.Write([]string{"cpc", fmt.Sprintf("%.4f", stats.Totals.CPC)})
	_ = w.Write([]string{"partner_reward", strconv.FormatInt(stats.Totals.PartnerReward, 10)})
	_ = w.Write([]string{"platform_revenue", strconv.FormatInt(stats.Totals.PlatformRevenue, 10)})
	_ = w.Write([]string{})

	_ = w.Write([]string{"date", "impressions", "clicks", "ctr", "spend", "cpc"})
	for _, point := range stats.Timeline {
		_ = w.Write([]string{
			point.Date.Format("2006-01-02"),
			strconv.FormatInt(point.Impressions, 10),
			strconv.FormatInt(point.Clicks, 10),
			fmt.Sprintf("%.4f", point.CTR),
			strconv.FormatInt(point.Spend, 10),
			fmt.Sprintf("%.4f", point.CPC),
		})
	}

	_ = w.Write([]string{})
	_ = w.Write([]string{"group_id", "group_name", "impressions", "clicks", "ctr", "spend"})
	for _, row := range stats.Groups {
		_ = w.Write([]string{
			strconv.Itoa(row.ID),
			row.Name,
			strconv.FormatInt(row.Impressions, 10),
			strconv.FormatInt(row.Clicks, 10),
			fmt.Sprintf("%.4f", row.CTR),
			strconv.FormatInt(row.Spend, 10),
		})
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("write campaign stats csv: %w", err)
	}

	return buf.Bytes(), nil
}
