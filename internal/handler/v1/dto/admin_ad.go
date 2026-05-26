package dto

import (
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
)

type UpdateAdStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=approve disapprove"`
}

func (r *UpdateAdStatusRequest) ToInput(adID int) *serviceinput.UpdateAdStatus {
	return &serviceinput.UpdateAdStatus{
		AdID:     adID,
		Decision: serviceinput.AdModerationDecision(r.Status),
	}
}

type ListAdminAdsResponse struct {
	Ads []*AdResponse `json:"ads"`
}

func ToListAdminAdsResponse(ads []*models.Ad) ListAdminAdsResponse {
	return ListAdminAdsResponse{
		Ads: ToAdResponses(ads),
	}
}
