package dto

import (
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
	"net/url"
	"time"
)

type CreateAppealRequest struct {
	Category    string `json:"category" validate:"required,oneof=bug suggestion complaint question"`
	Title       string `json:"title" validate:"required,max=100"`
	Description string `json:"description" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
}

type UploadedImage struct {
	Data        []byte
	ContentType string
	Ext         string
}

type CreateAppealResponse struct {
	ID int `json:"id"`
}

func NewCreateAppealRequestFromForm(values url.Values) *CreateAppealRequest {
	return &CreateAppealRequest{
		Category:    values.Get("category"),
		Title:       values.Get("title"),
		Description: values.Get("description"),
		Name:        values.Get("name"),
		Email:       values.Get("email"),
	}
}

func (c *CreateAppealRequest) ToInput(uploaded *UploadedImage) *serviceinput.CreateAppeal {
	in := &serviceinput.CreateAppeal{
		Title:       c.Title,
		Description: c.Description,
		Category:    models.AppealCategory(c.Category),
		Name:        c.Name,
		Email:       c.Email,
	}

	if uploaded != nil {
		in.Image = uploaded.Data
		in.ImageExt = uploaded.Ext
		in.ImageType = uploaded.ContentType
	}

	return in
}

type AppealResponse struct {
	ID          int    `json:"id"`
	Status      string `json:"status"`
	Category    string `json:"category"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url,omitempty"`
}

func ToAppealResponse(c *models.Appeal) *AppealResponse {
	return &AppealResponse{
		ID:          c.ID,
		Status:      string(c.Status),
		Category:    string(c.Category),
		Title:       c.Title,
		Description: c.Description,
		ImageURL:    c.ImageURL,
	}
}

type PostAppealMessageRequest struct {
	Text string `json:"text" validate:"required"`
}

type AppealMessageResponse struct {
	ID        int    `json:"id"`
	Author    string `json:"author"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
}

func ToAppealMessageResponse(m *models.AppealMessage) *AppealMessageResponse {
	return &AppealMessageResponse{
		ID:        m.ID,
		Author:    string(m.Author),
		Text:      m.Text,
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
	}
}

type ListAppealMessagesResponse struct {
	AppealID int                      `json:"appeal_id"`
	Messages []*AppealMessageResponse `json:"messages"`
}

func ToListAppealMessagesResponse(appealID int, msgs []*models.AppealMessage) ListAppealMessagesResponse {
	out := make([]*AppealMessageResponse, 0, len(msgs))
	for _, msg := range msgs {
		out = append(out, ToAppealMessageResponse(msg))
	}

	return ListAppealMessagesResponse{
		AppealID: appealID,
		Messages: out,
	}
}

func ToAppealResponses(appeals []*models.Appeal) []*AppealResponse {
	out := make([]*AppealResponse, 0, len(appeals))
	for _, appeal := range appeals {
		out = append(out, ToAppealResponse(appeal))
	}
	return out
}

type ListAppealsResponse struct {
	AdvertiserID int               `json:"advertiser_id"`
	Appeals      []*AppealResponse `json:"appeals"`
}

func ToListAppealsResponse(advertiserID int, appeals []*models.Appeal) ListAppealsResponse {
	return ListAppealsResponse{
		AdvertiserID: advertiserID,
		Appeals:      ToAppealResponses(appeals),
	}
}
