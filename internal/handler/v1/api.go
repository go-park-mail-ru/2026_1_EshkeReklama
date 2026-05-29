package v1

import (
	"context"
	"time"

	authrepo "eshkere/internal/auth/repository/postgres"
	"eshkere/internal/models"
	redisrepo "eshkere/internal/repository/redis"
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
	GetCredentials(ctx context.Context, advertiserID int64) (email, phone string, canChangePassword bool, err error)
	UpdateCredentials(ctx context.Context, advertiserID int64, email, phone string) (updatedEmail, updatedPhone string, err error)
	ChangePassword(ctx context.Context, advertiserID int64, currentPassword, newPassword string) error
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

	GetTariffInfo(ctx context.Context, advertiserID int) (*service.TariffInfo, error)
	PurchaseProSubscription(ctx context.Context, advertiserID int) (*service.TariffInfo, error)
	GenerateAdText(ctx context.Context, advertiserID int, in service.GenerateAdTextInput) (*service.GeneratedAdText, error)
	GenerateAdVariants(ctx context.Context, advertiserID int, in service.GenerateAdVariantsInput) (*service.GeneratedAdVariants, error)
	GenerateAdImage(ctx context.Context, advertiserID int, in service.GenerateAdImageInput) (*service.GeneratedAdImage, error)

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
	ListModerationAds(ctx context.Context) ([]*models.ModerationQueueItem, error)
	ExportCampaignStatsCSV(ctx context.Context, advertiserID, campaignID int, from, to time.Time) ([]byte, error)
	UpdateAd(ctx context.Context, advertiserID int, in *serviceinput.UpdateAd) error
	UpdateAdModerationStatus(ctx context.Context, in *serviceinput.UpdateAdStatus) error
	ListAds(ctx context.Context, advertiserID, groupID int) ([]*models.Ad, error)
	DeleteAd(ctx context.Context, advertiserID, adID int) error
	GetAdStats(ctx context.Context, advertiserID, campaignID, groupID, adID int, from, to time.Time) (*service.AdStats, error)

	GetSupportThreadByCampaign(ctx context.Context, advertiserID, campaignID int) (*models.SupportThread, error)
	GetSupportThread(ctx context.Context, threadID int) (*models.SupportThread, error)
	ListSupportMessages(ctx context.Context, advertiserID, threadID, limit int, beforeID *int) ([]*models.SupportMessage, error)
	CreateSupportMessage(ctx context.Context, advertiserID, threadID int, text string) (*models.SupportMessage, error)
	ListAdminSupportThreads(ctx context.Context) ([]*models.SupportThreadSummary, error)
	ListAdminSupportMessages(ctx context.Context, threadID, limit int, beforeID *int) ([]*models.SupportMessage, error)
	CreateAdminSupportMessage(ctx context.Context, adminID, threadID int, text string) (*models.SupportMessage, error)

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

type VerificationStore interface {
	Save(ctx context.Context, record redisrepo.EmailVerificationRecord, ttl time.Duration) error
	Get(ctx context.Context, email string) (*redisrepo.EmailVerificationRecord, error)
	Delete(ctx context.Context, email string) error
}

type EmailSender interface {
	SendEmailVerificationCode(ctx context.Context, to string, code string) error
	SendPasswordResetCode(ctx context.Context, to string, code string) error
}

type PasswordResetStore interface {
	Save(ctx context.Context, record redisrepo.PasswordResetRecord, ttl time.Duration) error
	Get(ctx context.Context, advertiserID int64) (*redisrepo.PasswordResetRecord, error)
	Delete(ctx context.Context, advertiserID int64) error
}

type CredentialsManager interface {
	FindByIdentifier(ctx context.Context, identifier string) (*authrepo.Credential, error)
	SetPassword(ctx context.Context, id int64, newPassword string) error
}

type APIConfig struct {
	AuthClient              AuthClient
	Service                 Service
	SupportHub              wsSubscriber
	CookieConfig            CookieConfig
	VerificationStore       VerificationStore
	VerificationEmailSender EmailSender
	RegistrationVerifyTTL   time.Duration
	PasswordResetStore      PasswordResetStore
	PasswordResetTTL        time.Duration
	CredentialsManager      CredentialsManager
}

type API struct {
	authClient              AuthClient
	service                 Service
	supportHub              wsSubscriber
	cookieConfig            CookieConfig
	verificationStore       VerificationStore
	verificationEmailSender EmailSender
	registrationVerifyTTL   time.Duration
	passwordResetStore      PasswordResetStore
	passwordResetTTL        time.Duration
	credentialsManager      CredentialsManager
}

const APIPrefix = "/api"

func NewAPI(config APIConfig) *API {
	return &API{
		authClient:              config.AuthClient,
		service:                 config.Service,
		supportHub:              config.SupportHub,
		cookieConfig:            config.CookieConfig,
		verificationStore:       config.VerificationStore,
		verificationEmailSender: config.VerificationEmailSender,
		registrationVerifyTTL:   config.RegistrationVerifyTTL,
		passwordResetStore:      config.PasswordResetStore,
		passwordResetTTL:        config.PasswordResetTTL,
		credentialsManager:      config.CredentialsManager,
	}
}

func (a *API) RegisterRoutes(r *mux.Router) {
	a.RegisterPaymentHandlers(r)
	a.RegisterAdvertiserHandlers(r)
	a.RegisterSubscriptionHandlers(r)
	a.RegisterAIHandlers(r)
	a.RegisterPartnerHandlers(r)
	a.RegisterPartnerDictionaryHandlers(r)
	a.RegisterPartnerSiteHandlers(r)
	a.RegisterPartnerStatsHandlers(r)
	a.RegisterPartnerBlockHandlers(r)
	a.RegisterAdRequestHandlers(r)
	a.RegisterAdvertiserStatsHandlers(r)
	a.RegisterAdCampaignHandlers(r)
	a.RegisterAdGroupHandlers(r)
	a.RegisterAdsHandlers(r)
	a.RegisterAdminHandlers(r)
	a.RegisterFeedHandlers(r)
	a.RegisterAppealHandlers(r)
	a.RegisterSupportHandlers(r)
	a.RegisterSupportWSHandler(r)
}
