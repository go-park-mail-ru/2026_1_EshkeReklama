package dto

import (
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
	"net/url"
)

type CreateAdRequest struct {
	Title     string `json:"title" validate:"required"`
	ShortDesc string `json:"short_desc" validate:"required"`
	TargetURL string `json:"target_url" validate:"required"`
}

func NewCreateAdRequestFromForm(values url.Values) *CreateAdRequest {
	return &CreateAdRequest{
		Title:     values.Get("title"),
		ShortDesc: values.Get("short_desc"),
		TargetURL: values.Get("target_url"),
	}
}

func (c *CreateAdRequest) ToInput(groupID int) *serviceinput.CreateAd {
	return &serviceinput.CreateAd{
		AdGroupID: groupID,
		Title:     c.Title,
		ShortDesc: c.ShortDesc,
		TargetURL: c.TargetURL,
	}
}

type CreateAdResponse struct {
	ID int `json:"id"`
}

type UpdateAdRequest struct {
	ID        int
	Title     *string `json:"title" validate:"omitempty,min=1"`
	Status    *string `json:"status" validate:"omitempty,min=1,oneof=turned_off moderation working rejected not_enough_money"`
	ShortDesc *string `json:"short_desc" validate:"omitempty,min=1"`
	TargetURL *string `json:"target_url" validate:"omitempty,min=1"`
}

func NewUpdateAdRequestFromForm(values url.Values) *UpdateAdRequest {
	return &UpdateAdRequest{
		Title:     optionalStringFromForm(values, "title"),
		Status:    optionalStringFromForm(values, "status"),
		ShortDesc: optionalStringFromForm(values, "short_desc"),
		TargetURL: optionalStringFromForm(values, "target_url"),
	}
}

func optionalStringFromForm(values url.Values, key string) *string {
	raw, ok := values[key]
	if !ok || len(raw) == 0 {
		return nil
	}

	value := raw[0]
	return &value
}

func (u *UpdateAdRequest) ToInput(adID int) *serviceinput.UpdateAd {
	return &serviceinput.UpdateAd{
		ID:        adID,
		Title:     u.Title,
		Status:    (*models.AdStatus)(u.Status),
		ShortDesc: u.ShortDesc,
		TargetURL: u.TargetURL,
	}
}

type AdResponse struct {
	ID        int    `json:"id"`
	Status    string `json:"status"`
	Title     string `json:"title"`
	ShortDesc string `json:"short_desc"`
	ImageURL  string `json:"image_url"`
	TargetURL string `json:"target_url"`
}

func ToAdResponse(ad *models.Ad) *AdResponse {
	adResponse := &AdResponse{
		ID:        ad.ID,
		Status:    string(ad.Status),
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

type AdRequest struct {
	EmbedToken string `json:"embed_token" validate:"required"`
	VisitorID  string `json:"visitor_id"`
}

type AdRequestResponse struct {
	RequestID string      `json:"request_id"`
	Ad        *AdResponse `json:"ad"`
	ClickURL  string      `json:"click_url"`
}
