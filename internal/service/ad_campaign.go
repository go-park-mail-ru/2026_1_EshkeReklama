package service

import (
	"context"
	"database/sql"
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
	"time"
)

func (s *Service) CreateAdCampaign(ctx context.Context, in *serviceinput.CreateAdCampaign) (*models.AdCampaign, error) {
	c := &models.AdCampaign{
		AdvertiserID: in.AdvertiserID,
		Status:       models.AdStatusModeration,
		Name:         in.Name,
		DailyBudget:  in.DailyBudget,
		CPMPrice:     in.CPMPrice,
		MainAction:   in.MainAction,
	}

	if err := s.adCampaignRepo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) UpdateAdCampaign(ctx context.Context, advertiserID int, in *serviceinput.UpdateAdCampaign) error {
	current, err := s.adCampaignRepo.GetByID(ctx, in.ID)
	if err != nil {
		return err
	}
	if current.AdvertiserID != advertiserID {
		return ownershipNotFound("ad campaign")
	}

	if in.Name != nil {
		current.Name = *in.Name
	}
	if in.DailyBudget != nil {
		current.DailyBudget = *in.DailyBudget
	}
	if in.CPMPrice != nil {
		current.CPMPrice = *in.CPMPrice
	}
	if in.MainAction != nil {
		current.MainAction = *in.MainAction
	}
	current.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}

	return s.adCampaignRepo.Update(ctx, current)
}

func (s *Service) ListAdCampaigns(ctx context.Context, advertiserID int) ([]*models.AdCampaign, error) {
	return s.adCampaignRepo.ListByAdvertiserID(ctx, advertiserID)
}

func (s *Service) DeleteAdCampaign(ctx context.Context, advertiserID, campaignID int) error {
	if _, err := s.ownedCampaign(ctx, advertiserID, campaignID); err != nil {
		return err
	}
	return s.adCampaignRepo.Delete(ctx, campaignID)
}
