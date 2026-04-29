package input

import "eshkere/internal/models"

type CreateAdCampaign struct {
	AdvertiserID int
	Name         string
	MainAction   string
}

type UpdateAdCampaign struct {
	ID         int
	Name       *string
	Status     *models.AdStatus
	MainAction *string
}
