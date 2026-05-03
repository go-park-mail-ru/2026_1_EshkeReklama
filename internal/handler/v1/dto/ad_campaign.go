package dto

import (
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
)

type CreateAdCampaignRequest struct {
	Name        string `json:"name" validate:"required"`
	DailyBudget int64  `json:"daily_budget" validate:"omitempty,gte=0"`
	CPMPrice    int64  `json:"cpm_price" validate:"omitempty,gte=0"`
	MainAction  string `json:"main_action" validate:"omitempty"`
}

func (c *CreateAdCampaignRequest) ToInput(advertiserID int) *serviceinput.CreateAdCampaign {
	return &serviceinput.CreateAdCampaign{
		AdvertiserID: advertiserID,
		Name:         c.Name,
		DailyBudget:  c.DailyBudget,
		CPMPrice:     c.CPMPrice,
		MainAction:   c.MainAction,
	}
}

type CreateAdCampaignResponse struct {
	ID int `json:"id"`
}

type UpdateAdCampaignRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=1"`
	Status      *string `json:"status" validate:"omitempty,min=1"`
	DailyBudget *int64  `json:"daily_budget" validate:"omitempty,gte=0"`
	CPMPrice    *int64  `json:"cpm_price" validate:"omitempty,gte=0"`
	MainAction  *string `json:"main_action" validate:"omitempty,min=1"`
}

func (u *UpdateAdCampaignRequest) ToInput(campaignID int) *serviceinput.UpdateAdCampaign {
	return &serviceinput.UpdateAdCampaign{
		ID:          campaignID,
		Name:        u.Name,
		DailyBudget: u.DailyBudget,
		CPMPrice:    u.CPMPrice,
		MainAction:  u.MainAction,
		Status:      (*models.AdStatus)(u.Status),
	}
}

type AdCampaignResponse struct {
	ID          int    `json:"id"`
	Status      string `json:"status"`
	Name        string `json:"name"`
	DailyBudget int64  `json:"daily_budget"`
	CPMPrice    int64  `json:"cpm_price"`
	MainAction  string `json:"main_action"`
}

func ToAdCampaignResponse(c *models.AdCampaign) *AdCampaignResponse {
	return &AdCampaignResponse{
		ID:          c.ID,
		Status:      string(c.Status),
		Name:        c.Name,
		MainAction:  c.MainAction,
		DailyBudget: c.DailyBudget,
		CPMPrice:    c.CPMPrice,
	}
}

func ToAdCampaignResponses(campaigns []*models.AdCampaign) []*AdCampaignResponse {
	out := make([]*AdCampaignResponse, 0, len(campaigns))
	for _, campaign := range campaigns {
		out = append(out, ToAdCampaignResponse(campaign))
	}
	return out
}

type ListAdCampaignsResponse struct {
	AdvertiserID int                   `json:"advertiser_id"`
	Campaigns    []*AdCampaignResponse `json:"campaigns"`
}

func ToListAdCampaignsResponse(advertiserID int, campaigns []*models.AdCampaign) ListAdCampaignsResponse {
	return ListAdCampaignsResponse{
		AdvertiserID: advertiserID,
		Campaigns:    ToAdCampaignResponses(campaigns),
	}
}
