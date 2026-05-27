package input

import "eshkere/internal/models"

type CreateAdGroup struct {
	AdCampaignID int
	Topic        string
	Region       string
	TopicID      int
	RegionID     int
	Name         string
	AgeFrom      int
	AgeTo        int
	Gender       models.GenderType
}

type UpdateAdGroup struct {
	ID       int
	Topic    *string
	Region   *string
	TopicID  *int
	RegionID *int
	Name     *string
	AgeFrom  *int
	AgeTo    *int
	Gender   *models.GenderType
}
