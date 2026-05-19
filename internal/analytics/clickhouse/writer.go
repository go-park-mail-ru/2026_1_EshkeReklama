package clickhouse

import (
	"context"
	"eshkere/internal/analytics"
	"fmt"
	"time"

	clickhousego "github.com/ClickHouse/clickhouse-go/v2"
)

type Config struct {
	Addr     string
	Database string
	Username string
	Password string
}

type Writer struct {
	conn clickhousego.Conn
}

func NewWriter(ctx context.Context, cfg Config) (*Writer, error) {
	if cfg.Addr == "" {
		return nil, fmt.Errorf("clickhouse addr cannot be empty")
	}
	if cfg.Database == "" {
		cfg.Database = "default"
	}

	conn, err := clickhousego.Open(&clickhousego.Options{
		Addr: []string{cfg.Addr},
		Auth: clickhousego.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("open clickhouse: %w", err)
	}
	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping clickhouse: %w", err)
	}

	return &Writer{conn: conn}, nil
}

func (w *Writer) InsertAdEvents(ctx context.Context, events []analytics.AdEvent) error {
	if len(events) == 0 {
		return nil
	}

	batch, err := w.conn.PrepareBatch(ctx, `INSERT INTO ad_events (
		event_id,
		event_type,
		request_id,
		occurred_at,
		visitor_id,
		advertiser_id,
		campaign_id,
		ad_group_id,
		ad_id,
		partner_block_id,
		partner_site_id,
		topic_id,
		price,
		partner_reward,
		platform_revenue
	)`)
	if err != nil {
		return fmt.Errorf("prepare ad events batch: %w", err)
	}

	for _, event := range events {
		if err := batch.Append(
			event.EventID,
			event.EventType,
			event.RequestID,
			event.OccurredAt,
			event.VisitorID,
			uint64(event.AdvertiserID),
			uint64(event.CampaignID),
			uint64(event.AdGroupID),
			uint64(event.AdID),
			uint64(event.PartnerBlockID),
			uint64(event.PartnerSiteID),
			uint64(event.TopicID),
			event.Price,
			event.PartnerReward,
			event.PlatformRevenue,
		); err != nil {
			return fmt.Errorf("append ad event: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("send ad events batch: %w", err)
	}
	return nil
}

func (w *Writer) Close() error {
	if w == nil || w.conn == nil {
		return nil
	}
	return w.conn.Close()
}
