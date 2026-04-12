package service

import (
	"context"
	"database/sql"
	"eshkere/internal/handler/dto"
	"eshkere/internal/models"
	"time"
)

func (s *Service) CreateAdGroup(ctx context.Context, g *models.AdGroup) (*models.AdGroup, error) {
	if err := s.adGroupRepo.Create(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *Service) UpdateAdGroup(ctx context.Context, groupID int, req dto.UpdateAdGroupRequest) error {
	current, err := s.adGroupRepo.GetByID(ctx, groupID)
	if err != nil {
		return err
	}

	if req.TopicID != nil {
		current.TopicID = *req.TopicID
	}
	if req.RegionID != nil {
		current.RegionID = *req.RegionID
	}
	if req.Name != nil {
		current.Name = *req.Name
	}
	if req.AgeFrom != nil {
		current.AgeFrom = *req.AgeFrom
	}
	if req.AgeTo != nil {
		current.AgeTo = *req.AgeTo
	}
	if req.Gender != nil {
		current.Gender = *req.Gender
	}
	current.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}

	return s.adGroupRepo.Update(ctx, current)
}

func (s *Service) ListAdGroups(ctx context.Context, campaignID int) ([]*models.AdGroup, error) {
	return s.adGroupRepo.ListByCampaignID(ctx, campaignID)
}

func (s *Service) DeleteAdGroup(ctx context.Context, groupID int) error {
	return s.adGroupRepo.Delete(ctx, groupID)
}
