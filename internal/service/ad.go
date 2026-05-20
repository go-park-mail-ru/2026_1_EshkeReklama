package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	errs "eshkere/internal/errors"
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
)

func (s *Service) CreateAd(ctx context.Context, advertiserID int, in *serviceinput.CreateAd) (*models.Ad, error) {
	if _, err := s.ownedGroup(ctx, advertiserID, in.AdGroupID); err != nil {
		return nil, err
	}
	ad := &models.Ad{
		AdGroupID: in.AdGroupID,
		Status:    models.AdStatusModeration,
		Title:     in.Title,
		ShortDesc: in.ShortDesc,
		TargetURL: in.TargetURL,
	}

	if len(in.Image) > 0 {
		if s.adStorage == nil {
			return nil, fmt.Errorf("%w: ad storage is not configured", errs.InternalServiceError)
		}

		imageKey, err := s.adStorage.UploadAdImage(ctx, in.Image, in.ImageExt, in.ImageType)
		if err != nil {
			return nil, err
		}

		ad.ImageURL = imageKey
	}

	if err := s.adRepo.Create(ctx, ad); err != nil {
		if ad.ImageURL != "" && s.adStorage != nil {
			_ = s.adStorage.DeleteAdImage(ctx, ad.ImageURL)
		}
		return nil, err
	}

	s.decorateAdImageURL(ad)

	return ad, nil
}

func (s *Service) GetAdByID(ctx context.Context, adID int) (*models.Ad, error) {
	if adID <= 0 {
		return nil, fmt.Errorf("%w: invalid ad id", errs.BadRequestError)
	}

	ad, err := s.adRepo.GetByID(ctx, adID)
	if err != nil {
		return nil, err
	}

	s.decorateAdImageURL(ad)
	return ad, nil
}

func (s *Service) UpdateAd(ctx context.Context, advertiserID int, in *serviceinput.UpdateAd) error {
	currentAd, err := s.adRepo.GetByID(ctx, in.ID)
	if err != nil {
		return err
	}
	group, err := s.ownedGroup(ctx, advertiserID, currentAd.AdGroupID)
	if err != nil {
		return err
	}
	campaign, err := s.adCampaignRepo.GetByID(ctx, group.AdCampaignID)
	if err != nil {
		return err
	}

	contentChanged := false
	if in.Title != nil {
		currentAd.Title = *in.Title
		contentChanged = true
	}
	if in.ShortDesc != nil {
		currentAd.ShortDesc = *in.ShortDesc
		contentChanged = true
	}
	if in.TargetURL != nil {
		currentAd.TargetURL = *in.TargetURL
		contentChanged = true
	}
	if in.Image != nil && len(*in.Image) > 0 {
		if s.adStorage == nil {
			return fmt.Errorf("%w: ad storage is not configured", errs.InternalServiceError)
		}

		imageExt := ""
		if in.ImageExt != nil {
			imageExt = *in.ImageExt
		}
		imageType := ""
		if in.ImageType != nil {
			imageType = *in.ImageType
		}

		imageKey, err := s.adStorage.UploadAdImage(ctx, *in.Image, imageExt, imageType)
		if err != nil {
			return err
		}
		currentAd.ImageURL = imageKey
		contentChanged = true
	}

	if contentChanged {
		currentAd.Status = models.AdStatusModeration
	} else if in.Status != nil {
		status, err := s.userAdStatusTransition(ctx, currentAd.Status, *in.Status, campaign)
		if err != nil {
			return err
		}
		currentAd.Status = status
	}
	currentAd.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}

	if err := s.adRepo.Update(ctx, currentAd); err != nil {
		if in.Image != nil && len(*in.Image) > 0 && s.adStorage != nil {
			_ = s.adStorage.DeleteAdImage(ctx, currentAd.ImageURL)
		}
		return err
	}

	return s.recalculateCampaignStatus(ctx, campaign.ID)
}

func (s *Service) ListAds(ctx context.Context, advertiserID, groupID int) ([]*models.Ad, error) {
	if _, err := s.ownedGroup(ctx, advertiserID, groupID); err != nil {
		return nil, err
	}
	ads, err := s.adRepo.ListByAdGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	for _, ad := range ads {
		s.decorateAdImageURL(ad)
	}

	return ads, nil
}

func (s *Service) ListModerationAds(ctx context.Context) ([]*models.Ad, error) {
	ads, err := s.adRepo.ListByStatus(ctx, models.AdStatusModeration)
	if err != nil {
		return nil, err
	}

	for _, ad := range ads {
		s.decorateAdImageURL(ad)
	}

	return ads, nil
}

