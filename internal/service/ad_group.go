package service

import (
	"context"
	"database/sql"
	"time"

	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
)

func (s *Service) CreateAdGroup(ctx context.Context, advertiserID int, in *serviceinput.CreateAdGroup) (*models.AdGroup, error) {
	if _, err := s.ownedCampaign(ctx, advertiserID, in.AdCampaignID); err != nil {
		return nil, err
	}
	g := &models.AdGroup{
		AdCampaignID: in.AdCampaignID,
		TopicID:      in.TopicID,
		RegionID:     in.RegionID,
		Name:         in.Name,
		AgeFrom:      in.AgeFrom,
		AgeTo:        in.AgeTo,
		Gender:       in.Gender,
	}

	if err := s.adGroupRepo.Create(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *Service) UpdateAdGroup(ctx context.Context, advertiserID int, in *serviceinput.UpdateAdGroup) error {
	current, err := s.adGroupRepo.GetByID(ctx, in.ID)
	if err != nil {
		return err
	}
	if _, err := s.ownedCampaign(ctx, advertiserID, current.AdCampaignID); err != nil {
		return err
	}

	if in.TopicID != nil {
		current.TopicID = *in.TopicID
	}
	if in.RegionID != nil {
		current.RegionID = *in.RegionID
	}
	if in.Name != nil {
		current.Name = *in.Name
	}
	if in.AgeFrom != nil {
		current.AgeFrom = *in.AgeFrom
	}
	if in.AgeTo != nil {
		current.AgeTo = *in.AgeTo
	}
	if in.Gender != nil {
		current.Gender = *in.Gender
	}
	current.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}

	return s.adGroupRepo.Update(ctx, current)
}

func (s *Service) ListAdGroups(ctx context.Context, advertiserID, campaignID int) ([]*models.AdGroup, error) {
	if _, err := s.ownedCampaign(ctx, advertiserID, campaignID); err != nil {
		return nil, err
	}
	return s.adGroupRepo.ListByCampaignID(ctx, campaignID)
}

func (s *Service) DeleteAdGroup(ctx context.Context, advertiserID, groupID int) error {
	group, err := s.ownedGroup(ctx, advertiserID, groupID)
	if err != nil {
		return err
	}
	if err := s.adGroupRepo.Delete(ctx, groupID); err != nil {
		return err
	}
	return s.recalculateCampaignStatus(ctx, group.AdCampaignID)
}
