package service

import (
	"context"
	"database/sql"
	"eshkere/internal/handler/dto"
	"eshkere/internal/models"
	"time"
)

func (s *Service) CreateAdCampaign(ctx context.Context, c *models.AdCampaign) (*models.AdCampaign, error) {
	if err := s.adCampaignRepo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) UpdateAdCampaign(ctx context.Context, campaignID int, req dto.UpdateAdCampaignRequest) error {
	current, err := s.adCampaignRepo.GetByID(ctx, campaignID)
	if err != nil {
		return err
	}

	if req.Name != nil {
		current.Name = *req.Name
	}
	if req.Status != nil {
		current.Status = *req.Status
	}
	if req.DailyBudget != nil {
		current.DailyBudget = *req.DailyBudget
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
