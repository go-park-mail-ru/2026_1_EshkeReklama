package dto

import (
	"time"

	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
)

type CreatePartnerSiteRequest struct {
	Domain   string `json:"domain" validate:"required"`
	SiteName string `json:"site_name" validate:"required"`
}

func (r *CreatePartnerSiteRequest) ToInput(partnerID int) *serviceinput.CreatePartnerSite {
	return &serviceinput.CreatePartnerSite{
		PartnerID: partnerID,
		Domain:    r.Domain,
		SiteName:  r.SiteName,
	}
}

type UpdatePartnerSiteRequest struct {
	Domain   *string `json:"domain"`
	SiteName *string `json:"site_name"`
	Status   *string `json:"status"`
}

func (r *UpdatePartnerSiteRequest) ToInput(partnerID, siteID int) *serviceinput.UpdatePartnerSite {
	var status *models.PartnerSiteStatus
	if r.Status != nil {
		value := models.PartnerSiteStatus(*r.Status)
		status = &value
	}
	return &serviceinput.UpdatePartnerSite{
		ID:        siteID,
		PartnerID: partnerID,
		Domain:    r.Domain,
		SiteName:  r.SiteName,
		Status:    status,
	}
}

type PartnerSiteResponse struct {
	ID        int    `json:"id"`
	Domain    string `json:"domain"`
	SiteName  string `json:"site_name"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

func ToPartnerSiteResponse(site *models.PartnerSite) *PartnerSiteResponse {
	if site == nil {
		return nil
	}
	resp := &PartnerSiteResponse{
		ID:        site.ID,
		Domain:    site.Domain,
		SiteName:  site.SiteName,
		Status:    string(site.Status),
		CreatedAt: site.CreatedAt.Format(time.RFC3339),
	}
	if site.UpdatedAt.Valid {
		resp.UpdatedAt = site.UpdatedAt.Time.Format(time.RFC3339)
	}
	return resp
}

type ListPartnerSitesResponse struct {
	PartnerID int                    `json:"partner_id"`
	Sites     []*PartnerSiteResponse `json:"sites"`
}

func ToPartnerSiteResponses(sites []*models.PartnerSite) []*PartnerSiteResponse {
	out := make([]*PartnerSiteResponse, 0, len(sites))
	for _, site := range sites {
		out = append(out, ToPartnerSiteResponse(site))
	}
	return out
}
