package input

import "eshkere/internal/models"

type CreateAdGroup struct {
	AdCampaignID int
	TopicID      int
	RegionID     int
	Name         string
	AgeFrom      int
	AgeTo        int
	Gender       models.GenderType
}

type UpdateAdGroup struct {
	ID       int
	TopicID  *int
	RegionID *int
	Name     *string
	AgeFrom  *int
	AgeTo    *int
	Gender   *models.GenderType
}
