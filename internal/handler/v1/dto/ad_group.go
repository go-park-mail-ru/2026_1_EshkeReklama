package dto

import (
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
)

type CreateAdGroupRequest struct {
	TopicID  int               `json:"topic_id" validate:"required"`
	RegionID int               `json:"region_id" validate:"required"`
	Name     string            `json:"name" validate:"required"`
	AgeFrom  int               `json:"age_from" validate:"required"`
	AgeTo    int               `json:"age_to" validate:"required"`
	Gender   models.GenderType `json:"gender" validate:"required"`
}

func (c *CreateAdGroupRequest) ToInput(campaignID int) *serviceinput.CreateAdGroup {
	return &serviceinput.CreateAdGroup{
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
	TopicID  *int               `json:"topic_id" validate:"omitempty,min=1"`
	RegionID *int               `json:"region_id" validate:"omitempty,min=1"`
	Name     *string            `json:"name" validate:"omitempty,min=1"`
	AgeFrom  *int               `json:"age_from" validate:"omitempty,min=1"`
	AgeTo    *int               `json:"age_to" validate:"omitempty,min=1"`
	Gender   *models.GenderType `json:"gender" validate:"omitempty,min=1"`
}

func (u *UpdateAdGroupRequest) ToInput(groupID int) *serviceinput.UpdateAdGroup {
	return &serviceinput.UpdateAdGroup{
		ID:       groupID,
		TopicID:  u.TopicID,
		RegionID: u.RegionID,
		Name:     u.Name,
		AgeFrom:  u.AgeFrom,
		AgeTo:    u.AgeTo,
		Gender:   u.Gender,
	}
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
