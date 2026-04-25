package dto

import (
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
)

type CreateAdRequest struct {
	Title     string `json:"title" validate:"required"`
	ShortDesc string `json:"short_desc" validate:"required"`
	ImageURL  string `json:"image_url" validate:"required"`
	TargetURL string `json:"target_url" validate:"required"`
}

func (c *CreateAdRequest) ToInput(groupID int) *serviceinput.CreateAd {
	return &serviceinput.CreateAd{
		AdGroupID: groupID,
		Title:     c.Title,
		ShortDesc: c.ShortDesc,
		ImageURL:  c.ImageURL,
		TargetURL: c.TargetURL,
	}
}

type CreateAdResponse struct {
	ID int `json:"id"`
}

type UpdateAdRequest struct {
	ID        int
	Title     *string          `json:"title" validate:"omitempty,min=1"`
	Status    *models.AdStatus `json:"status" validate:"omitempty,min=1,oneof=turned_off moderation working rejected not_enough_money"`
	ShortDesc *string          `json:"short_desc" validate:"omitempty,min=1"`
	ImageURL  *string          `json:"image_url" validate:"omitempty,min=1"`
	TargetURL *string          `json:"target_url" validate:"omitempty,min=1"`
}

func (u *UpdateAdRequest) ToInput(adID int) *serviceinput.UpdateAd {
	return &serviceinput.UpdateAd{
		ID:        adID,
		Title:     u.Title,
		Status:    u.Status,
		ShortDesc: u.ShortDesc,
		ImageURL:  u.ImageURL,
		TargetURL: u.TargetURL,
	}
}

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

func ToAdResponses(ads []*models.Ad) []*AdResponse {
	out := make([]*AdResponse, 0, len(ads))
	for _, ad := range ads {
		out = append(out, ToAdResponse(ad))
	}
	return out
}

type ListAdsResponse struct {
	GroupID int           `json:"group_id"`
	Ads     []*AdResponse `json:"ads"`
}

func ToListAdsResponse(groupID int, ads []*models.Ad) ListAdsResponse {
	return ListAdsResponse{
		GroupID: groupID,
		Ads:     ToAdResponses(ads),
	}
}
