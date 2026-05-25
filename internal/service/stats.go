package service

import (
	"context"
	"fmt"
	"math"
	"time"

	errs "eshkere/internal/errors"
)

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

type CampaignStats struct {
	Period             StatsPeriod              `json:"period"`
	Totals             StatsMetric              `json:"totals"`
	PreviousTotals     StatsMetric              `json:"previous_totals"`
	Timeline           []StatsPoint             `json:"timeline"`
	Groups             []StatsEntityRow         `json:"groups"`
	Placements         []StatsEntityRow         `json:"placements"`
	UnavailableMetrics []StatsUnavailableMetric `json:"unavailable_metrics"`
}

type GroupStats struct {
	Period             StatsPeriod              `json:"period"`
	Totals             StatsMetric              `json:"totals"`
	PreviousTotals     StatsMetric              `json:"previous_totals"`
	Timeline           []StatsPoint             `json:"timeline"`
	Ads                []StatsEntityRow         `json:"ads"`
	Placements         []StatsEntityRow         `json:"placements"`
	UnavailableMetrics []StatsUnavailableMetric `json:"unavailable_metrics"`
}

type AdStats struct {
	Period             StatsPeriod              `json:"period"`
	Totals             StatsMetric              `json:"totals"`
	PreviousTotals     StatsMetric              `json:"previous_totals"`
	Timeline           []StatsPoint             `json:"timeline"`
	Placements         []StatsEntityRow         `json:"placements"`
	UnavailableMetrics []StatsUnavailableMetric `json:"unavailable_metrics"`
}

var unavailableConversionMetrics = []StatsUnavailableMetric{
	{Key: "conversions", Label: "Конверсии", Reason: "События конверсий пока не собираются."},
	{Key: "landing", Label: "Переходы на лендинг", Reason: "После клика нет отдельного события просмотра лендинга."},
	{Key: "leads", Label: "Заявки", Reason: "События заявок пока не подключены."},
	{Key: "ab_test", Label: "A/B-тест", Reason: "Варианты креативов не размечаются в событиях."},
	{Key: "audience", Label: "Аудитория", Reason: "В событиях нет возраста, пола и региона посетителя."},
}

func (s *Service) GetCampaignStats(ctx context.Context, advertiserID, campaignID int, from, to time.Time) (*CampaignStats, error) {
	if _, err := s.ownedCampaign(ctx, advertiserID, campaignID); err != nil {
		return nil, err
	}
	if err := s.ensureStatsReady(); err != nil {
		return nil, err
	}

	filter, previousFilter := buildStatsFilters(campaignID, 0, 0, from, to)
	totals, previous, timeline, err := s.readStatsBase(ctx, filter, previousFilter)
	if err != nil {
		return nil, err
	}

	groups, err := s.adGroupRepo.ListByCampaignID(ctx, campaignID)
	if err != nil {
		return nil, err
	}
	groupNames := make(map[int]string, len(groups))
	for _, group := range groups {
		groupNames[group.ID] = group.Name
	}

	groupBreakdown, err := s.statsReader.Breakdown(ctx, filter, "ad_group_id")
	if err != nil {
		return nil, err
	}
	placementBreakdown, err := s.statsReader.Breakdown(ctx, filter, "partner_site_id")
	if err != nil {
		return nil, err
	}

	return &CampaignStats{
		Period:             periodResponse(filter.From, filter.To),
		Totals:             metricFromTotals(totals),
		PreviousTotals:     metricFromTotals(previous),
		Timeline:           pointsFromTimeline(timeline),
		Groups:             rowsFromBreakdown(groupBreakdown, groupNames, "Группа"),
		Placements:         rowsFromBreakdown(placementBreakdown, nil, "Площадка"),
		UnavailableMetrics: unavailableConversionMetrics,
	}, nil
}

func (s *Service) GetGroupStats(ctx context.Context, advertiserID, campaignID, groupID int, from, to time.Time) (*GroupStats, error) {
	group, err := s.ownedGroup(ctx, advertiserID, groupID)
	if err != nil {
		return nil, err
	}
	if group.AdCampaignID != campaignID {
		return nil, ownershipNotFound("ad group")
	}
	if err := s.ensureStatsReady(); err != nil {
		return nil, err
	}

	filter, previousFilter := buildStatsFilters(campaignID, groupID, 0, from, to)
	totals, previous, timeline, err := s.readStatsBase(ctx, filter, previousFilter)
	if err != nil {
		return nil, err
	}

	ads, err := s.adRepo.ListByAdGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	adNames := make(map[int]string, len(ads))
	for _, ad := range ads {
		adNames[ad.ID] = ad.Title
	}

	adBreakdown, err := s.statsReader.Breakdown(ctx, filter, "ad_id")
	if err != nil {
		return nil, err
	}
	placementBreakdown, err := s.statsReader.Breakdown(ctx, filter, "partner_site_id")
	if err != nil {
		return nil, err
	}

	return &GroupStats{
		Period:             periodResponse(filter.From, filter.To),
		Totals:             metricFromTotals(totals),
		PreviousTotals:     metricFromTotals(previous),
		Timeline:           pointsFromTimeline(timeline),
		Ads:                rowsFromBreakdown(adBreakdown, adNames, "Объявление"),
		Placements:         rowsFromBreakdown(placementBreakdown, nil, "Площадка"),
		UnavailableMetrics: unavailableConversionMetrics,
	}, nil
}

