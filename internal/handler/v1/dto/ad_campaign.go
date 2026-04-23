package dto

import (
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
)

type CreateAdCampaignRequest struct {
	Name        string `json:"name" validate:"required"`
	DailyBudget int64  `json:"daily_budget" validate:"required"`
}

func (c *CreateAdCampaignRequest) ToInput(advertiserID int) *serviceinput.CreateAdCampaign {
	return &serviceinput.CreateAdCampaign{
		AdvertiserID: advertiserID,
		Name:         c.Name,
		DailyBudget:  c.DailyBudget,
	}
}

type CreateAdCampaignResponse struct {
	ID int `json:"id"`
}

type UpdateAdCampaignRequest struct {
	Name        *string          `json:"name" validate:"omitempty"`
	Status      *models.AdStatus `json:"status" validate:"omitempty"`
	DailyBudget *int64           `json:"daily_budget" validate:"omitempty"`
}

func (u *UpdateAdCampaignRequest) ToInput(campaignID int) *serviceinput.UpdateAdCampaign {
	return &serviceinput.UpdateAdCampaign{
		ID:          campaignID,
		Name:        u.Name,
		Status:      u.Status,
		DailyBudget: u.DailyBudget,
	}
}

type AdCampaignResponse struct {
	ID          int             `json:"id"`
	Status      models.AdStatus `json:"status"`
	Name        string          `json:"name"`
	DailyBudget int64           `json:"daily_budget"`
}

func ToAdCampaignResponse(c *models.AdCampaign) *AdCampaignResponse {
	return &AdCampaignResponse{
		ID:          c.ID,
		Status:      c.Status,
		Name:        c.Name,
		DailyBudget: c.DailyBudget,
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
