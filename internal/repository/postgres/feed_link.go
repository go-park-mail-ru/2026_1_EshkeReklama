package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type FeedLinkRepository struct {
	db *sql.DB
}

func NewFeedLinkRepository(db *sql.DB) *FeedLinkRepository {
	return &FeedLinkRepository{db: db}
}

const (
	upsertFeedLink = `INSERT INTO eshkere.ad_feed_link (advertiser_id, token)
		VALUES ($1, $2)
		ON CONFLICT (advertiser_id) DO UPDATE SET token = EXCLUDED.token, updated_at = NOW()`

	selectAdvertiserByFeedToken = `SELECT advertiser_id FROM eshkere.ad_feed_link WHERE token = $1`
)

func (r *FeedLinkRepository) UpsertByAdvertiserID(ctx context.Context, advertiserID int, token string) error {
	startedAt := time.Now()
	_, err := r.db.ExecContext(ctx, upsertFeedLink, advertiserID, token)
	logDBQuery(ctx, "feed_link.upsert", startedAt, err)
	if err != nil {
		return fmt.Errorf("upsert feed link: %w", err)
	}
	return nil
}

func (r *FeedLinkRepository) GetAdvertiserIDByToken(ctx context.Context, token string) (int, error) {
	var advertiserID int
	startedAt := time.Now()
	err := r.db.QueryRowContext(ctx, selectAdvertiserByFeedToken, token).Scan(&advertiserID)
	logDBQuery(ctx, "feed_link.get_by_token", startedAt, err)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, sql.ErrNoRows
		}
		return 0, fmt.Errorf("get advertiser by feed token: %w", err)
	}
	return advertiserID, nil
}
