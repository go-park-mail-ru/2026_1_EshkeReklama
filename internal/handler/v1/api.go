package v1

import (
	"context"
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
	"eshkere/internal/session"

	"github.com/gorilla/mux"
)

type Service interface {
	RegisterAdvertiser(ctx context.Context, name, email, phone, password string) (*models.Advertiser, error)
	AuthenticateAdvertiser(ctx context.Context, identifier, password string) (*models.Advertiser, error)
	GetAdvertiserByID(ctx context.Context, id int) (*models.Advertiser, error)
	UpdateAdvertiserProfile(ctx context.Context, in *serviceinput.UpdateAdvertiserProfile) (*models.Advertiser, error)
	UpdateAdvertiserAvatar(ctx context.Context, advertiserID int, avatar []byte, avatarExt, avatarContentType string) (*models.Advertiser, error)
	TopUpAdvertiserBalance(ctx context.Context, advertiserID int, amount int64) (int64, error)

	GenerateFeedLink(ctx context.Context, campaignID int) (string, error)
	GetAdsByFeedToken(ctx context.Context, token string) ([]*models.Ad, error)

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

	AdminListAppeals(ctx context.Context, filter *serviceinput.AdminListAppealsFilter) ([]*models.Appeal, error)
	AdminGetAppealWithHistory(ctx context.Context, appealID int) (*models.Appeal, []*models.AppealMessage, []*models.AppealStatusHistory, error)
	AdminPatchAppealStatus(ctx context.Context, in *serviceinput.AdminPatchAppealStatus) error
	AdminPostAppealMessage(ctx context.Context, in *serviceinput.AdminPostAppealMessage) (*models.AppealMessage, error)
}

type APIConfig struct {
	SessionManager *session.Manager
	Service        Service
	AdminToken     string
}

type API struct {
	sessionManager *session.Manager
	service        Service
	adminToken     string
}

func NewAPI(config APIConfig) *API {
	return &API{
		sessionManager: config.SessionManager,
		service:        config.Service,
		adminToken:     config.AdminToken,
	}
}

func (a *API) RegisterRoutes(r *mux.Router) {
	a.RegisterAdvertiserHandlers(r)
	a.RegisterAdCampaignHandlers(r)
	a.RegisterAdGroupHandlers(r)
	a.RegisterAdsHandlers(r)
	a.RegisterFeedHandlers(r)
	a.RegisterAppealHandlers(r)
	a.RegisterAdminAppealHandlers(r)
}
