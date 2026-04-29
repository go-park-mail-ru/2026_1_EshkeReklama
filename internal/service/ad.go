package service

import (
	"context"
	"database/sql"
	errs "eshkere/internal/errors"
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
	"fmt"
	"time"
)

func (s *Service) CreateAd(ctx context.Context, in *serviceinput.CreateAd) (*models.Ad, error) {
	ad := &models.Ad{
		AdGroupID: in.AdGroupID,
		Status:    models.AdStatusModeration,
		Title:     in.Title,
		ShortDesc: in.ShortDesc,
		TargetURL: in.TargetURL,
	}

	if err := s.adRepo.Create(ctx, ad); err != nil {
		return nil, err
	}

	if len(in.Image) > 0 {
		if s.adStorage == nil {
			return nil, fmt.Errorf("%w: ad storage is not configured", errs.InternalServiceError)
		}

		imageKey, err := s.adStorage.UploadAdImage(ctx, ad.ID, in.Image, in.ImageExt, in.ImageType)
		if err != nil {
			return nil, err
		}

		if err = s.adRepo.UpdateImage(ctx, ad.ID, imageKey); err != nil {
			_ = s.adStorage.DeleteAdImage(ctx, ad.ID, imageKey)
			return nil, err
		}

		ad.ImageURL = imageKey
	}

	s.decorateAdImageURL(ad)

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
	if in.TargetURL != nil {
		currentAd.TargetURL = *in.TargetURL
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

		imageKey, err := s.adStorage.UploadAdImage(ctx, currentAd.ID, *in.Image, imageExt, imageType)
		if err != nil {
			return err
		}
		currentAd.ImageURL = imageKey
	}
	currentAd.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}

	return s.adRepo.Update(ctx, currentAd)
}

func (s *Service) ListAds(ctx context.Context, groupID int) ([]*models.Ad, error) {
	ads, err := s.adRepo.ListByAdGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	for _, ad := range ads {
		s.decorateAdImageURL(ad)
	}

	return ads, nil
}

func (s *Service) DeleteAd(ctx context.Context, adID int) error {
	return s.adRepo.Delete(ctx, adID)
}

func (s *Service) decorateAdImageURL(ad *models.Ad) {
	if s == nil || s.adStorage == nil || ad == nil {
		return
	}

	imageURL := s.adStorage.GetAdImageURL(ad.ImageURL)
	if imageURL == "" {
		return
	}

	ad.ImageURL = imageURL
}
