package v1

import (
	"context"
	"eshkere/internal/models"
	"eshkere/internal/service"
	serviceinput "eshkere/internal/service/input"

	"github.com/gorilla/mux"
)

type AuthClient interface {
	Register(ctx context.Context, email, phone, password string) (advertiserID int64, sessionID string, expiresAt int64, err error)
	Login(ctx context.Context, identifier, password string) (advertiserID int64, sessionID string, expiresAt int64, err error)
	ValidateSession(ctx context.Context, sessionID string) (advertiserID int64, err error)
	Logout(ctx context.Context, sessionID string) error
	GetCredentials(ctx context.Context, advertiserID int64) (email, phone string, err error)
	UpdateCredentials(ctx context.Context, advertiserID int64, email, phone string) (updatedEmail, updatedPhone string, err error)
}

type Service interface {
	CreateAdvertiserProfile(ctx context.Context, id int64, name, email string) error
	GetAdvertiserByID(ctx context.Context, id int) (*models.Advertiser, error)
	UpdateAdvertiserProfile(ctx context.Context, in *serviceinput.UpdateAdvertiserProfile) (*models.Advertiser, error)
	UpdateAdvertiserAvatar(ctx context.Context, advertiserID int, avatar []byte, avatarExt, avatarContentType string) (*models.Advertiser, error)
	TopUpAdvertiserBalance(ctx context.Context, advertiserID int, amount int64) (int64, error)

	GenerateFeedLink(ctx context.Context, campaignID int) (string, error)
	GetAdByFeedToken(ctx context.Context, token string) (*models.Ad, error)
	RequestAd(ctx context.Context, embedToken string) (*service.AdRequestResult, error)

	CreateAdCampaign(ctx context.Context, in *serviceinput.CreateAdCampaign) (*models.AdCampaign, error)
	UpdateAdCampaign(ctx context.Context, in *serviceinput.UpdateAdCampaign) error
	ListAdCampaigns(ctx context.Context, advertiserID int) ([]*models.AdCampaign, error)
	DeleteAdCampaign(ctx context.Context, campaignID int) error

	CreateAdGroup(ctx context.Context, in *serviceinput.CreateAdGroup) (*models.AdGroup, error)
	UpdateAdGroup(ctx context.Context, in *serviceinput.UpdateAdGroup) error
	ListAdGroups(ctx context.Context, campaignID int) ([]*models.AdGroup, error)
	DeleteAdGroup(ctx context.Context, groupID int) error

	CreateAd(ctx context.Context, in *serviceinput.CreateAd) (*models.Ad, error)
	UpdateAd(ctx context.Context, in *serviceinput.UpdateAd) error
	ListAds(ctx context.Context, groupID int) ([]*models.Ad, error)
	DeleteAd(ctx context.Context, adID int) error

	CreateAppeal(ctx context.Context, in *serviceinput.CreateAppeal) (*models.Appeal, error)
	ListAppeals(ctx context.Context, advertiserID int) ([]*models.Appeal, error)
	GetAppealByID(ctx context.Context, appealID int) (*models.Appeal, error)

	CreatePartnerProfile(ctx context.Context, in *serviceinput.CreatePartnerProfile) error
	GetPartnerByID(ctx context.Context, id int) (*models.Partner, error)
	UpdatePartnerProfile(ctx context.Context, in *serviceinput.UpdatePartnerProfile) (*models.Partner, error)
	CreatePartnerSite(ctx context.Context, in *serviceinput.CreatePartnerSite) (*models.PartnerSite, error)
	GetPartnerSite(ctx context.Context, siteID int) (*models.PartnerSite, error)
	ListPartnerSites(ctx context.Context, partnerID int) ([]*models.PartnerSite, error)
	UpdatePartnerSite(ctx context.Context, in *serviceinput.UpdatePartnerSite) (*models.PartnerSite, error)
	DeletePartnerSite(ctx context.Context, partnerID, siteID int) error
	CreatePartnerBlock(ctx context.Context, partnerID int, in *serviceinput.CreatePartnerBlock) (*models.PartnerBlock, error)
	GetPartnerBlock(ctx context.Context, partnerID, siteID, blockID int) (*models.PartnerBlock, []*models.PartnerBlockGeoRule, error)
	ListPartnerBlocks(ctx context.Context, partnerID, siteID int) ([]*models.PartnerBlock, error)
	UpdatePartnerBlockMeta(ctx context.Context, partnerID, siteID int, in *serviceinput.UpdatePartnerBlockMeta) (*models.PartnerBlock, error)
	UpdatePartnerBlockGeneralSettings(ctx context.Context, partnerID, siteID int, in *serviceinput.UpdatePartnerBlockGeneral) (*models.PartnerBlock, error)
	UpdatePartnerBlockGeographySettings(ctx context.Context, partnerID, siteID int, in *serviceinput.UpdatePartnerBlockGeography) ([]*models.PartnerBlockGeoRule, error)
	UpdatePartnerBlockSelfAdSettings(ctx context.Context, partnerID, siteID int, in *serviceinput.UpdatePartnerBlockSelfAd) (*models.PartnerBlock, error)
	DeletePartnerBlock(ctx context.Context, partnerID, siteID, blockID int) error
	GetPartnerBlockEmbedCode(ctx context.Context, partnerID, siteID, blockID int, baseURL string) (string, string, string, error)
	ListPartnerCountries(ctx context.Context) []service.DictionaryItem
	ListPartnerRegistrationRegions(ctx context.Context, countryCode string) []service.DictionaryItem
	ListPartnerCooperationForms(ctx context.Context) []service.DictionaryItem
	ListPartnerPayoutCurrencies(ctx context.Context) []service.DictionaryItem
	ListPartnerBlockTypes(ctx context.Context) []service.BlockTypeDictionaryItem
	GetPartnerGeoTree(ctx context.Context) []*service.GeoTreeNode
}

type CookieConfig struct {
	Name     string
	Path     string
	HTTPOnly bool
	Secure   bool
}

type APIConfig struct {
	AuthClient          AuthClient
	Service             Service
	CookieConfig        CookieConfig
	PartnerCookieConfig CookieConfig
}

type API struct {
	authClient          AuthClient
	service             Service
	cookieConfig        CookieConfig
	partnerCookieConfig CookieConfig
}

func NewAPI(config APIConfig) *API {
	return &API{
		authClient:          config.AuthClient,
		service:             config.Service,
		cookieConfig:        config.CookieConfig,
		partnerCookieConfig: resolvePartnerCookieConfig(config.CookieConfig, config.PartnerCookieConfig),
	}
}

func resolvePartnerCookieConfig(defaultCfg, partnerCfg CookieConfig) CookieConfig {
	if partnerCfg.Name == "" {
		return defaultCfg
	}
	if partnerCfg.Path == "" {
		partnerCfg.Path = defaultCfg.Path
	}
	return partnerCfg
}

func (a *API) RegisterRoutes(r *mux.Router) {
	a.RegisterAdvertiserHandlers(r)
	// Partner auth/profile HTTP layer is intentionally disabled for now.
	// We keep the underlying service/repository code in place, but do not
	// mount these endpoints until the partner onboarding flow is enabled.
	a.RegisterPartnerDictionaryHandlers(r)
	a.RegisterPartnerSiteHandlers(r)
	a.RegisterPartnerBlockHandlers(r)
	a.RegisterAdRequestHandlers(r)
	a.RegisterAdCampaignHandlers(r)
	a.RegisterAdGroupHandlers(r)
	a.RegisterAdsHandlers(r)
	a.RegisterFeedHandlers(r)
	a.RegisterAppealHandlers(r)
}
