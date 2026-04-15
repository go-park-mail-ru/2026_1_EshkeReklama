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
	insertFeedLink = `INSERT INTO eshkere.ad_feed_link (advertiser_id, token)
		VALUES ($1, $2)
		ON CONFLICT (advertiser_id) DO UPDATE SET token = EXCLUDED.token, updated_at = NOW()`

	selectAdvertiserByFeedToken = `SELECT advertiser_id FROM eshkere.ad_feed_link WHERE token = $1`
)

func (r *FeedLinkRepository) Create(ctx context.Context, advertiserID int, token string) error {
	logger.GetLoggerFromCtx(ctx).Debugf("db: create feed link by advertiserID: %d", advertiserID)

	_, err := r.db.ExecContext(ctx, insertFeedLink, advertiserID, token)
	if err != nil {
		return fmt.Errorf("upsert feed link: %w", err)
	}
	return nil
}

func (r *FeedLinkRepository) GetAdvertiserIDByToken(ctx context.Context, token string) (int, error) {
	logger.GetLoggerFromCtx(ctx).Debugf("db: get feed link by token: %s", token)

	var advertiserID int
	err := r.db.QueryRowContext(ctx, selectAdvertiserByFeedToken, token).Scan(&advertiserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, sql.ErrNoRows
		}
		return 0, fmt.Errorf("get advertiser by feed token: %w", err)
	}
	return advertiserID, nil
}
