package dto

import (
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
)

type UpdateAdStatusRequest struct {
	Status  string `json:"status" validate:"required,oneof=approve disapprove"`
	Message string `json:"message" validate:"omitempty,max=2000"`
}

func (r *UpdateAdStatusRequest) ToInput(adID int) *serviceinput.UpdateAdStatus {
	return &serviceinput.UpdateAdStatus{
		AdID:     adID,
		Decision: serviceinput.AdModerationDecision(r.Status),
		Message:  r.Message,
	}
}

type AdminAdResponse struct {
	ID                 int    `json:"id"`
	Status             string `json:"status"`
	Title              string `json:"title"`
	ShortDesc          string `json:"short_desc"`
	ImageURL           string `json:"image_url"`
	TargetURL          string `json:"target_url"`
	PriorityModeration bool   `json:"priority_moderation"`
}

type ListAdminAdsResponse struct {
	Ads []*AdminAdResponse `json:"ads"`
}

func ToAdminAdResponse(item *models.ModerationQueueItem) *AdminAdResponse {
	if item == nil || item.Ad == nil {
		return nil
	}
	base := ToAdResponse(item.Ad)
	return &AdminAdResponse{
		ID:                 base.ID,
		Status:             base.Status,
		Title:              base.Title,
		ShortDesc:          base.ShortDesc,
		ImageURL:           base.ImageURL,
		TargetURL:          base.TargetURL,
		PriorityModeration: item.PriorityModeration,
	}
}

func ToListAdminAdsResponse(items []*models.ModerationQueueItem) ListAdminAdsResponse {
	ads := make([]*AdminAdResponse, 0, len(items))
	for _, item := range items {
		if resp := ToAdminAdResponse(item); resp != nil {
			ads = append(ads, resp)
		}
	}
	return ListAdminAdsResponse{Ads: ads}
}
