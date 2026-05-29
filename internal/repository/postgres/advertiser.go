package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	errs "eshkere/internal/errors"
	"eshkere/internal/models"
	"eshkere/pkg/logger"
)

type AdvertiserRepository struct {
	db *sql.DB
}

func NewAdvertiserRepository(db *sql.DB) *AdvertiserRepository {
	return &AdvertiserRepository{db: db}
}

const (
	selectAdvertiserByID = `SELECT
        id, name, surname, avatar_url, balance, company, city, tariff, tariff_expires_at, role, saved_payment_method_id, saved_payment_method_title, created_at, updated_at
    FROM eshkere.advertiser
    WHERE id = $1`

	updateAdvertiser = `UPDATE eshkere.advertiser SET
    name = $1, surname = $2, avatar_url = $3, balance = $4, company = $5, city = $6, tariff = $7, tariff_expires_at = $8, role = $9, saved_payment_method_id = $10, saved_payment_method_title = $11
    WHERE id = $12`

	// insertProfile вставляет профиль рекламодателя с явным id (из auth-сервиса).
	// OVERRIDING SYSTEM VALUE позволяет передать id явно при GENERATED ALWAYS AS IDENTITY.
	insertProfile = `INSERT INTO eshkere.advertiser
		(id, name, balance, tariff, created_at)
		OVERRIDING SYSTEM VALUE
		VALUES ($1, $2, 0, 'basic', NOW())`

	listExpiredProAdvertiserIDs = `SELECT id
		FROM eshkere.advertiser
		WHERE tariff = 'pro'
		  AND tariff_expires_at IS NOT NULL
		  AND tariff_expires_at <= NOW()`
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
		&a.TariffExpiresAt,
		&a.Role,
		&a.SavedPaymentMethodID,
		&a.SavedPaymentMethodTitle,
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

	_, err := r.db.ExecContext(ctx, updateAdvertiser,
		a.Name, a.Surname, a.AvatarURL, a.Balance, a.Company, a.City,
		a.Tariff, a.TariffExpiresAt, a.Role,
		a.SavedPaymentMethodID, a.SavedPaymentMethodTitle, a.ID,
	)
	if err != nil {
		return fmt.Errorf("update advertiser: %w", err)
	}

	return nil
}

func (r *AdvertiserRepository) ListExpiredProAdvertiserIDs(ctx context.Context) ([]int, error) {
	logger.GetLoggerFromCtx(ctx).Debug("db: list expired pro advertiser ids")

	rows, err := r.db.QueryContext(ctx, listExpiredProAdvertiserIDs)
	if err != nil {
		return nil, fmt.Errorf("list expired pro advertiser ids: %w", err)
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan expired pro advertiser id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate expired pro advertiser ids: %w", err)
	}

	return ids, nil
}

func (r *AdvertiserRepository) CreateProfile(ctx context.Context, id int64, name string) error {
	_, err := r.db.ExecContext(ctx, insertProfile, id, name)
	if err != nil {
		return fmt.Errorf("insert advertiser profile: %w", err)
	}
	return nil
}
