package dto

import (
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
)

type CreateAdGroupRequest struct {
	Topic   string `json:"topic" validate:"required"`
	Region  string `json:"region" validate:"required"`
	Name    string `json:"name" validate:"required"`
	AgeFrom int    `json:"age_from" validate:"required"`
	AgeTo   int    `json:"age_to" validate:"required"`
	Gender  string `json:"gender" validate:"required,oneof=man woman any"`
}

func (c *CreateAdGroupRequest) ToInput(campaignID int) *serviceinput.CreateAdGroup {
	return &serviceinput.CreateAdGroup{
		AdCampaignID: campaignID,
		Topic:        c.Topic,
		Region:       c.Region,
		Name:         c.Name,
		AgeFrom:      c.AgeFrom,
		AgeTo:        c.AgeTo,
		Gender:       models.GenderType(c.Gender),
	}
}

type CreateAdGroupResponse struct {
	ID int `json:"id"`
}

type UpdateAdGroupRequest struct {
	Topic   *string `json:"topic" validate:"omitempty,min=1"`
	Region  *string `json:"region" validate:"omitempty,min=1"`
	Name    *string `json:"name" validate:"omitempty,min=1"`
	AgeFrom *int    `json:"age_from" validate:"omitempty,min=1"`
	AgeTo   *int    `json:"age_to" validate:"omitempty,min=1"`
	Gender  *string `json:"gender" validate:"omitempty,min=1,oneof=man woman any"`
}

func (u *UpdateAdGroupRequest) ToInput(groupID int) *serviceinput.UpdateAdGroup {
	return &serviceinput.UpdateAdGroup{
		ID:      groupID,
		Topic:   u.Topic,
		Region:  u.Region,
		Name:    u.Name,
		AgeFrom: u.AgeFrom,
		AgeTo:   u.AgeTo,
		Gender:  (*models.GenderType)(u.Gender),
	}
}

type AdGroupResponse struct {
	ID      int    `json:"id"`
	Topic   string `json:"topic"`
	Region  string `json:"region"`
	Name    string `json:"name"`
	AgeFrom int    `json:"age_from"`
	AgeTo   int    `json:"age_to"`
	Gender  string `json:"gender"`
}

func ToAdGroupResponse(g *models.AdGroup) *AdGroupResponse {
	return &AdGroupResponse{
		ID:      g.ID,
		Topic:   g.Topic,
		Region:  g.Region,
		Name:    g.Name,
		AgeFrom: g.AgeFrom,
		AgeTo:   g.AgeTo,
		Gender:  string(g.Gender),
	}
}

func ToAdGroupResponses(groups []*models.AdGroup) []*AdGroupResponse {
	out := make([]*AdGroupResponse, 0, len(groups))
	for _, group := range groups {
		out = append(out, ToAdGroupResponse(group))
	}
	return out
}

type ListAdGroupsResponse struct {
	AdCampaignID int                `json:"ad_campaign_id"`
	Groups       []*AdGroupResponse `json:"groups"`
}

func ToListAdGroupsResponse(campaignID int, groups []*models.AdGroup) ListAdGroupsResponse {
	return ListAdGroupsResponse{
		AdCampaignID: campaignID,
		Groups:       ToAdGroupResponses(groups),
	}
}
