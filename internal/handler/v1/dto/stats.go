package dto

import "eshkere/internal/service"

const dateLayout = "2006-01-02"

type StatsMetric struct {
	Impressions     int64   `json:"impressions"`
	Clicks          int64   `json:"clicks"`
	CTR             float64 `json:"ctr"`
	Spend           int64   `json:"spend"`
	CPC             float64 `json:"cpc"`
	PartnerReward   int64   `json:"partner_reward"`
	PlatformRevenue int64   `json:"platform_revenue"`
}

type StatsPeriod struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type StatsPoint struct {
	Date string `json:"date"`
	StatsMetric
}

type StatsEntityRow struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	StatsMetric
}

type StatsUnavailableMetric struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	Reason string `json:"reason"`
}

type CampaignStatsResponse struct {
	Period             StatsPeriod              `json:"period"`
	Totals             StatsMetric              `json:"totals"`
	PreviousTotals     StatsMetric              `json:"previous_totals"`
	Timeline           []StatsPoint             `json:"timeline"`
	Groups             []StatsEntityRow         `json:"groups"`
	Placements         []StatsEntityRow         `json:"placements"`
	UnavailableMetrics []StatsUnavailableMetric `json:"unavailable_metrics"`
}

type GroupStatsResponse struct {
	Period             StatsPeriod              `json:"period"`
	Totals             StatsMetric              `json:"totals"`
	PreviousTotals     StatsMetric              `json:"previous_totals"`
	Timeline           []StatsPoint             `json:"timeline"`
	Ads                []StatsEntityRow         `json:"ads"`
	Placements         []StatsEntityRow         `json:"placements"`
	UnavailableMetrics []StatsUnavailableMetric `json:"unavailable_metrics"`
}

type AdStatsResponse struct {
	Period             StatsPeriod              `json:"period"`
	Totals             StatsMetric              `json:"totals"`
	PreviousTotals     StatsMetric              `json:"previous_totals"`
	Timeline           []StatsPoint             `json:"timeline"`
	Placements         []StatsEntityRow         `json:"placements"`
	UnavailableMetrics []StatsUnavailableMetric `json:"unavailable_metrics"`
}

type PartnerIncomeRowResponse struct {
	Date        string `json:"date"`
	SiteID      int    `json:"site_id"`
	SiteName    string `json:"site_name"`
	Domain      string `json:"domain"`
	BlockID     int    `json:"block_id"`
	BlockName   string `json:"block_name"`
	Impressions int64  `json:"impressions"`
	Reward      int64  `json:"reward"`
}

type PartnerIncomeStatsResponse struct {
	From        string                     `json:"from"`
	To          string                     `json:"to"`
	Impressions int64                      `json:"impressions"`
	Reward      int64                      `json:"reward"`
	ECPM        float64                    `json:"ecpm"`
	Rows        []PartnerIncomeRowResponse `json:"rows"`
}

func ToCampaignStatsResponse(stats *service.CampaignStats) CampaignStatsResponse {
	return CampaignStatsResponse{
		Period:             toStatsPeriod(stats.Period),
		Totals:             toStatsMetric(stats.Totals),
		PreviousTotals:     toStatsMetric(stats.PreviousTotals),
		Timeline:           toStatsPoints(stats.Timeline),
		Groups:             toStatsEntityRows(stats.Groups),
		Placements:         toStatsEntityRows(stats.Placements),
		UnavailableMetrics: toStatsUnavailableMetrics(stats.UnavailableMetrics),
	}
}

func ToGroupStatsResponse(stats *service.GroupStats) GroupStatsResponse {
	return GroupStatsResponse{
		Period:             toStatsPeriod(stats.Period),
		Totals:             toStatsMetric(stats.Totals),
		PreviousTotals:     toStatsMetric(stats.PreviousTotals),
		Timeline:           toStatsPoints(stats.Timeline),
		Ads:                toStatsEntityRows(stats.Ads),
		Placements:         toStatsEntityRows(stats.Placements),
		UnavailableMetrics: toStatsUnavailableMetrics(stats.UnavailableMetrics),
	}
}

func ToAdStatsResponse(stats *service.AdStats) AdStatsResponse {
	return AdStatsResponse{
		Period:             toStatsPeriod(stats.Period),
		Totals:             toStatsMetric(stats.Totals),
		PreviousTotals:     toStatsMetric(stats.PreviousTotals),
		Timeline:           toStatsPoints(stats.Timeline),
		Placements:         toStatsEntityRows(stats.Placements),
		UnavailableMetrics: toStatsUnavailableMetrics(stats.UnavailableMetrics),
	}
}

func ToPartnerIncomeStatsResponse(stats *service.PartnerIncomeStats) PartnerIncomeStatsResponse {
	rows := make([]PartnerIncomeRowResponse, 0, len(stats.Rows))
	for _, row := range stats.Rows {
		rows = append(rows, PartnerIncomeRowResponse{
			Date:        row.Date.Format(dateLayout),
			SiteID:      row.SiteID,
			SiteName:    row.SiteName,
			Domain:      row.Domain,
			BlockID:     row.BlockID,
			BlockName:   row.BlockName,
			Impressions: row.Impressions,
			Reward:      row.Reward,
		})
	}

	return PartnerIncomeStatsResponse{
		From:        stats.From.Format(dateLayout),
		To:          stats.To.Format(dateLayout),
		Impressions: stats.Impressions,
		Reward:      stats.Reward,
		ECPM:        stats.ECPM,
		Rows:        rows,
	}
}

func toStatsPeriod(period service.StatsPeriod) StatsPeriod {
	return StatsPeriod{
		From: period.From.Format(dateLayout),
		To:   period.To.Format(dateLayout),
	}
}

func toStatsMetric(metric service.StatsMetric) StatsMetric {
	return StatsMetric{
		Impressions:     metric.Impressions,
		Clicks:          metric.Clicks,
		CTR:             metric.CTR,
		Spend:           metric.Spend,
		CPC:             metric.CPC,
		PartnerReward:   metric.PartnerReward,
		PlatformRevenue: metric.PlatformRevenue,
	}
}

func toStatsPoints(points []service.StatsPoint) []StatsPoint {
	out := make([]StatsPoint, 0, len(points))
	for _, point := range points {
		out = append(out, StatsPoint{
			Date:        point.Date.Format(dateLayout),
			StatsMetric: toStatsMetric(point.StatsMetric),
		})
	}
	return out
}

func toStatsEntityRows(rows []service.StatsEntityRow) []StatsEntityRow {
	out := make([]StatsEntityRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, StatsEntityRow{
			ID:          row.ID,
			Name:        row.Name,
			StatsMetric: toStatsMetric(row.StatsMetric),
		})
	}
	return out
}

func toStatsUnavailableMetrics(metrics []service.StatsUnavailableMetric) []StatsUnavailableMetric {
	out := make([]StatsUnavailableMetric, 0, len(metrics))
	for _, metric := range metrics {
		out = append(out, StatsUnavailableMetric{
			Key:    metric.Key,
			Label:  metric.Label,
			Reason: metric.Reason,
		})
	}
	return out
}
