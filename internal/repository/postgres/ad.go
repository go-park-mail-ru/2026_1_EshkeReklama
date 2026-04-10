package postgres

import (
	"context"
	"database/sql"
	"errors"
	"eshkere/internal/models"
	"fmt"
)

type AdRepository struct {
	db *sql.DB
}

func NewAdRepository(db *sql.DB) *AdRepository {
	return &AdRepository{db: db}
}

const (
	insertAd = `INSERT INTO eshkere.ad (
		ad_group_id, status, title, short_desc, image_url, target_url)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	selectAdByID = `SELECT
		id, ad_group_id, status, title, short_desc, image_url, target_url, created_at, updated_at
	FROM eshkere.ad
	WHERE id = $1`

	selectAdsByAdGroupID = `SELECT
		id, ad_group_id, status, title, short_desc, image_url, target_url, created_at, updated_at
	FROM eshkere.ad
	WHERE ad_group_id = $1
	ORDER BY created_at DESC, id DESC`

	updateAd = `UPDATE eshkere.ad SET
		ad_group_id = $1, status = $2, title = $3, short_desc = $4, image_url = $5, target_url = $6
	WHERE id = $7`

	deleteAd = `DELETE FROM eshkere.ad WHERE id = $1`
)

func (r *AdRepository) Create(ctx context.Context, ad *models.Ad) (int, error) {
	if ad == nil {
		return 0, fmt.Errorf("ad cannot be nil")
	}

	err := r.db.QueryRowContext(ctx, insertAd,
		ad.AdGroupID, ad.Status, ad.Title, ad.ShortDesc, ad.ImageURL, ad.TargetURL,
	).Scan(&ad.ID)
	if err != nil {
		return 0, fmt.Errorf("insert ad: %w", err)
	}

	return ad.ID, nil
}

func (r *AdRepository) GetByID(ctx context.Context, id int) (*models.Ad, error) {
	var ad models.Ad

	err := r.db.QueryRowContext(ctx, selectAdByID, id).Scan(
		&ad.ID,
		&ad.AdGroupID,
		&ad.Status,
		&ad.Title,
		&ad.ShortDesc,
		&ad.ImageURL,
		&ad.TargetURL,
		&ad.CreatedAt,
		&ad.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("ad not found: %w", err)
		}
		return nil, fmt.Errorf("get ad by id: %w", err)
	}

	return &ad, nil
}

func (r *AdRepository) ListByAdGroupID(ctx context.Context, adGroupID int) ([]*models.Ad, error) {
	rows, err := r.db.QueryContext(ctx, selectAdsByAdGroupID, adGroupID)
	if err != nil {
		return nil, fmt.Errorf("list ads by ad_group_id: %w", err)
	}
	defer rows.Close()

	ads := make([]*models.Ad, 0)
	for rows.Next() {
		var ad models.Ad
		if err = rows.Scan(
			&ad.ID,
			&ad.AdGroupID,
			&ad.Status,
			&ad.Title,
			&ad.ShortDesc,
			&ad.ImageURL,
			&ad.TargetURL,
			&ad.CreatedAt,
			&ad.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan ad: %w", err)
		}

		ads = append(ads, &ad)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ads rows: %w", err)
	}

	return ads, nil
}

func (r *AdRepository) Update(ctx context.Context, ad *models.Ad) error {
	if ad == nil {
		return fmt.Errorf("ad cannot be nil")
	}

	_, err := r.db.ExecContext(ctx, updateAd,
		ad.AdGroupID, ad.Status, ad.Title, ad.ShortDesc, ad.ImageURL, ad.TargetURL, ad.ID,
	)
	if err != nil {
		return fmt.Errorf("update ad: %w", err)
	}

	return nil
}

func (r *AdRepository) Delete(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, deleteAd, id)
	if err != nil {
		return fmt.Errorf("delete ad: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("ad not found")
	}

	return nil
}
