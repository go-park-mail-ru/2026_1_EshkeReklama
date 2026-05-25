package clickhouse

import (
	"context"
	"fmt"
	"strings"
	"time"

	"eshkere/internal/service"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type Reader struct {
	conn clickhouseConn
}

type clickhouseConn interface {
	Query(ctx context.Context, query string, args ...any) (driver.Rows, error)
	Close() error
}

func NewReader(ctx context.Context, cfg Config) (*Reader, error) {
	writer, err := NewWriter(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Reader{conn: writer.conn}, nil
}

func (r *Reader) Totals(ctx context.Context, filter service.StatsFilter) (service.StatsTotals, error) {
	query := fmt.Sprintf(`SELECT
		toInt64(countIf(event_type = 'impression')) AS impressions,
		toInt64(countIf(event_type = 'click')) AS clicks,
		coalesce(sumIf(price, event_type = 'impression'), 0) AS spend,
		coalesce(sumIf(partner_reward, event_type = 'impression'), 0) AS partner_reward,
		coalesce(sumIf(platform_revenue, event_type = 'impression'), 0) AS platform_revenue
	FROM ad_events
	WHERE %s`, filterWhere(filter))

	queryRows, err := r.conn.Query(ctx, query, filterArgs(filter)...)
	if err != nil {
		return service.StatsTotals{}, fmt.Errorf("query stats totals: %w", err)
	}
	defer queryRows.Close()

	if !queryRows.Next() {
		return service.StatsTotals{}, queryRows.Err()
	}

	var totals service.StatsTotals
	if err := queryRows.Scan(
		&totals.Impressions,
		&totals.Clicks,
		&totals.Spend,
		&totals.PartnerReward,
		&totals.PlatformRevenue,
	); err != nil {
		return service.StatsTotals{}, fmt.Errorf("scan stats totals: %w", err)
	}

	return totals, queryRows.Err()
}

func (r *Reader) Timeline(ctx context.Context, filter service.StatsFilter) ([]service.StatsTimelinePoint, error) {
	query := fmt.Sprintf(`SELECT
		event_date,
		toInt64(countIf(event_type = 'impression')) AS impressions,
		toInt64(countIf(event_type = 'click')) AS clicks,
		coalesce(sumIf(price, event_type = 'impression'), 0) AS spend,
		coalesce(sumIf(partner_reward, event_type = 'impression'), 0) AS partner_reward,
		coalesce(sumIf(platform_revenue, event_type = 'impression'), 0) AS platform_revenue
	FROM ad_events
	WHERE %s
	GROUP BY event_date
	ORDER BY event_date ASC`, filterWhere(filter))

	queryRows, err := r.conn.Query(ctx, query, filterArgs(filter)...)
	if err != nil {
		return nil, fmt.Errorf("query stats timeline: %w", err)
	}
	defer queryRows.Close()

	points := make([]service.StatsTimelinePoint, 0)
	for queryRows.Next() {
		var point service.StatsTimelinePoint
		if err := queryRows.Scan(
			&point.Date,
			&point.Totals.Impressions,
			&point.Totals.Clicks,
			&point.Totals.Spend,
			&point.Totals.PartnerReward,
			&point.Totals.PlatformRevenue,
		); err != nil {
			return nil, fmt.Errorf("scan stats timeline: %w", err)
		}
		points = append(points, point)
	}

	return points, queryRows.Err()
}

func (r *Reader) Breakdown(ctx context.Context, filter service.StatsFilter, dimension string) ([]service.StatsBreakdownRow, error) {
	if !isAllowedDimension(dimension) {
		return nil, fmt.Errorf("unsupported stats dimension: %s", dimension)
	}

	query := fmt.Sprintf(`SELECT
		toInt64(%s) AS entity_id,
		toInt64(countIf(event_type = 'impression')) AS impressions,
		toInt64(countIf(event_type = 'click')) AS clicks,
		coalesce(sumIf(price, event_type = 'impression'), 0) AS spend,
		coalesce(sumIf(partner_reward, event_type = 'impression'), 0) AS partner_reward,
		coalesce(sumIf(platform_revenue, event_type = 'impression'), 0) AS platform_revenue
	FROM ad_events
	WHERE %s
	GROUP BY entity_id
	HAVING impressions > 0 OR clicks > 0
	ORDER BY impressions DESC, clicks DESC`, dimension, filterWhere(filter))

	queryRows, err := r.conn.Query(ctx, query, filterArgs(filter)...)
	if err != nil {
		return nil, fmt.Errorf("query stats breakdown: %w", err)
	}
	defer queryRows.Close()

	rows := make([]service.StatsBreakdownRow, 0)
	for queryRows.Next() {
		var row service.StatsBreakdownRow
		var entityID int64
		if err := queryRows.Scan(
			&entityID,
			&row.Totals.Impressions,
			&row.Totals.Clicks,
			&row.Totals.Spend,
			&row.Totals.PartnerReward,
			&row.Totals.PlatformRevenue,
		); err != nil {
			return nil, fmt.Errorf("scan stats breakdown: %w", err)
		}
		row.ID = int(entityID)
		rows = append(rows, row)
	}

	return rows, queryRows.Err()
}

func (r *Reader) Close() error {
	if r == nil || r.conn == nil {
		return nil
	}
	return r.conn.Close()
}

func filterWhere(filter service.StatsFilter) string {
	parts := []string{
		"campaign_id = ?",
		"event_date >= ?",
		"event_date <= ?",
	}
	if filter.AdGroupID > 0 {
		parts = append(parts, "ad_group_id = ?")
	}
	if filter.AdID > 0 {
		parts = append(parts, "ad_id = ?")
	}
	return strings.Join(parts, " AND ")
}

func filterArgs(filter service.StatsFilter) []any {
	args := []any{
		uint64(filter.CampaignID),
		dateOnly(filter.From),
		dateOnly(filter.To),
	}
	if filter.AdGroupID > 0 {
		args = append(args, uint64(filter.AdGroupID))
	}
	if filter.AdID > 0 {
		args = append(args, uint64(filter.AdID))
	}
	return args
}

func isAllowedDimension(dimension string) bool {
	switch dimension {
	case "ad_group_id", "ad_id", "partner_site_id", "partner_block_id", "topic_id":
		return true
	default:
		return false
	}
}

func dateOnly(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
