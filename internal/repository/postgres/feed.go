package postgres

import (
	"context"
	"database/sql"
	"errors"
	"eshkere/pkg/logger"
	"fmt"
)

type FeedLinkRepository struct {
	db *sql.DB
}

func NewFeedLinkRepository(db *sql.DB) *FeedLinkRepository {
	return &FeedLinkRepository{db: db}
}

const (
	insertFeedLink = `INSERT INTO eshkere.ad_feed_link (ad_campaign_id, token)
		VALUES ($1, $2)
		ON CONFLICT (ad_campaign_id) DO UPDATE SET token = EXCLUDED.token, updated_at = NOW()`

	selectCampaignIDByFeedToken = `SELECT ad_campaign_id FROM eshkere.ad_feed_link WHERE token = $1`
)

func (r *FeedLinkRepository) Create(ctx context.Context, campaignID int, token string) error {
	logger.GetLoggerFromCtx(ctx).Debugf("db: create feed link by campaignID: %d", campaignID)

	_, err := r.db.ExecContext(ctx, insertFeedLink, campaignID, token)
	if err != nil {
		return fmt.Errorf("upsert feed link: %w", err)
	}
	return nil
}

func (r *FeedLinkRepository) GetCampaignIDByToken(ctx context.Context, token string) (int, error) {
	logger.GetLoggerFromCtx(ctx).Debugf("db: get feed link by token: %s", token)

	var campaignID int
	err := r.db.QueryRowContext(ctx, selectCampaignIDByFeedToken, token).Scan(&campaignID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, sql.ErrNoRows
		}
		return 0, fmt.Errorf("get campaign by feed token: %w", err)
	}
	return campaignID, nil
}
