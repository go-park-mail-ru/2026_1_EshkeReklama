package service

import (
	"context"
	"fmt"

	errs "eshkere/internal/errors"
	"eshkere/internal/models"
)

func (s *Service) GetAdvertiserAutopaySettings(ctx context.Context, advertiserID int) (*models.AdvertiserAutopaySettings, error) {
	if advertiserID <= 0 {
		return nil, fmt.Errorf("%w: invalid advertiser id", errs.ErrInvalidAdvertiserArg)
	}
	return s.autopaySettingsRepo.GetByAdvertiserID(ctx, advertiserID)
}

func (s *Service) UpdateAdvertiserAutopaySettings(ctx context.Context, settings *models.AdvertiserAutopaySettings) error {
	if settings == nil || settings.AdvertiserID <= 0 {
		return fmt.Errorf("%w: invalid advertiser id", errs.ErrInvalidAdvertiserArg)
	}
	if settings.Enabled {
		if settings.ThresholdAmount < 1000 {
			return fmt.Errorf("%w: threshold must be at least 1000", errs.ErrInvalidAdvertiserArg)
		}
		if settings.TopUpAmount < settings.ThresholdAmount {
			return fmt.Errorf("%w: top up amount must be greater than or equal to threshold", errs.ErrInvalidAdvertiserArg)
		}
	}
	return s.autopaySettingsRepo.Upsert(ctx, settings)
}

func (s *Service) GetAdvertiserNotificationSettings(ctx context.Context, advertiserID int) (*models.AdvertiserNotificationSettings, error) {
	if advertiserID <= 0 {
		return nil, fmt.Errorf("%w: invalid advertiser id", errs.ErrInvalidAdvertiserArg)
	}
	return s.notificationSettingsRepo.GetByAdvertiserID(ctx, advertiserID)
}

func (s *Service) UpdateAdvertiserNotificationSettings(ctx context.Context, settings *models.AdvertiserNotificationSettings) error {
	if settings == nil || settings.AdvertiserID <= 0 {
		return fmt.Errorf("%w: invalid advertiser id", errs.ErrInvalidAdvertiserArg)
	}
	if settings.WarningThreshold <= settings.CriticalThreshold {
		return fmt.Errorf("%w: warning threshold must be greater than critical threshold", errs.ErrInvalidAdvertiserArg)
	}
	return s.notificationSettingsRepo.Upsert(ctx, settings)
}

func (s *Service) GetEnabledNotificationSettings(ctx context.Context) ([]models.AdvertiserNotificationSettings, error) {
	if s.notificationSettingsRepo == nil {
		return nil, nil
	}
	return s.notificationSettingsRepo.ListEnabled(ctx)
}
