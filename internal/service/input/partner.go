package input

import "eshkere/internal/models"

type CreatePartnerProfile struct {
	ID                     int64
	LastName               string
	FirstName              string
	MiddleName             string
	BirthDate              string
	CountryCode            string
	RegistrationRegionCode string
	CooperationForm        string
	PayoutCurrency         string
}

type UpdatePartnerProfile struct {
	PartnerID              int
	LastName               *string
	FirstName              *string
	MiddleName             *string
	BirthDate              *string
	CountryCode            *string
	RegistrationRegionCode *string
	CooperationForm        *string
	PayoutCurrency         *string
}

type CreatePartnerSite struct {
	PartnerID int
	Domain    string
	SiteName  string
}

type UpdatePartnerSite struct {
	ID        int
	PartnerID int
	Domain    *string
	SiteName  *string
	Status    *models.PartnerSiteStatus
}

type CreatePartnerBlock struct {
	PartnerSiteID int
	Name          string
	BlockType     models.PartnerBlockType
}

type UpdatePartnerBlockMeta struct {
	ID     int
	Name   *string
	Status *models.PartnerBlockStatus
}

type UpdatePartnerBlockGeneral struct {
	ID                           int
	CPMStrategy                  *models.CPMStrategy
	AmpMode                      *models.AmpMode
	SizeMode                     *models.SizeMode
	BorderMode                   *models.BorderMode
	CornerMode                   *models.CornerMode
	Theme                        *models.ThemeMode
	InterscrollerMode            *models.InterscrollerMode
	InterscrollerBackgroundColor *string
	RevenueShareBPS              *int
}

type UpdatePartnerBlockGeography struct {
	ID             int
	OnlyConfigured bool
	GlobalCPMV     *int64
	Rules          []PartnerBlockGeoRuleInput
}

type PartnerBlockGeoRuleInput struct {
	GeoCode   string
	IsEnabled bool
	CPMV      *int64
}

type UpdatePartnerBlockSelfAd struct {
	ID       int
	Settings []byte
}
