package dto

import (
	"eshkere/internal/models"
)

type CreateAdRequest struct {
	Title     string `json:"title" validate:"required"`
	ShortDesc string `json:"short_desc" validate:"required"`
	ImageURL  string `json:"image_url" validate:"required"`
	TargetURL string `json:"target_url" validate:"required"`
}

func (c *CreateAdRequest) ToModel() (*models.Ad, error) {
	ad := &models.Ad{
		Title:     c.Title,
		ShortDesc: c.ShortDesc,
		ImageURL:  c.ImageURL,
		TargetURL: c.TargetURL,
	}

	return ad, nil
}

type CreateAdResponse struct {
	ID int `json:"id"`
}

type UpdateAdRequest struct {
	ID        int
	Title     *string          `json:"title" validate:"omitempty"`
	Status    *models.AdStatus `json:"status" validate:"omitempty, oneof=turned_off moderation working rejected not_enough_money"`
	ShortDesc *string          `json:"short_desc" validate:"omitempty"`
	ImageURL  *string          `json:"image_url" validate:"omitempty"`
	TargetURL *string          `json:"target_url" validate:"omitempty"`
}

//func (u *UpdateAdRequest) ToModel(adID int) (*models.UpdateAd, error) {
//	ad := &models.UpdateAd{
//		ID:        adID,
//		Title:     u.Title,
//		Status:    u.Status,
//		ShortDesc: u.ShortDesc,
//		ImageURL:  u.ImageURL,
//		TargetURL: u.TargetURL,
//	}
//	return ad, nil
//}

type AdResponse struct {
	ID        int             `json:"id"`
	Status    models.AdStatus `json:"status"`
	Title     string          `json:"title"`
	ShortDesc string          `json:"short_desc"`
	ImageURL  string          `json:"image_url"`
	TargetURL string          `json:"target_url"`
}

func ToAdResponse(ad *models.Ad) *AdResponse {
	adResponse := &AdResponse{
		ID:        ad.ID,
		Status:    ad.Status,
		Title:     ad.Title,
		ShortDesc: ad.ShortDesc,
		ImageURL:  ad.ImageURL,
		TargetURL: ad.TargetURL,
	}
	return adResponse
}

type ListAdsResponse struct {
	GroupID int           `json:"group_id"`
	Ads     []*AdResponse `json:"ads"`
}
