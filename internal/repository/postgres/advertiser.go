package postgres

import (
	"context"
	"database/sql"
	"errors"
	errs "eshkere/internal/errors"
	"eshkere/internal/models"
	"eshkere/pkg/logger"
	"fmt"
)

type AdvertiserRepository struct {
	db *sql.DB
}

func NewAdvertiserRepository(db *sql.DB) *AdvertiserRepository {
	return &AdvertiserRepository{db: db}
}

const (
	selectAdvertiserByID = `SELECT
        id, name, surname, avatar_url, balance, company, city, tariff, role, created_at, updated_at
    FROM eshkere.advertiser
    WHERE id = $1`

	updateAdvertiser = `UPDATE eshkere.advertiser SET 
    name = $1, surname = $2, avatar_url = $3, balance = $4, company = $5, city = $6, tariff = $7, role = $8
    WHERE id = $9`

	// insertProfile вставляет профиль рекламодателя с явным id (из auth-сервиса).
	// OVERRIDING SYSTEM VALUE позволяет передать id явно при GENERATED ALWAYS AS IDENTITY.
	insertProfile = `INSERT INTO eshkere.advertiser
		(id, name, balance, tariff, created_at)
		OVERRIDING SYSTEM VALUE
		VALUES ($1, $2, 0, 'noob', NOW())`
)

func (r *AdvertiserRepository) GetByID(ctx context.Context, id int) (*models.Advertiser, error) {
	logger.GetLoggerFromCtx(ctx).Debugf("db: get advertiser by id: %d", id)

	var a models.Advertiser

	err := r.db.QueryRowContext(ctx, selectAdvertiserByID, id).Scan(
		&a.ID,
		&a.Name,
		&a.Surname,
		&a.AvatarURL,
		&a.Balance,
		&a.Company,
		&a.City,
		&a.Tariff,
		&a.Role,
		&a.CreatedAt,
		&a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: advertiser not found", errs.NotFoundError)
		}
		return nil, fmt.Errorf("get advertiser by id: %w", err)
	}

	return &a, nil
}

func (r *AdvertiserRepository) Update(ctx context.Context, a *models.Advertiser) error {
	if a == nil {
		return fmt.Errorf("advertiser cannot be nil")
	}

	logger.GetLoggerFromCtx(ctx).Debugf("db: update advertiser by id: %d", a.ID)

	_, err := r.db.ExecContext(ctx, updateAdvertiser, a.Name, a.Surname, a.AvatarURL, a.Balance, a.Company, a.City, a.Tariff, a.Role, a.ID)
	if err != nil {
		return fmt.Errorf("update advertiser: %w", err)
	}

	return nil
}

func (r *AdvertiserRepository) CreateProfile(ctx context.Context, id int64, name string) error {
	_, err := r.db.ExecContext(ctx, insertProfile, id, name)
	if err != nil {
		return fmt.Errorf("insert advertiser profile: %w", err)
	}
	return nil
}
