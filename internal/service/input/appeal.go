package input

import "eshkere/internal/models"

type CreateAppeal struct {
	AdvertiserID *int
	Category     models.AppealCategory
	Title        string
	Description  string
	Name         string
	Email        string
}
