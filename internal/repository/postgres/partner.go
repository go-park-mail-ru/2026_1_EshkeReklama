package postgres

import (
	"context"
	"database/sql"
	"errors"
	"eshkere/internal/models"
	"fmt"
)

type PartnerRepository struct {
	db *sql.DB
}

func NewPartnerRepository(db *sql.DB) *PartnerRepository {
	return &PartnerRepository{db: db}
}

const (
	insertPartnerProfile = `INSERT INTO eshkere.partner (
		id, last_name, first_name, middle_name, birth_date, country_code, registration_region_code, cooperation_form, payout_currency, created_at
	) OVERRIDING SYSTEM VALUE VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())`

	selectPartnerByID = `SELECT
		id, last_name, first_name, middle_name, birth_date, country_code, registration_region_code, cooperation_form, payout_currency, created_at, updated_at
	FROM eshkere.partner
	WHERE id = $1`

	updatePartner = `UPDATE eshkere.partner SET
		last_name = $1, first_name = $2, middle_name = $3, birth_date = $4, country_code = $5, registration_region_code = $6, cooperation_form = $7, payout_currency = $8
	WHERE id = $9`
)

func (r *PartnerRepository) CreateProfile(ctx context.Context, partner *models.Partner) error {
	if partner == nil {
		return fmt.Errorf("partner cannot be nil")
	}

	_, err := r.db.ExecContext(ctx, insertPartnerProfile,
		partner.ID,
		partner.LastName,
		partner.FirstName,
		partner.MiddleName,
		partner.BirthDate,
		partner.CountryCode,
		partner.RegistrationRegionCode,
		partner.CooperationForm,
		partner.PayoutCurrency,
	)
	if err != nil {
		return fmt.Errorf("insert partner profile: %w", err)
	}
	return nil
}

func (r *PartnerRepository) GetByID(ctx context.Context, id int) (*models.Partner, error) {
	var partner models.Partner
	err := r.db.QueryRowContext(ctx, selectPartnerByID, id).Scan(
		&partner.ID,
		&partner.LastName,
		&partner.FirstName,
		&partner.MiddleName,
		&partner.BirthDate,
		&partner.CountryCode,
		&partner.RegistrationRegionCode,
		&partner.CooperationForm,
		&partner.PayoutCurrency,
		&partner.CreatedAt,
		&partner.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("partner not found: %w", err)
		}
		return nil, fmt.Errorf("get partner by id: %w", err)
	}
	return &partner, nil
}

func (r *PartnerRepository) Update(ctx context.Context, partner *models.Partner) error {
	if partner == nil {
		return fmt.Errorf("partner cannot be nil")
	}

	_, err := r.db.ExecContext(ctx, updatePartner,
		partner.LastName,
		partner.FirstName,
		partner.MiddleName,
		partner.BirthDate,
		partner.CountryCode,
		partner.RegistrationRegionCode,
		partner.CooperationForm,
		partner.PayoutCurrency,
		partner.ID,
	)
	if err != nil {
		return fmt.Errorf("update partner: %w", err)
	}
	return nil
}
