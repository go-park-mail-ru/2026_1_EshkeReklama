package input

import "eshkere/internal/models"

type CreateAd struct {
	AdGroupID int
	Title     string
	ShortDesc string
	Image     []byte
	ImageExt  string
	ImageType string
	TargetURL string
}

type UpdateAd struct {
	ID        int
	Title     *string
	Status    *models.AdStatus
	ShortDesc *string
	Image     *[]byte
	ImageExt  *string
	ImageType *string
	TargetURL *string
}
