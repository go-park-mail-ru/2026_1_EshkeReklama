package input

import "eshkere/internal/models"

type CreateAdCampaign struct {
	AdvertiserID int
	Name         string
	DailyBudget  int64
	CPMPrice     int64
	MainAction   string
}

type UpdateAdCampaign struct {
	ID          int
	Name        *string
	Status      *models.AdStatus
	DailyBudget *int64
	CPMPrice    *int64
	MainAction  *string
}
