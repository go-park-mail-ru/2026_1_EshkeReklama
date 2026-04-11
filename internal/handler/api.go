package handlers

import (
	"context"
	"eshkere/internal/handler/dto"
	"eshkere/internal/models"
	"eshkere/internal/session"

	"github.com/gorilla/mux"
)

type Service interface {
	CreateAd(ctx context.Context, ad *models.Ad) (*models.Ad, error)
	UpdateAd(ctx context.Context, adID int, req dto.UpdateAdRequest) error
	ListAds(ctx context.Context, groupID int) ([]*models.Ad, error)
	DeleteAd(ctx context.Context, adID int) error

	CreateAdCampaign(ctx context.Context, c *models.AdCampaign) (*models.AdCampaign, error)
	UpdateAdCampaign(ctx context.Context, campaignID int, req dto.UpdateAdCampaignRequest) error
	ListAdCampaigns(ctx context.Context, advertiserID int) ([]*models.AdCampaign, error)
	DeleteAdCampaign(ctx context.Context, campaignID int) error

	CreateAdGroup(ctx context.Context, g *models.AdGroup) (*models.AdGroup, error)
	UpdateAdGroup(ctx context.Context, groupID int, req dto.UpdateAdGroupRequest) error
	ListAdGroups(ctx context.Context, campaignID int) ([]*models.AdGroup, error)
	DeleteAdGroup(ctx context.Context, groupID int) error
}

type APIConfig struct {
	SessionManager *session.Manager
	Service        Service
}

type API struct {
	sessionManager *session.Manager
	service        Service
}

func NewAPI(config APIConfig) *API {
	return &API{
		sessionManager: config.SessionManager,
		service:        config.Service,
	}
}

func (a *API) RegisterRoutes(r *mux.Router) {
	a.RegisterAdvertiserHandlers(r)
	a.RegisterAdCampaignHandlers(r)
	a.RegisterAdGroupHandlers(r)
	a.RegisterAdsHandlers(r)
}