func (s *Service) DeleteAd(ctx context.Context, advertiserID, adID int) error {
	ad, err := s.adRepo.GetByID(ctx, adID)
	if err != nil {
		return err
	}
	group, err := s.ownedGroup(ctx, advertiserID, ad.AdGroupID)
	if err != nil {
		return err
	}
	if err := s.adRepo.Delete(ctx, adID); err != nil {
		return err
	}
	return s.recalculateCampaignStatus(ctx, group.AdCampaignID)
}

func (s *Service) UpdateAdModerationStatus(ctx context.Context, in *serviceinput.UpdateAdStatus) error {
	if in == nil || in.AdID <= 0 {
		return fmt.Errorf("%w: invalid ad status update", errs.BadRequestError)
	}

	ad, err := s.adRepo.GetByID(ctx, in.AdID)
	if err != nil {
		return err
	}

	group, err := s.adGroupRepo.GetByID(ctx, ad.AdGroupID)
	if err != nil {
		return err
	}

	campaign, err := s.adCampaignRepo.GetByID(ctx, group.AdCampaignID)
	if err != nil {
		return err
	}

	switch in.Decision {
	case serviceinput.AdModerationApprove:
		if s.canStartAd(ctx, campaign) {
			ad.Status = models.AdStatusWorking
		} else {
			ad.Status = models.AdStatusNotEnoughMoney
		}
	case serviceinput.AdModerationDisapprove:
		ad.Status = models.AdStatusRejected
	default:
		return fmt.Errorf("%w: invalid moderation decision", errs.BadRequestError)
	}

	ad.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}
	if err := s.adRepo.Update(ctx, ad); err != nil {
		return err
	}

	return s.recalculateCampaignStatus(ctx, campaign.ID)
}

func (s *Service) canStartAd(ctx context.Context, campaign *models.AdCampaign) bool {
	if campaign == nil || s.advertiserRepo == nil {
		return false
	}

	price := ImpressionPrice(campaign.CPMPrice)
	if campaign.DailyBudget <= 0 || campaign.CPMPrice <= 0 || price <= 0 || campaign.DailyBudget < price {
		return false
	}

	advertiser, err := s.advertiserRepo.GetByID(ctx, campaign.AdvertiserID)
	if err != nil {
		return false
	}

	return advertiser.Balance >= price
}

func campaignStatusFromAds(ads []*models.Ad) models.AdStatus {
	if hasAdStatus(ads, models.AdStatusWorking) {
		return models.AdStatusWorking
	}
	if hasAdStatus(ads, models.AdStatusModeration) {
		return models.AdStatusModeration
	}
	if hasAdStatus(ads, models.AdStatusNotEnoughMoney) {
		return models.AdStatusNotEnoughMoney
	}
	if hasAdStatus(ads, models.AdStatusRejected) {
		return models.AdStatusRejected
	}
	return models.AdStatusTurnedOff
}

func hasAdStatus(ads []*models.Ad, status models.AdStatus) bool {
	for _, ad := range ads {
		if ad != nil && ad.Status == status {
			return true
		}
	}
	return false
}

func (s *Service) userAdStatusTransition(ctx context.Context, current, requested models.AdStatus, campaign *models.AdCampaign) (models.AdStatus, error) {
	switch {
	case current == models.AdStatusWorking && requested == models.AdStatusTurnedOff:
		return models.AdStatusTurnedOff, nil
	case current == models.AdStatusTurnedOff && requested == models.AdStatusWorking:
		if s.canStartAd(ctx, campaign) {
			return models.AdStatusWorking, nil
		}
		return models.AdStatusNotEnoughMoney, nil
	default:
		return "", fmt.Errorf("%w: invalid ad status transition", errs.BusinessLogicError)
	}
}

func (s *Service) recalculateCampaignStatus(ctx context.Context, campaignID int) error {
	campaign, err := s.adCampaignRepo.GetByID(ctx, campaignID)
	if err != nil {
		return err
	}
	ads, err := s.adRepo.ListByAdCampaignID(ctx, campaignID)
	if err != nil {
		return err
	}
	campaign.Status = campaignStatusFromAds(ads)
	campaign.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}
	return s.adCampaignRepo.Update(ctx, campaign)
}

func (s *Service) decorateAdImageURL(ad *models.Ad) {
	if s == nil || s.adStorage == nil || ad == nil {
		return
	}
	if ad.ImageURL == "" || strings.HasPrefix(ad.ImageURL, "http://") || strings.HasPrefix(ad.ImageURL, "https://") {
		return
	}

	imageURL := s.adStorage.GetAdImageURL(ad.ImageURL)
	if imageURL == "" {
		return
	}

	ad.ImageURL = imageURL
}
