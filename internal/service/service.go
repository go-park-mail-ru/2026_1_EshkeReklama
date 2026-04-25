package service

import (
	"context"
	"errors"
	"eshkere/internal/models"
)

type AdvertiserRepository interface {
	Create(ctx context.Context, a *models.Advertiser) (int, error)
	GetByID(ctx context.Context, id int) (*models.Advertiser, error)
	GetByEmail(ctx context.Context, email string) (*models.Advertiser, error)
	GetByPhone(ctx context.Context, phone string) (*models.Advertiser, error)
	Update(ctx context.Context, a *models.Advertiser) error
}

type PartnerRepository interface{}

type PartnerSiteRepository interface{}

type AdCampaignRepository interface {
	Create(ctx context.Context, c *models.AdCampaign) error
	GetByID(ctx context.Context, id int) (*models.AdCampaign, error)
	ListByAdvertiserID(ctx context.Context, advertiserID int) ([]*models.AdCampaign, error)
	Update(ctx context.Context, c *models.AdCampaign) error
	Delete(ctx context.Context, id int) error
}

type AdGroupRepository interface {
	Create(ctx context.Context, g *models.AdGroup) error
	GetByID(ctx context.Context, id int) (*models.AdGroup, error)
	ListByCampaignID(ctx context.Context, campaignID int) ([]*models.AdGroup, error)
	Update(ctx context.Context, g *models.AdGroup) error
	Delete(ctx context.Context, id int) error
}

type AdRepository interface {
	Create(ctx context.Context, ad *models.Ad) error
	GetByID(ctx context.Context, adID int) (*models.Ad, error)
	ListByAdGroupID(ctx context.Context, adGroupID int) ([]*models.Ad, error)
	ListByAdCampaignID(ctx context.Context, campaignID int) ([]*models.Ad, error)
	Update(ctx context.Context, ad *models.Ad) error
	Delete(ctx context.Context, id int) error
}

type FeedLinkRepository interface {
	Create(ctx context.Context, advertiserID int, token string) error
	GetCampaignIDByToken(ctx context.Context, token string) (int, error)
}

type AvatarStorage interface {
	UploadAvatar(ctx context.Context, advertiserID int, data []byte, ext string, contentType string) (string, error)
	DeleteAvatar(ctx context.Context, advertiserID int, avatarKey string) error
	GetAvatarURL(avatarKey string) string
}

type AdActionRepository interface{}

type TopicRepository interface{}

type RegionRepository interface{}

type AppealRepository interface {
	Create(ctx context.Context, appeal *models.Appeal) error
	GetByID(ctx context.Context, appealID int) (*models.Appeal, error)
	ListByAdvertiserID(ctx context.Context, advertiserID int) ([]*models.Appeal, error)
}

type Config struct {
	AdvertiserRepo  AdvertiserRepository
	PartnerRepo     PartnerRepository
	PartnerSiteRepo PartnerSiteRepository
	AdCampaignRepo  AdCampaignRepository
	AdGroupRepo     AdGroupRepository
	AdRepo          AdRepository
	FeedLinkRepo    FeedLinkRepository
	AvatarStorage   AvatarStorage
	AdActionRepo    AdActionRepository
	TopicRepo       TopicRepository
	RegionRepo      RegionRepository
	AppealRepo      AppealRepository
}

type Service struct {
	advertiserRepo  AdvertiserRepository
	partnerRepo     PartnerRepository
	partnerSiteRepo PartnerSiteRepository
	adCampaignRepo  AdCampaignRepository
	adGroupRepo     AdGroupRepository
	adRepo          AdRepository
	feedLinkRepo    FeedLinkRepository
	avatarStorage   AvatarStorage
	adActionRepo    AdActionRepository
	topicRepo       TopicRepository
	regionRepo      RegionRepository
	appealRepo      AppealRepository
}

func NewService(cfg *Config) (*Service, error) {
	if cfg == nil {
		return nil, errors.New("service config is nil")
	}

	return &Service{
		advertiserRepo:  cfg.AdvertiserRepo,
		partnerRepo:     cfg.PartnerRepo,
		partnerSiteRepo: cfg.PartnerSiteRepo,
		adCampaignRepo:  cfg.AdCampaignRepo,
		adGroupRepo:     cfg.AdGroupRepo,
		adRepo:          cfg.AdRepo,
		feedLinkRepo:    cfg.FeedLinkRepo,
		avatarStorage:   cfg.AvatarStorage,
		adActionRepo:    cfg.AdActionRepo,
		topicRepo:       cfg.TopicRepo,
		regionRepo:      cfg.RegionRepo,
		appealRepo:      cfg.AppealRepo,
	}, nil
}
