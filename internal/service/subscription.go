package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	errs "eshkere/internal/errors"
	"eshkere/internal/models"
)

type TariffInfo struct {
	Tariff        string  `json:"tariff"`
	IsProActive   bool    `json:"is_pro_active"`
	ExpiresAt     *string `json:"expires_at,omitempty"`
	MaxCampaigns  int     `json:"max_campaigns"`
	UsedCampaigns int     `json:"used_campaigns"`
	PriceRub      int     `json:"price_rub"`
}

func checkActiveCampaignLimit(adv *models.Advertiser, activeCount int) error {
	max := adv.MaxCampaigns()
	if activeCount < max {
		return nil
	}
	if adv.Tariff == models.TariffTypePro && !adv.IsProActive() {
		return fmt.Errorf("%w: pro subscription has expired, renew to create more campaigns", errs.ErrProRequired)
	}
	return fmt.Errorf("%w: active campaigns limit is %d for %s plan",
		errs.ErrPlanLimitExceeded, max, adv.Tariff)
}

// RequireProActive возвращает рекламодателя с активной Pro-подпиской или ErrProRequired.
func (s *Service) RequireProActive(ctx context.Context, advertiserID int) (*models.Advertiser, error) {
	adv, err := s.advertiserRepo.GetByID(ctx, advertiserID)
	if err != nil {
		return nil, err
	}
	if !adv.IsProActive() {
		return nil, fmt.Errorf("%w: active pro subscription required", errs.ErrProRequired)
	}
	return adv, nil
}

// RunSubscriptionExpiryCycle сбрасывает истёкшие Pro-подписки на Basic.
func (s *Service) RunSubscriptionExpiryCycle(ctx context.Context) (int, error) {
	ids, err := s.advertiserRepo.ListExpiredProAdvertiserIDs(ctx)
	if err != nil {
		return 0, err
	}

	deactivated := 0
	for _, id := range ids {
		if err := s.DeactivatePro(ctx, id); err != nil {
			return deactivated, fmt.Errorf("deactivate expired pro for advertiser %d: %w", id, err)
		}
		deactivated++
	}

	return deactivated, nil
}

// GetTariffInfo возвращает информацию о текущем тарифе рекламодателя.
func (s *Service) GetTariffInfo(ctx context.Context, advertiserID int) (*TariffInfo, error) {
	adv, err := s.advertiserRepo.GetByID(ctx, advertiserID)
	if err != nil {
		return nil, err
	}

	count, err := s.adCampaignRepo.CountActiveByAdvertiserID(ctx, advertiserID)
	if err != nil {
		return nil, err
	}

	info := &TariffInfo{
		Tariff:        string(adv.Tariff),
		IsProActive:   adv.IsProActive(),
		MaxCampaigns:  adv.MaxCampaigns(),
		UsedCampaigns: count,
		PriceRub:      0,
	}

	if adv.Tariff != models.TariffTypeCheater {
		info.PriceRub = models.SubscriptionPriceRub
	}

	if adv.TariffExpiresAt.Valid {
		formatted := adv.TariffExpiresAt.Time.Format(time.RFC3339)
		info.ExpiresAt = &formatted
	}

	return info, nil
}

// PurchaseProSubscription списывает стоимость подписки с баланса и активирует Pro на 30 дней.
func (s *Service) PurchaseProSubscription(ctx context.Context, advertiserID int) (*TariffInfo, error) {
	if advertiserID <= 0 {
		return nil, fmt.Errorf("%w: invalid advertiser id", errs.ErrInvalidAdvertiserArg)
	}

	adv, err := s.advertiserRepo.GetByID(ctx, advertiserID)
	if err != nil {
		return nil, err
	}

	price := int64(models.SubscriptionPriceRub)
	if adv.Balance < price {
		return nil, fmt.Errorf("%w: need %d rub, balance is %d", errs.ErrInsufficientBalance, price, adv.Balance)
	}

	adv.Balance -= price
	if err = s.advertiserRepo.Update(ctx, adv); err != nil {
		return nil, fmt.Errorf("debit balance for pro subscription: %w", err)
	}

	info, err := s.ActivatePro(ctx, advertiserID)
	if err != nil {
		adv.Balance += price
		_ = s.advertiserRepo.Update(ctx, adv)
		return nil, fmt.Errorf("activate pro after balance debit: %w", err)
	}

	return info, nil
}

// ActivatePro переводит рекламодателя на Pro-тариф на 30 дней.
// Вызывается после успешного подтверждения оплаты.
func (s *Service) ActivatePro(ctx context.Context, advertiserID int) (*TariffInfo, error) {
	adv, err := s.advertiserRepo.GetByID(ctx, advertiserID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var newExpiry time.Time

	// Если Pro уже активен — продлеваем от текущей даты окончания
	if adv.IsProActive() && adv.TariffExpiresAt.Valid {
		newExpiry = adv.TariffExpiresAt.Time.Add(models.SubscriptionDuration)
	} else {
		newExpiry = now.Add(models.SubscriptionDuration)
	}

	adv.Tariff = models.TariffTypePro
	adv.TariffExpiresAt = sql.NullTime{Time: newExpiry, Valid: true}

	if err = s.advertiserRepo.Update(ctx, adv); err != nil {
		return nil, fmt.Errorf("activate pro: %w", err)
	}

	count, err := s.adCampaignRepo.CountActiveByAdvertiserID(ctx, advertiserID)
	if err != nil {
		return nil, err
	}

	formatted := newExpiry.Format(time.RFC3339)
	return &TariffInfo{
		Tariff:        string(models.TariffTypePro),
		IsProActive:   true,
		ExpiresAt:     &formatted,
		MaxCampaigns:  models.MaxCampaignsPro,
		UsedCampaigns: count,
		PriceRub:      models.SubscriptionPriceRub,
	}, nil
}

// DeactivatePro вручную переводит рекламодателя обратно на Basic (для админа или по истечении).
func (s *Service) DeactivatePro(ctx context.Context, advertiserID int) error {
	adv, err := s.advertiserRepo.GetByID(ctx, advertiserID)
	if err != nil {
		return err
	}

	if adv.Tariff == models.TariffTypeCheater {
		return fmt.Errorf("%w: cheater tariff cannot be changed via this method", errs.ForbiddenError)
	}

	adv.Tariff = models.TariffTypeBasic
	adv.TariffExpiresAt = sql.NullTime{Valid: false}

	return s.advertiserRepo.Update(ctx, adv)
}
