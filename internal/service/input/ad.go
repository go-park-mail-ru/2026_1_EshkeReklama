package input

import "eshkere/internal/models"

type CreateAd struct {
	AdGroupID int
	Title     string
	ShortDesc string
	ImageURL  string
	TargetURL string
}

type UpdateAd struct {
	ID        int
	Title     *string
	Status    *models.AdStatus
	ShortDesc *string
	ImageURL  *string
	TargetURL *string
}
