package v1

import (
	"context"
	"eshkere/internal/models"
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
	a.RegisterAdvertiserHandlers(r)
	a.RegisterAdCampaignHandlers(r)
	a.RegisterAdGroupHandlers(r)
	a.RegisterAdsHandlers(r)
	a.RegisterFeedHandlers(r)
	a.RegisterAppealHandlers(r)
}
