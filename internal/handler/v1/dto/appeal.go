package dto

import (
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
)

type CreateAppealRequest struct {
	Category    string `json:"category" validate:"required,oneof=bug suggestion complaint question"`
	Title       string `json:"title" validate:"required,max=100"`
	Description string `json:"description" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
}

type CreateAppealResponse struct {
	ID int `json:"id"`
}

func (c *CreateAppealRequest) ToInput() *serviceinput.CreateAppeal {
	return &serviceinput.CreateAppeal{
		Title:       c.Title,
		Description: c.Description,
		Category:    models.AppealCategory(c.Category),
		Name:        c.Name,
		Email:       c.Email,
	}
}

type AppealResponse struct {
	ID          int    `json:"id"`
	Status      string `json:"status"`
	Category    string `json:"category"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func ToAppealResponse(c *models.Appeal) *AppealResponse {
	return &AppealResponse{
		ID:          c.ID,
		Status:      string(c.Status),
		Category:    string(c.Category),
		Title:       c.Title,
		Description: c.Description,
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
