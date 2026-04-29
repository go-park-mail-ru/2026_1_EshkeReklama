package postgres

import (
	"context"
	"database/sql"
	"errors"
	"eshkere/internal/models"
	"eshkere/pkg/logger"
	"fmt"
)

type AdCampaignRepository struct {
	db *sql.DB
}

func NewAdCampaignRepository(db *sql.DB) *AdCampaignRepository {
	return &AdCampaignRepository{db: db}
}

const (
	insertAdCampaign = `INSERT INTO eshkere.ad_campaign (
		advertiser_id, status, name, daily_budget, main_action)
		VALUES ($1, $2, $3, $4, $5) RETURNING id`

	selectAdCampaignByID = `SELECT
		id, advertiser_id, status, name, daily_budget, main_action, created_at, updated_at
	FROM eshkere.ad_campaign
	WHERE id = $1`

	selectAdCampaignsByAdvertiserID = `SELECT
		id, advertiser_id, status, name, daily_budget, main_action, created_at, updated_at
	FROM eshkere.ad_campaign
	WHERE advertiser_id = $1
	ORDER BY created_at DESC, id DESC`

	updateAdCampaign = `UPDATE eshkere.ad_campaign SET
		advertiser_id = $1, status = $2, name = $3, daily_budget = $4, main_action = $5, updated_at = $6
	WHERE id = $7`

	deleteAdCampaign = `DELETE FROM eshkere.ad_campaign WHERE id = $1`
)

func (r *AdCampaignRepository) Create(ctx context.Context, c *models.AdCampaign) error {
	if c == nil {
		return fmt.Errorf("ad campaign cannot be nil")
	}

	logger.GetLoggerFromCtx(ctx).Debug("db: create ad_campaign")

	err := r.db.QueryRowContext(ctx, insertAdCampaign,
		c.AdvertiserID, c.Status, c.Name, c.DailyBudget, c.MainAction,
	).Scan(&c.ID)
	if err != nil {
		return fmt.Errorf("insert ad campaign: %w", err)
	}

	return nil
}

func (r *AdCampaignRepository) GetByID(ctx context.Context, id int) (*models.AdCampaign, error) {
	logger.GetLoggerFromCtx(ctx).Debugf("db: get ad_campaign by id: %d", id)

	var c models.AdCampaign

	err := r.db.QueryRowContext(ctx, selectAdCampaignByID, id).Scan(
		&c.ID,
		&c.AdvertiserID,
		&c.Status,
		&c.Name,
		&c.DailyBudget,
		&c.MainAction,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("ad campaign not found: %w", err)
		}
		return nil, fmt.Errorf("get ad campaign by id: %w", err)
	}

	return &c, nil
}

func (r *AdCampaignRepository) ListByAdvertiserID(ctx context.Context, advertiserID int) ([]*models.AdCampaign, error) {
	logger.GetLoggerFromCtx(ctx).Debugf("db: get ad_campaigns by advertiserID: %d", advertiserID)

	rows, err := r.db.QueryContext(ctx, selectAdCampaignsByAdvertiserID, advertiserID)
	if err != nil {
		return nil, fmt.Errorf("list ad campaigns by advertiser_id: %w", err)
	}
	defer rows.Close()

	campaigns := make([]*models.AdCampaign, 0)
	for rows.Next() {
		var c models.AdCampaign
		if err = rows.Scan(
			&c.ID,
			&c.AdvertiserID,
			&c.Status,
			&c.Name,
			&c.DailyBudget,
			&c.MainAction,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan ad campaign: %w", err)
		}

		campaigns = append(campaigns, &c)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ad campaign rows: %w", err)
	}

	return campaigns, nil
}

func (r *AdCampaignRepository) Update(ctx context.Context, c *models.AdCampaign) error {
	if c == nil {
		return fmt.Errorf("ad campaign cannot be nil")
	}

	logger.GetLoggerFromCtx(ctx).Debugf("db: update ad_campaign by id: %d", c.ID)

	_, err := r.db.ExecContext(ctx, updateAdCampaign,
		c.AdvertiserID, c.Status, c.Name, c.DailyBudget, c.MainAction, c.UpdatedAt, c.ID,
	)
	if err != nil {
		return fmt.Errorf("update ad campaign: %w", err)
	}

	return nil
}

func (r *AdCampaignRepository) Delete(ctx context.Context, id int) error {
	logger.GetLoggerFromCtx(ctx).Debugf("db: delete ad_campaign by id: %d", id)

	result, err := r.db.ExecContext(ctx, deleteAdCampaign, id)
	if err != nil {
		return fmt.Errorf("delete ad campaign: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("ad campaign not found")
	}

	return nil
}
