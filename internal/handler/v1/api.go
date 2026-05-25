package v1

import (
	"context"
	"time"

	"eshkere/internal/models"
	"eshkere/internal/service"
	serviceinput "eshkere/internal/service/input"

	"github.com/gorilla/mux"
)

type AuthClient interface {
	Register(ctx context.Context, email, phone, password string) (advertiserID int64, sessionID string, expiresAt int64, err error)
	Login(ctx context.Context, identifier, password string) (advertiserID int64, sessionID string, expiresAt int64, err error)
	LoginVKID(ctx context.Context, accessToken string, userID int64) (advertiserID int64, sessionID string, expiresAt int64, firstName string, lastName string, err error)
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
	CreateBalancePayment(ctx context.Context, advertiserID int, amount int64) (*service.BalancePaymentResult, error)
	CompletePaymentByWebhook(ctx context.Context, paymentID string) (*service.WebhookResult, error)
	GetAdvertiserAutopaySettings(ctx context.Context, advertiserID int) (*models.AdvertiserAutopaySettings, error)
	UpdateAdvertiserAutopaySettings(ctx context.Context, settings *models.AdvertiserAutopaySettings) error
	GetAdvertiserNotificationSettings(ctx context.Context, advertiserID int) (*models.AdvertiserNotificationSettings, error)
	UpdateAdvertiserNotificationSettings(ctx context.Context, settings *models.AdvertiserNotificationSettings) error
	RunAutopayCycle(ctx context.Context) (int, error)

	GenerateFeedLink(ctx context.Context, campaignID int) (string, error)
	GetAdByFeedToken(ctx context.Context, token string) (*models.Ad, error)
	RequestAd(ctx context.Context, embedToken, visitorID string) (*service.AdRequestResult, error)
	ClickAd(ctx context.Context, requestID string) (string, error)

	CreateAdCampaign(ctx context.Context, in *serviceinput.CreateAdCampaign) (*models.AdCampaign, error)
	UpdateAdCampaign(ctx context.Context, advertiserID int, in *serviceinput.UpdateAdCampaign) error
	TurnOffAdCampaign(ctx context.Context, advertiserID, campaignID int) error
	ListAdCampaigns(ctx context.Context, advertiserID int) ([]*models.AdCampaign, error)
	DeleteAdCampaign(ctx context.Context, advertiserID, campaignID int) error
	GetCampaignStats(ctx context.Context, advertiserID, campaignID int, from, to time.Time) (*service.CampaignStats, error)

	CreateAdGroup(ctx context.Context, advertiserID int, in *serviceinput.CreateAdGroup) (*models.AdGroup, error)
	UpdateAdGroup(ctx context.Context, advertiserID int, in *serviceinput.UpdateAdGroup) error
	ListAdGroups(ctx context.Context, advertiserID, campaignID int) ([]*models.AdGroup, error)
	DeleteAdGroup(ctx context.Context, advertiserID, groupID int) error
	GetGroupStats(ctx context.Context, advertiserID, campaignID, groupID int, from, to time.Time) (*service.GroupStats, error)

	CreateAd(ctx context.Context, advertiserID int, in *serviceinput.CreateAd) (*models.Ad, error)
	GetAdByID(ctx context.Context, adID int) (*models.Ad, error)
	ListModerationAds(ctx context.Context) ([]*models.Ad, error)
	UpdateAd(ctx context.Context, advertiserID int, in *serviceinput.UpdateAd) error
	UpdateAdModerationStatus(ctx context.Context, in *serviceinput.UpdateAdStatus) error
	ListAds(ctx context.Context, advertiserID, groupID int) ([]*models.Ad, error)
	DeleteAd(ctx context.Context, advertiserID, adID int) error
	GetAdStats(ctx context.Context, advertiserID, campaignID, groupID, adID int, from, to time.Time) (*service.AdStats, error)

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
	GetPartnerBlockEmbedCode(ctx context.Context, partnerID, siteID, blockID int, baseURL string) (string, string, error)
	ListPartnerCountries(ctx context.Context) []service.DictionaryItem
	ListPartnerRegistrationRegions(ctx context.Context, countryCode string) []service.DictionaryItem
	ListPartnerCooperationForms(ctx context.Context) []service.DictionaryItem
	ListPartnerPayoutCurrencies(ctx context.Context) []service.DictionaryItem
	ListPartnerBlockTypes(ctx context.Context) []service.BlockTypeDictionaryItem
	GetPartnerGeoTree(ctx context.Context) []*service.GeoTreeNode
	GetPartnerIncomeStats(ctx context.Context, partnerID int, from, to time.Time) (*service.PartnerIncomeStats, error)
}

type CookieConfig struct {
	Name     string
	Path     string
	HTTPOnly bool
	Secure   bool
}

type APIConfig struct {
	AuthClient   AuthClient
	Service      Service
	CookieConfig CookieConfig
}

type API struct {
	authClient   AuthClient
	service      Service
	cookieConfig CookieConfig
}

func NewAPI(config APIConfig) *API {
	return &API{
		authClient:   config.AuthClient,
		service:      config.Service,
		cookieConfig: config.CookieConfig,
	}
}

func (a *API) RegisterRoutes(r *mux.Router) {
	a.RegisterPaymentHandlers(r)
	a.RegisterAdvertiserHandlers(r)
	a.RegisterPartnerDictionaryHandlers(r)
	a.RegisterPartnerSiteHandlers(r)
	a.RegisterPartnerIncomeHandlers(r)
	a.RegisterPartnerBlockHandlers(r)
	a.RegisterAdRequestHandlers(r)
	a.RegisterStatsHandlers(r)
	a.RegisterAdCampaignHandlers(r)
	a.RegisterAdGroupHandlers(r)
	a.RegisterAdsHandlers(r)
	a.RegisterAdminHandlers(r)
	a.RegisterFeedHandlers(r)
	a.RegisterAppealHandlers(r)
}
