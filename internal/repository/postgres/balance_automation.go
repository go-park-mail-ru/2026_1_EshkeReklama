package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	errs "eshkere/internal/errors"
	"eshkere/internal/models"
)

type AdvertiserAutopaySettingsRepository struct {
	db *sql.DB
}

func NewAdvertiserAutopaySettingsRepository(db *sql.DB) *AdvertiserAutopaySettingsRepository {
	return &AdvertiserAutopaySettingsRepository{db: db}
}

const (
	upsertAdvertiserAutopaySettings = `INSERT INTO eshkere.advertiser_autopay_settings
		(advertiser_id, enabled, threshold_amount, top_up_amount, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (advertiser_id) DO UPDATE SET
			enabled = EXCLUDED.enabled,
			threshold_amount = EXCLUDED.threshold_amount,
			top_up_amount = EXCLUDED.top_up_amount,
			updated_at = NOW()`

	selectAdvertiserAutopaySettings = `SELECT
		advertiser_id, enabled, threshold_amount, top_up_amount, updated_at
	FROM eshkere.advertiser_autopay_settings
	WHERE advertiser_id = $1`

	selectEligibleAutopayCandidates = `SELECT
		aas.advertiser_id,
		adv.balance,
		COALESCE(adv.saved_payment_method_id, ''),
		aas.threshold_amount,
		aas.top_up_amount
	FROM eshkere.advertiser_autopay_settings aas
	JOIN eshkere.advertiser adv ON adv.id = aas.advertiser_id
	WHERE aas.enabled = true
	  AND adv.saved_payment_method_id IS NOT NULL
	  AND adv.balance < aas.threshold_amount`
)

func (r *AdvertiserAutopaySettingsRepository) GetByAdvertiserID(ctx context.Context, advertiserID int) (*models.AdvertiserAutopaySettings, error) {
	var settings models.AdvertiserAutopaySettings
	err := r.db.QueryRowContext(ctx, selectAdvertiserAutopaySettings, advertiserID).Scan(
		&settings.AdvertiserID,
		&settings.Enabled,
		&settings.ThresholdAmount,
		&settings.TopUpAmount,
		&settings.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &models.AdvertiserAutopaySettings{
				AdvertiserID:    advertiserID,
				Enabled:         false,
				ThresholdAmount: 5000,
				TopUpAmount:     30000,
			}, nil
		}
		return nil, fmt.Errorf("get advertiser autopay settings: %w", err)
	}

	return &settings, nil
}

func (r *AdvertiserAutopaySettingsRepository) Upsert(ctx context.Context, settings *models.AdvertiserAutopaySettings) error {
	if settings == nil {
		return fmt.Errorf("autopay settings cannot be nil")
	}

	if _, err := r.db.ExecContext(ctx, upsertAdvertiserAutopaySettings, settings.AdvertiserID, settings.Enabled, settings.ThresholdAmount, settings.TopUpAmount); err != nil {
		return fmt.Errorf("upsert advertiser autopay settings: %w", err)
	}

	return nil
}

func (r *AdvertiserAutopaySettingsRepository) ListEligible(ctx context.Context) ([]models.AdvertiserAutopayCandidate, error) {
	rows, err := r.db.QueryContext(ctx, selectEligibleAutopayCandidates)
	if err != nil {
		return nil, fmt.Errorf("list eligible autopay candidates: %w", err)
	}
	defer rows.Close()

	var out []models.AdvertiserAutopayCandidate
	for rows.Next() {
		var candidate models.AdvertiserAutopayCandidate
		if err := rows.Scan(
			&candidate.AdvertiserID,
			&candidate.Balance,
			&candidate.SavedPaymentMethodID,
			&candidate.ThresholdAmount,
			&candidate.TopUpAmount,
		); err != nil {
			return nil, fmt.Errorf("scan autopay candidate: %w", err)
		}
		out = append(out, candidate)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate autopay candidates: %w", err)
	}

	return out, nil
}

type AdvertiserNotificationSettingsRepository struct {
	db *sql.DB
}

func NewAdvertiserNotificationSettingsRepository(db *sql.DB) *AdvertiserNotificationSettingsRepository {
	return &AdvertiserNotificationSettingsRepository{db: db}
}

const (
	upsertAdvertiserNotificationSettings = `INSERT INTO eshkere.advertiser_notification_settings
		(advertiser_id, email_enabled, warning_threshold, critical_threshold, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (advertiser_id) DO UPDATE SET
			email_enabled = EXCLUDED.email_enabled,
			warning_threshold = EXCLUDED.warning_threshold,
			critical_threshold = EXCLUDED.critical_threshold,
			updated_at = NOW()`

	selectAdvertiserNotificationSettings = `SELECT
		advertiser_id, email_enabled, warning_threshold, critical_threshold, updated_at
	FROM eshkere.advertiser_notification_settings
	WHERE advertiser_id = $1`

	selectEmailNotificationCandidates = `SELECT
		ans.advertiser_id,
		ans.email_enabled,
		ans.warning_threshold,
		ans.critical_threshold,
		ans.updated_at
	FROM eshkere.advertiser_notification_settings ans
	WHERE ans.email_enabled = true`
)

func (r *AdvertiserNotificationSettingsRepository) GetByAdvertiserID(ctx context.Context, advertiserID int) (*models.AdvertiserNotificationSettings, error) {
	var settings models.AdvertiserNotificationSettings
	err := r.db.QueryRowContext(ctx, selectAdvertiserNotificationSettings, advertiserID).Scan(
		&settings.AdvertiserID,
		&settings.EmailEnabled,
		&settings.WarningThreshold,
		&settings.CriticalThreshold,
		&settings.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &models.AdvertiserNotificationSettings{
				AdvertiserID:      advertiserID,
				EmailEnabled:      false,
				WarningThreshold:  500,
				CriticalThreshold: 100,
			}, nil
		}
		return nil, fmt.Errorf("get advertiser notification settings: %w", err)
	}

	return &settings, nil
}

func (r *AdvertiserNotificationSettingsRepository) Upsert(ctx context.Context, settings *models.AdvertiserNotificationSettings) error {
	if settings == nil {
		return fmt.Errorf("notification settings cannot be nil")
	}

	if settings.WarningThreshold <= settings.CriticalThreshold {
		return fmt.Errorf("%w: warning threshold must be greater than critical threshold", errs.ErrInvalidAdvertiserArg)
	}

	if _, err := r.db.ExecContext(ctx, upsertAdvertiserNotificationSettings, settings.AdvertiserID, settings.EmailEnabled, settings.WarningThreshold, settings.CriticalThreshold); err != nil {
		return fmt.Errorf("upsert advertiser notification settings: %w", err)
	}

	return nil
}

func (r *AdvertiserNotificationSettingsRepository) ListEnabled(ctx context.Context) ([]models.AdvertiserNotificationSettings, error) {
	rows, err := r.db.QueryContext(ctx, selectEmailNotificationCandidates)
	if err != nil {
		return nil, fmt.Errorf("list notification settings: %w", err)
	}
	defer rows.Close()

	var out []models.AdvertiserNotificationSettings
	for rows.Next() {
		var settings models.AdvertiserNotificationSettings
		if err := rows.Scan(
			&settings.AdvertiserID,
			&settings.EmailEnabled,
			&settings.WarningThreshold,
			&settings.CriticalThreshold,
			&settings.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan notification settings: %w", err)
		}
		out = append(out, settings)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notification settings: %w", err)
	}

	return out, nil
}
