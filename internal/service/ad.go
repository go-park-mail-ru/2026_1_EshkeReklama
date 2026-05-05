package service

import (
	"context"
	"database/sql"
	errs "eshkere/internal/errors"
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
	"fmt"
	"strings"
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

	if len(in.Image) > 0 {
		if s.adStorage == nil {
			return nil, fmt.Errorf("%w: ad storage is not configured", errs.InternalServiceError)
		}

		imageKey, err := s.adStorage.UploadAdImage(ctx, in.Image, in.ImageExt, in.ImageType)
		if err != nil {
			return nil, err
		}

		ad.ImageURL = imageKey
	}

	if err := s.adRepo.Create(ctx, ad); err != nil {
		if ad.ImageURL != "" && s.adStorage != nil {
			_ = s.adStorage.DeleteAdImage(ctx, ad.ImageURL)
		}
		return nil, err
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

		imageKey, err := s.adStorage.UploadAdImage(ctx, *in.Image, imageExt, imageType)
		if err != nil {
			return err
		}
		currentAd.ImageURL = imageKey
	}
	currentAd.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}

	if err := s.adRepo.Update(ctx, currentAd); err != nil {
		if in.Image != nil && len(*in.Image) > 0 && s.adStorage != nil {
			_ = s.adStorage.DeleteAdImage(ctx, currentAd.ImageURL)
		}
		return err
	}

	return nil
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
	if ad.ImageURL == "" || strings.HasPrefix(ad.ImageURL, "http://") || strings.HasPrefix(ad.ImageURL, "https://") {
		return
	}

	imageURL := s.adStorage.GetAdImageURL(ad.ImageURL)
	if imageURL == "" {
		return
	}

	ad.ImageURL = imageURL
}
