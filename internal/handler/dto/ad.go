package dto

import (
	"errors"
	"eshkere/internal/models"
	"time"
)

type CreateAdRequest struct {
	AdGroupID int    `json:"ad_group_id" validate:"required"`
	Title     string `json:"title" validate:"required"`
	ShortDesc string `json:"short_desc" validate:"required"`
	ImageURL  string `json:"image_url" validate:"required"`
	TargetURL string `json:"target_url" validate:"required"`
}
type CreateAdResponse struct {
	ID int `json:"id"`
}

func (car *CreateAdRequest) ToModel() (*models.Ad, error) {
	ad := &models.Ad{
		AdGroupID: car.AdGroupID,
		Title:     car.Title,
		ShortDesc: car.ShortDesc,
		ImageURL:  car.ImageURL,
		TargetURL: car.TargetURL,
	}

	return ad, nil
}

type ListAdsResponse struct {
	AdvertiserID int          `json:"advertiser_id"`
	Ads          []AdResponse `json:"ads"`
}
