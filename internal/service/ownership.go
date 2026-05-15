package service

import (
	"context"
	errs "eshkere/internal/errors"
	"eshkere/internal/models"
	"fmt"
)

func ownershipNotFound(entity string) error {
	return fmt.Errorf("%w: %s not found", errs.NotFoundError, entity)
}

func (s *Service) ownedCampaign(ctx context.Context, advertiserID, campaignID int) (*models.AdCampaign, error) {
	campaign, err := s.adCampaignRepo.GetByID(ctx, campaignID)
	if err != nil {
		return nil, err
	}
	if campaign.AdvertiserID != advertiserID {
		return nil, ownershipNotFound("ad campaign")
	}
	return campaign, nil
}

func (s *Service) ownedGroup(ctx context.Context, advertiserID, groupID int) (*models.AdGroup, error) {
	group, err := s.adGroupRepo.GetByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if _, err := s.ownedCampaign(ctx, advertiserID, group.AdCampaignID); err != nil {
		return nil, err
	}
	return group, nil
}
