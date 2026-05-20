package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"eshkere/internal/models"
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
		id, last_name, first_name, middle_name, birth_date, country_code, registration_region_code, cooperation_form, payout_currency, balance, created_at, updated_at
	FROM eshkere.partner
	WHERE id = $1`

	updatePartner = `UPDATE eshkere.partner SET
		last_name = $1, first_name = $2, middle_name = $3, birth_date = $4, country_code = $5, registration_region_code = $6, cooperation_form = $7, payout_currency = $8
	WHERE id = $9`

	selectUnsettledPartnerEarnings = `SELECT ps.partner_id, COALESCE(SUM(e.reward), 0)
	FROM eshkere.partner_block_daily_earning e
	JOIN eshkere.partner_block pb ON pb.id = e.partner_block_id
	JOIN eshkere.partner_site ps ON ps.id = pb.partner_site_id
	WHERE e.earning_date = $1 AND e.is_settled = FALSE
	GROUP BY ps.partner_id`

	updatePartnerBalanceByEarning = `UPDATE eshkere.partner
	SET balance = balance + $1
	WHERE id = $2`

	markPartnerBlockEarningsSettled = `UPDATE eshkere.partner_block_daily_earning
	SET is_settled = TRUE, settled_at = NOW(), updated_at = NOW()
	WHERE earning_date = $1 AND is_settled = FALSE`
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
		&partner.Balance,
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

func (r *PartnerRepository) SettleDailyEarnings(ctx context.Context, earningDate time.Time) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin settle partner earnings tx: %w", err)
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, selectUnsettledPartnerEarnings, earningDate)
	if err != nil {
		return 0, fmt.Errorf("select unsettled partner earnings: %w", err)
	}

	type partnerEarning struct {
		partnerID int
		reward    int64
	}
	earnings := make([]partnerEarning, 0)
	for rows.Next() {
		var earning partnerEarning
		if err := rows.Scan(&earning.partnerID, &earning.reward); err != nil {
			_ = rows.Close()
			return 0, fmt.Errorf("scan partner earning: %w", err)
		}
		earnings = append(earnings, earning)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return 0, fmt.Errorf("iterate partner earnings: %w", err)
	}
	_ = rows.Close()

	var totalSettled int64
	for _, earning := range earnings {
		if earning.reward <= 0 {
			continue
		}
		if _, err := tx.ExecContext(ctx, updatePartnerBalanceByEarning, earning.reward, earning.partnerID); err != nil {
			return 0, fmt.Errorf("update partner balance by earning: %w", err)
		}
		totalSettled += earning.reward
	}

	if _, err := tx.ExecContext(ctx, markPartnerBlockEarningsSettled, earningDate); err != nil {
		return 0, fmt.Errorf("mark partner block earnings settled: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit settle partner earnings tx: %w", err)
	}

	return totalSettled, nil
}
