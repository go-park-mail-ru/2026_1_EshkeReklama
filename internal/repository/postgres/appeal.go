package postgres

import (
	"context"
	"database/sql"
	"errors"
	"eshkere/internal/models"
	"eshkere/pkg/logger"
	"fmt"
)

type AppealRepository struct {
	db *sql.DB
}

func NewAppealRepository(db *sql.DB) *AppealRepository {
	return &AppealRepository{db: db}
}

const (
	insertAppeal = `INSERT INTO eshkere.appeal (
		advertiser_id, status, category, title, description, image_url, name, email)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	selectAppealByID = `SELECT
		id, advertiser_id, status, category, title, description, image_url, name, email, created_at, updated_at
	FROM eshkere.appeal
	WHERE id = $1`

	selectAppealsByAdvertiserID = `SELECT
		id, advertiser_id, status, category, title, description, image_url, name, email, created_at, updated_at
	FROM eshkere.appeal
	WHERE advertiser_id = $1
	ORDER BY created_at DESC, id DESC`
)

func (r *AppealRepository) Create(ctx context.Context, appeal *models.Appeal) error {
	if appeal == nil {
		return fmt.Errorf("appeal cannot be nil")
	}

	logger.GetLoggerFromCtx(ctx).Debug("db: create appeal")

	err := r.db.QueryRowContext(ctx, insertAppeal,
		appeal.AdvertiserID, appeal.Status, appeal.Category, appeal.Title, appeal.Description, appeal.ImageURL, appeal.Name, appeal.Email,
	).Scan(&appeal.ID)
	if err != nil {
		return fmt.Errorf("insert appeal: %w", err)
	}

	return nil
}

func (r *AppealRepository) GetByID(ctx context.Context, appealID int) (*models.Appeal, error) {
	logger.GetLoggerFromCtx(ctx).Debugf("db: get appeal by id: %d", appealID)
	var appeal models.Appeal

	err := r.db.QueryRowContext(ctx, selectAppealByID, appealID).Scan(
		&appeal.ID,
		&appeal.AdvertiserID,
		&appeal.Status,
		&appeal.Category,
		&appeal.Title,
		&appeal.Description,
		&appeal.ImageURL,
		&appeal.Name,
		&appeal.Email,
		&appeal.CreatedAt,
		&appeal.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("appeal not found: %w", err)
		}
		return nil, fmt.Errorf("get appeal by id: %w", err)
	}

	return &appeal, nil
}

func (r *AppealRepository) ListByAdvertiserID(ctx context.Context, advertiserID int) ([]*models.Appeal, error) {
	logger.GetLoggerFromCtx(ctx).Debugf("db: get appeals by advertiserID: %d", advertiserID)

	rows, err := r.db.QueryContext(ctx, selectAppealsByAdvertiserID, advertiserID)
	if err != nil {
		return nil, fmt.Errorf("list appeals by advertiserID: %w", err)
	}
	defer rows.Close()

	appeals := make([]*models.Appeal, 0)
	for rows.Next() {
		var appeal models.Appeal
		if err = rows.Scan(
			&appeal.ID,
			&appeal.AdvertiserID,
			&appeal.Status,
			&appeal.Category,
			&appeal.Title,
			&appeal.Description,
			&appeal.ImageURL,
			&appeal.Name,
			&appeal.Email,
			&appeal.CreatedAt,
			&appeal.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan appeal: %w", err)
		}

		appeals = append(appeals, &appeal)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate appeals rows: %w", err)
	}

	return appeals, nil
}
