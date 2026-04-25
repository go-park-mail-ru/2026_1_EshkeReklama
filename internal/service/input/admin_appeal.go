package input

import "eshkere/internal/models"

type AdminListAppealsFilter struct {
	AdvertiserID *int
	Status       *models.AppealStatus
	Category     *models.AppealCategory
	Limit        int
	Offset       int
}

type AdminPostAppealMessage struct {
	AppealID int
	Text     string
}

type AdminPatchAppealStatus struct {
	AppealID int
	Status   models.AppealStatus
}
