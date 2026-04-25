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
	ID          int       `json:"id"`
	Status      string    `json:"status"`
	Category    string    `json:"category"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func ToAppealResponse(c *models.Appeal) *AppealResponse {
	return &AppealResponse{
		ID:          c.ID,
		Status:      string(c.Status),
		Category:    string(c.Category),
		Title:       c.Title,
		Description: c.Description,
		ImageURL:    c.ImageURL,
		CreatedAt:   c.CreatedAt,
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
