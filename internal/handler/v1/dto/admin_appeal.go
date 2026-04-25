package dto

import (
	"eshkere/internal/models"
	"time"
)

type AdminAppealResponse struct {
	ID           int     `json:"id"`
	AdvertiserID *int    `json:"advertiser_id,omitempty"`
	Status       string  `json:"status"`
	Category     string  `json:"category"`
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	ImageURL     string  `json:"image_url,omitempty"`
	Name         string  `json:"name"`
	Email        string  `json:"email"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    *string `json:"updated_at,omitempty"`
}

func ToAdminAppealResponse(a *models.Appeal) *AdminAppealResponse {
	var advertiserID *int
	if a.AdvertiserID.Valid {
		v := int(a.AdvertiserID.Int64)
		advertiserID = &v
	}
	var updatedAt *string
	if a.UpdatedAt.Valid {
		s := a.UpdatedAt.Time.Format(time.RFC3339)
		updatedAt = &s
	}
	return &AdminAppealResponse{
		ID:           a.ID,
		AdvertiserID: advertiserID,
		Status:       string(a.Status),
		Category:     string(a.Category),
		Title:        a.Title,
		Description:  a.Description,
		ImageURL:     a.ImageURL,
		Name:         a.Name,
		Email:        a.Email,
		CreatedAt:    a.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    updatedAt,
	}
}

type AdminListAppealsResponse struct {
	Appeals []*AdminAppealResponse `json:"appeals"`
	Limit   int                    `json:"limit"`
	Offset  int                    `json:"offset"`
}

func ToAdminListAppealsResponse(appeals []*models.Appeal, limit, offset int) *AdminListAppealsResponse {
	out := make([]*AdminAppealResponse, 0, len(appeals))
	for _, a := range appeals {
		out = append(out, ToAdminAppealResponse(a))
	}
	return &AdminListAppealsResponse{Appeals: out, Limit: limit, Offset: offset}
}

type AdminAppealMessageResponse struct {
	ID        int    `json:"id"`
	Author    string `json:"author"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
}

func ToAdminAppealMessageResponse(m *models.AppealMessage) *AdminAppealMessageResponse {
	return &AdminAppealMessageResponse{
		ID:        m.ID,
		Author:    string(m.Author),
		Text:      m.Text,
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
	}
}

type AdminAppealStatusHistoryResponse struct {
	ID        int    `json:"id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

func ToAdminAppealStatusHistoryResponse(h *models.AppealStatusHistory) *AdminAppealStatusHistoryResponse {
	return &AdminAppealStatusHistoryResponse{
		ID:        h.ID,
		Status:    string(h.Status),
		CreatedAt: h.CreatedAt.Format(time.RFC3339),
	}
}

type AdminGetAppealResponse struct {
	Appeal        *AdminAppealResponse                `json:"appeal"`
	Messages      []*AdminAppealMessageResponse       `json:"messages"`
	StatusHistory []*AdminAppealStatusHistoryResponse `json:"status_history"`
}

func ToAdminGetAppealResponse(a *models.Appeal, msgs []*models.AppealMessage, hist []*models.AppealStatusHistory) *AdminGetAppealResponse {
	mOut := make([]*AdminAppealMessageResponse, 0, len(msgs))
	for _, m := range msgs {
		mOut = append(mOut, ToAdminAppealMessageResponse(m))
	}
	hOut := make([]*AdminAppealStatusHistoryResponse, 0, len(hist))
	for _, h := range hist {
		hOut = append(hOut, ToAdminAppealStatusHistoryResponse(h))
	}
	return &AdminGetAppealResponse{
		Appeal:        ToAdminAppealResponse(a),
		Messages:      mOut,
		StatusHistory: hOut,
	}
}

type AdminPatchAppealStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=open in_progress closed"`
}

type AdminPostAppealMessageRequest struct {
	Text string `json:"text" validate:"required"`
}
