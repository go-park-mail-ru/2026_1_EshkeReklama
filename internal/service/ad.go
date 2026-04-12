package service

import (
	"context"
	"database/sql"
	"eshkere/internal/handler/dto"
	"eshkere/internal/models"
	"time"
)

func (s *Service) CreateAd(ctx context.Context, ad *models.Ad) (*models.Ad, error) {
	ad.Status = models.AdStatusModeration

	if err := s.adRepo.Create(ctx, ad); err != nil {
		return nil, err
	}

	return ad, nil
}

func (s *Service) UpdateAd(ctx context.Context, adID int, req dto.UpdateAdRequest) error {
	currentAd, err := s.adRepo.GetByID(ctx, adID)
	if err != nil {
		return err
	}

	if req.Title != nil {
		currentAd.Title = *req.Title
	}
	if req.Status != nil {
		currentAd.Status = *req.Status
	}
	if req.ShortDesc != nil {
		currentAd.ShortDesc = *req.ShortDesc
	}
	if req.ImageURL != nil {
		currentAd.ImageURL = *req.ImageURL
	}
	if req.TargetURL != nil {
		currentAd.TargetURL = *req.TargetURL
	}
	currentAd.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}

	err = s.adRepo.Update(ctx, currentAd)
	return err
}

func (s *Service) ListAds(ctx context.Context, groupID int) ([]*models.Ad, error) {
	ads, err := s.adRepo.ListByAdGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	return ads, nil
}

func (s *Service) DeleteAd(ctx context.Context, adID int) error {
	err := s.adRepo.Delete(ctx, adID)
	return err
}
