package service

import (
	"context"
	"database/sql"
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
	"time"
)

func (s *Service) CreateAd(ctx context.Context, in *serviceinput.CreateAd) (*models.Ad, error) {
	ad := &models.Ad{
		AdGroupID: in.AdGroupID,
		Status:    models.AdStatusModeration,
		Title:     in.Title,
		ShortDesc: in.ShortDesc,
		ImageURL:  in.ImageURL,
		TargetURL: in.TargetURL,
	}

	if err := s.adRepo.Create(ctx, ad); err != nil {
		return nil, err
	}

	return ad, nil
}

func (s *Service) UpdateAd(ctx context.Context, in *serviceinput.UpdateAd) error {
	currentAd, err := s.adRepo.GetByID(ctx, in.ID)
	if err != nil {
		return err
	}

	if in.Title != nil {
		currentAd.Title = *in.Title
	}
	if in.Status != nil {
		currentAd.Status = *in.Status
	}
	if in.ShortDesc != nil {
		currentAd.ShortDesc = *in.ShortDesc
	}
	if in.ImageURL != nil {
		currentAd.ImageURL = *in.ImageURL
	}
	if in.TargetURL != nil {
		currentAd.TargetURL = *in.TargetURL
	}
	currentAd.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}

	return s.adRepo.Update(ctx, currentAd)
}

func (s *Service) ListAds(ctx context.Context, groupID int) ([]*models.Ad, error) {
	return s.adRepo.ListByAdGroupID(ctx, groupID)
}

func (s *Service) DeleteAd(ctx context.Context, adID int) error {
	return s.adRepo.Delete(ctx, adID)
}
