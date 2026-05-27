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

func (s *Service) CreateAdGroup(ctx context.Context, advertiserID int, in *serviceinput.CreateAdGroup) (*models.AdGroup, error) {
	if _, err := s.ownedCampaign(ctx, advertiserID, in.AdCampaignID); err != nil {
		return nil, err
	}

	topicID, err := s.resolveTopicID(ctx, in.Topic, in.TopicID)
	if err != nil {
		return nil, err
	}
	regionID, err := s.resolveRegionID(ctx, in.Region, in.RegionID)
	if err != nil {
		return nil, err
	}

	g := &models.AdGroup{
		AdCampaignID: in.AdCampaignID,
		TopicID:      topicID,
		RegionID:     regionID,
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
	} else if in.Topic != nil {
		topicID, err := s.resolveTopicID(ctx, *in.Topic, 0)
		if err != nil {
			return err
		}
		current.TopicID = topicID
	}
	if in.RegionID != nil {
		current.RegionID = *in.RegionID
	} else if in.Region != nil {
		regionID, err := s.resolveRegionID(ctx, *in.Region, 0)
		if err != nil {
			return err
		}
		current.RegionID = regionID
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

func (s *Service) resolveTopicID(ctx context.Context, topic string, fallbackID int) (int, error) {
	topic = strings.TrimSpace(topic)
	if topic == "" {
		if fallbackID > 0 {
			return fallbackID, nil
		}
		return 0, fmt.Errorf("%w: topic is required", errs.BadRequestError)
	}
	if s.topicRepo == nil {
		return 0, fmt.Errorf("%w: topic repository is not configured", errs.InternalServiceError)
	}
	id, err := s.topicRepo.GetIDByName(ctx, topic)
	if err != nil {
		return 0, fmt.Errorf("%w: unknown topic %q", errs.BadRequestError, topic)
	}
	return id, nil
}

func (s *Service) resolveRegionID(ctx context.Context, region string, fallbackID int) (int, error) {
	region = strings.TrimSpace(region)
	if region == "" {
		if fallbackID > 0 {
			return fallbackID, nil
		}
		return 0, fmt.Errorf("%w: region is required", errs.BadRequestError)
	}
	if s.regionRepo == nil {
		return 0, fmt.Errorf("%w: region repository is not configured", errs.InternalServiceError)
	}
	id, err := s.regionRepo.GetIDByName(ctx, region)
	if err != nil {
		return 0, fmt.Errorf("%w: unknown region %q", errs.BadRequestError, region)
	}
	return id, nil
}
