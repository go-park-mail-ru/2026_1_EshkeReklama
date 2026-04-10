package postgres

import (
	"context"
	"database/sql"
	"errors"
	"eshkere/internal/models"
	"fmt"
)

type AdGroupRepository struct {
	db *sql.DB
}

func NewAdGroupRepository(db *sql.DB) *AdGroupRepository {
	return &AdGroupRepository{db: db}
}

const (
	insertAdGroup = `INSERT INTO eshkere.ad_group (
		ad_campaign_id, topic_id, region_id, name, age_from, age_to, gender)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	selectAdGroupByID = `SELECT
		id, ad_campaign_id, topic_id, region_id, name, age_from, age_to, gender, created_at, updated_at
	FROM eshkere.ad_group
	WHERE id = $1`

	selectAdGroupsByCampaignID = `SELECT
		id, ad_campaign_id, topic_id, region_id, name, age_from, age_to, gender, created_at, updated_at
	FROM eshkere.ad_group
	WHERE ad_campaign_id = $1
	ORDER BY created_at DESC, id DESC`

	updateAdGroup = `UPDATE eshkere.ad_group SET
		ad_campaign_id = $1, topic_id = $2, region_id = $3, name = $4, age_from = $5, age_to = $6, gender = $7
	WHERE id = $8`

	deleteAdGroup = `DELETE FROM eshkere.ad_group WHERE id = $1`
)

func (r *AdGroupRepository) Create(ctx context.Context, g *models.AdGroup) (int, error) {
	if g == nil {
		return 0, fmt.Errorf("ad group cannot be nil")
	}

	err := r.db.QueryRowContext(ctx, insertAdGroup,
		g.AdCampaignID, g.TopicID, g.RegionID, g.Name, g.AgeFrom, g.AgeTo, g.Gender,
	).Scan(&g.ID)
	if err != nil {
		return 0, fmt.Errorf("insert ad group: %w", err)
	}

	return g.ID, nil
}

func (r *AdGroupRepository) GetByID(ctx context.Context, id int) (*models.AdGroup, error) {
	var g models.AdGroup

	err := r.db.QueryRowContext(ctx, selectAdGroupByID, id).Scan(
		&g.ID,
		&g.AdCampaignID,
		&g.TopicID,
		&g.RegionID,
		&g.Name,
		&g.AgeFrom,
		&g.AgeTo,
		&g.Gender,
		&g.CreatedAt,
		&g.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("ad group not found: %w", err)
		}
		return nil, fmt.Errorf("get ad group by id: %w", err)
	}

	return &g, nil
}

func (r *AdGroupRepository) ListByCampaignID(ctx context.Context, campaignID int) ([]*models.AdGroup, error) {
	rows, err := r.db.QueryContext(ctx, selectAdGroupsByCampaignID, campaignID)
	if err != nil {
		return nil, fmt.Errorf("list ad groups by ad_campaign_id: %w", err)
	}
	defer rows.Close()

	groups := make([]*models.AdGroup, 0)
	for rows.Next() {
		var g models.AdGroup
		if err = rows.Scan(
			&g.ID,
			&g.AdCampaignID,
			&g.TopicID,
			&g.RegionID,
			&g.Name,
			&g.AgeFrom,
			&g.AgeTo,
			&g.Gender,
			&g.CreatedAt,
			&g.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan ad group: %w", err)
		}

		groups = append(groups, &g)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ad groups rows: %w", err)
	}

	return groups, nil
}

func (r *AdGroupRepository) Update(ctx context.Context, g *models.AdGroup) error {
	if g == nil {
		return fmt.Errorf("ad group cannot be nil")
	}

	_, err := r.db.ExecContext(ctx, updateAdGroup,
		g.AdCampaignID, g.TopicID, g.RegionID, g.Name, g.AgeFrom, g.AgeTo, g.Gender, g.ID,
	)
	if err != nil {
		return fmt.Errorf("update ad group: %w", err)
	}

	return nil
}

func (r *AdGroupRepository) Delete(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, deleteAdGroup, id)
	if err != nil {
		return fmt.Errorf("delete ad group: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("ad group not found")
	}

	return nil
}
