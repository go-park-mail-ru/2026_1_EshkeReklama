package dto

import "eshkere/internal/models"

type CreateAdCampaignRequest struct {
	Name        string `json:"name" validate:"required"`
	DailyBudget int64  `json:"daily_budget" validate:"required"`
	Title       string `json:"title"`
	ShortDesc   string `json:"short_desc"`
	ImageURL    string `json:"image_url"`
	TargetURL   string `json:"target_url"`
}

func (c *CreateAdCampaignRequest) ToModel(advertiserID int) *models.AdCampaign {
	return &models.AdCampaign{
		AdvertiserID: advertiserID,
		Status:       models.AdStatusModeration,
		Name:         c.Name,
		DailyBudget:  c.DailyBudget,
		Title:        c.Title,
		ShortDesc:    c.ShortDesc,
		ImageURL:     c.ImageURL,
		TargetURL:    c.TargetURL,
	}
}

type CreateAdCampaignResponse struct {
	ID int `json:"id"`
}

type UpdateAdCampaignRequest struct {
	Name        *string          `json:"name" validate:"omitempty"`
	Status      *models.AdStatus `json:"status" validate:"omitempty"`
	DailyBudget *int64           `json:"daily_budget" validate:"omitempty"`
	Title       *string          `json:"title" validate:"omitempty"`
	ShortDesc   *string          `json:"short_desc" validate:"omitempty"`
	ImageURL    *string          `json:"image_url" validate:"omitempty"`
	TargetURL   *string          `json:"target_url" validate:"omitempty"`
}

type AdCampaignResponse struct {
	ID          int             `json:"id"`
	Status      models.AdStatus `json:"status"`
	Name        string          `json:"name"`
	DailyBudget int64           `json:"daily_budget"`
	Title       string          `json:"title"`
	ShortDesc   string          `json:"short_desc"`
	ImageURL    string          `json:"image_url"`
	TargetURL   string          `json:"target_url"`
}

func ToAdCampaignResponse(c *models.AdCampaign) *AdCampaignResponse {
	return &AdCampaignResponse{
		ID:          c.ID,
		Status:      c.Status,
		Name:        c.Name,
		DailyBudget: c.DailyBudget,
		Title:       c.Title,
		ShortDesc:   c.ShortDesc,
		ImageURL:    c.ImageURL,
		TargetURL:   c.TargetURL,
	}
}

type ListAdCampaignsResponse struct {
	AdvertiserID int                   `json:"advertiser_id"`
	Campaigns    []*AdCampaignResponse `json:"campaigns"`
}