func (s *Service) GetAdStats(ctx context.Context, advertiserID, campaignID, groupID, adID int, from, to time.Time) (*AdStats, error) {
	ad, err := s.adRepo.GetByID(ctx, adID)
	if err != nil {
		return nil, err
	}
	if ad.AdGroupID != groupID {
		return nil, ownershipNotFound("ad")
	}
	group, err := s.ownedGroup(ctx, advertiserID, groupID)
	if err != nil {
		return nil, err
	}
	if group.AdCampaignID != campaignID {
		return nil, ownershipNotFound("ad")
	}
	if err := s.ensureStatsReady(); err != nil {
		return nil, err
	}

	filter, previousFilter := buildStatsFilters(campaignID, groupID, adID, from, to)
	totals, previous, timeline, err := s.readStatsBase(ctx, filter, previousFilter)
	if err != nil {
		return nil, err
	}
	placementBreakdown, err := s.statsReader.Breakdown(ctx, filter, "partner_site_id")
	if err != nil {
		return nil, err
	}

	return &AdStats{
		Period:             periodResponse(filter.From, filter.To),
		Totals:             metricFromTotals(totals),
		PreviousTotals:     metricFromTotals(previous),
		Timeline:           pointsFromTimeline(timeline),
		Placements:         rowsFromBreakdown(placementBreakdown, nil, "Площадка"),
		UnavailableMetrics: unavailableConversionMetrics,
	}, nil
}

func (s *Service) ensureStatsReady() error {
	if s.statsReader == nil {
		return fmt.Errorf("%w: stats reader is not configured", errs.NotImplementedError)
	}
	return nil
}

func (s *Service) readStatsBase(ctx context.Context, filter, previousFilter StatsFilter) (StatsTotals, StatsTotals, []StatsTimelinePoint, error) {
	totals, err := s.statsReader.Totals(ctx, filter)
	if err != nil {
		return StatsTotals{}, StatsTotals{}, nil, err
	}
	previous, err := s.statsReader.Totals(ctx, previousFilter)
	if err != nil {
		return StatsTotals{}, StatsTotals{}, nil, err
	}
	timeline, err := s.statsReader.Timeline(ctx, filter)
	if err != nil {
		return StatsTotals{}, StatsTotals{}, nil, err
	}
	return totals, previous, fillTimeline(filter.From, filter.To, timeline), nil
}

func buildStatsFilters(campaignID, groupID, adID int, from, to time.Time) (StatsFilter, StatsFilter) {
	from = dateOnly(from)
	to = dateOnly(to)
	if from.After(to) {
		from, to = to, from
	}

	days := int(to.Sub(from).Hours()/24) + 1
	previousTo := from.AddDate(0, 0, -1)
	previousFrom := previousTo.AddDate(0, 0, -days+1)

	current := StatsFilter{CampaignID: campaignID, AdGroupID: groupID, AdID: adID, From: from, To: to}
	previous := StatsFilter{CampaignID: campaignID, AdGroupID: groupID, AdID: adID, From: previousFrom, To: previousTo}
	return current, previous
}

func periodResponse(from, to time.Time) StatsPeriod {
	return StatsPeriod{From: from.Format("2006-01-02"), To: to.Format("2006-01-02")}
}

func metricFromTotals(t StatsTotals) StatsMetric {
	ctr := 0.0
	if t.Impressions > 0 {
		ctr = float64(t.Clicks) / float64(t.Impressions) * 100
	}
	cpc := 0.0
	if t.Clicks > 0 {
		cpc = float64(t.Spend) / float64(t.Clicks)
	}

	return StatsMetric{
		Impressions:     t.Impressions,
		Clicks:          t.Clicks,
		CTR:             round2(ctr),
		Spend:           t.Spend,
		CPC:             round2(cpc),
		PartnerReward:   t.PartnerReward,
		PlatformRevenue: t.PlatformRevenue,
	}
}

func pointsFromTimeline(points []StatsTimelinePoint) []StatsPoint {
	out := make([]StatsPoint, 0, len(points))
	for _, point := range points {
		out = append(out, StatsPoint{
			Date:        point.Date.Format("2006-01-02"),
			StatsMetric: metricFromTotals(point.Totals),
		})
	}
	return out
}

func rowsFromBreakdown(rows []StatsBreakdownRow, names map[int]string, fallbackPrefix string) []StatsEntityRow {
	out := make([]StatsEntityRow, 0, len(rows))
	for _, row := range rows {
		name := names[row.ID]
		if name == "" {
			name = fmt.Sprintf("%s #%d", fallbackPrefix, row.ID)
		}
		out = append(out, StatsEntityRow{
			ID:          row.ID,
			Name:        name,
			StatsMetric: metricFromTotals(row.Totals),
		})
	}
	return out
}

func fillTimeline(from, to time.Time, points []StatsTimelinePoint) []StatsTimelinePoint {
	byDate := make(map[string]StatsTotals, len(points))
	for _, point := range points {
		byDate[dateOnly(point.Date).Format("2006-01-02")] = point.Totals
	}

	out := make([]StatsTimelinePoint, 0, int(to.Sub(from).Hours()/24)+1)
	for day := from; !day.After(to); day = day.AddDate(0, 0, 1) {
		out = append(out, StatsTimelinePoint{
			Date:   day,
			Totals: byDate[day.Format("2006-01-02")],
		})
	}
	return out
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
