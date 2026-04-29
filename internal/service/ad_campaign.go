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
		MainAction:   in.MainAction,
	}

	if err := s.adCampaignRepo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) UpdateAdCampaign(ctx context.Context, in *serviceinput.UpdateAdCampaign) error {
	current, err := s.adCampaignRepo.GetByID(ctx, in.ID)
	if err != nil {
		return err
	}

	if in.Name != nil {
		current.Name = *in.Name
	}
	if in.Status != nil {
		current.Status = *in.Status
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

func (s *Service) DeleteAdCampaign(ctx context.Context, campaignID int) error {
	return s.adCampaignRepo.Delete(ctx, campaignID)
}
