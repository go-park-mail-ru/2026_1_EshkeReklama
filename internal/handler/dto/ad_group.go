package dto

import "eshkere/internal/models"

type CreateAdGroupRequest struct {
	TopicID  int               `json:"topic_id" validate:"required"`
	RegionID int               `json:"region_id" validate:"required"`
	Name     string            `json:"name" validate:"required"`
	AgeFrom  int               `json:"age_from" validate:"required"`
	AgeTo    int               `json:"age_to" validate:"required"`
	Gender   models.GenderType `json:"gender" validate:"required"`
}

func (c *CreateAdGroupRequest) ToModel(campaignID int) *models.AdGroup {
	return &models.AdGroup{
		AdCampaignID: campaignID,
		TopicID:      c.TopicID,
		RegionID:     c.RegionID,
		Name:         c.Name,
		AgeFrom:      c.AgeFrom,
		AgeTo:        c.AgeTo,
		Gender:       c.Gender,
	}
}

type CreateAdGroupResponse struct {
	ID int `json:"id"`
}

type UpdateAdGroupRequest struct {
	TopicID  *int               `json:"topic_id" validate:"omitempty"`
	RegionID *int               `json:"region_id" validate:"omitempty"`
	Name     *string            `json:"name" validate:"omitempty"`
	AgeFrom  *int               `json:"age_from" validate:"omitempty"`
	AgeTo    *int               `json:"age_to" validate:"omitempty"`
	Gender   *models.GenderType `json:"gender" validate:"omitempty"`
}

type AdGroupResponse struct {
	ID       int               `json:"id"`
	TopicID  int               `json:"topic_id"`
	RegionID int               `json:"region_id"`
	Name     string            `json:"name"`
	AgeFrom  int               `json:"age_from"`
	AgeTo    int               `json:"age_to"`
	Gender   models.GenderType `json:"gender"`
}

func ToAdGroupResponse(g *models.AdGroup) *AdGroupResponse {
	return &AdGroupResponse{
		ID:       g.ID,
		TopicID:  g.TopicID,
		RegionID: g.RegionID,
		Name:     g.Name,
		AgeFrom:  g.AgeFrom,
		AgeTo:    g.AgeTo,
		Gender:   g.Gender,
	}
}

type ListAdGroupsResponse struct {
	AdCampaignID int                `json:"ad_campaign_id"`
	Groups       []*AdGroupResponse `json:"groups"`
}
