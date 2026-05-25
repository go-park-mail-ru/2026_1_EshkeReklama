package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	errs "eshkere/internal/errors"
	"eshkere/internal/models"
)

func (s *Service) TopUpAdvertiserBalance(ctx context.Context, advertiserID int, amount int64) (int64, error) {
	if advertiserID <= 0 {
		return 0, fmt.Errorf("%w: invalid advertiser id", errs.ErrInvalidAdvertiserArg)
	}
	if amount <= 0 {
		return 0, fmt.Errorf("%w: amount must be positive", errs.ErrInvalidAdvertiserArg)
	}

	adv, err := s.advertiserRepo.GetByID(ctx, advertiserID)
	if err != nil {
		return 0, err
	}

	adv.Balance += amount
	if err = s.advertiserRepo.Update(ctx, adv); err != nil {
		return 0, err
	}

	if err = s.reactivateAdsWaitingForBalance(ctx, advertiserID); err != nil {
		return 0, err
	}

	return adv.Balance, nil
}

func (s *Service) reactivateAdsWaitingForBalance(ctx context.Context, advertiserID int) error {
	if s.adCampaignRepo == nil || s.adRepo == nil {
		return nil
	}

	campaigns, err := s.adCampaignRepo.ListByAdvertiserID(ctx, advertiserID)
	if err != nil {
		return err
	}

	for _, campaign := range campaigns {
		if campaign == nil || !s.canStartAd(ctx, campaign) {
			continue
		}

		ads, err := s.adRepo.ListByAdCampaignID(ctx, campaign.ID)
		if err != nil {
			return err
		}

		hasReactivatedAds := false
		for _, ad := range ads {
			if ad == nil || ad.Status != models.AdStatusNotEnoughMoney {
				continue
			}

			ad.Status = models.AdStatusWorking
			ad.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}
			if err = s.adRepo.Update(ctx, ad); err != nil {
				return err
			}
			hasReactivatedAds = true
		}

		if hasReactivatedAds {
			campaign.Status = campaignStatusFromAds(ads)
			campaign.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}
			if err = s.adCampaignRepo.Update(ctx, campaign); err != nil {
				return err
			}
		}
	}

	return nil
}
