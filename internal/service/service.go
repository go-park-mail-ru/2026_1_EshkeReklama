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
	Update(ctx context.Context, ad *models.Ad) error
	Delete(ctx context.Context, id int) error
}

type AdActionRepository interface{}

type TopicRepository interface{}

type RegionRepository interface{}

type Config struct {
	AdvertiserRepo  AdvertiserRepository
	PartnerRepo     PartnerRepository
	PartnerSiteRepo PartnerSiteRepository
	AdCampaignRepo  AdCampaignRepository
	AdGroupRepo     AdGroupRepository
	AdRepo          AdRepository
	AdActionRepo    AdActionRepository
	TopicRepo       TopicRepository
	RegionRepo      RegionRepository
}

type Service struct {
	advertiserRepo  AdvertiserRepository
	partnerRepo     PartnerRepository
	partnerSiteRepo PartnerSiteRepository
	adCampaignRepo  AdCampaignRepository
	adGroupRepo     AdGroupRepository
	adRepo          AdRepository
	adActionRepo    AdActionRepository
	topicRepo       TopicRepository
	regionRepo      RegionRepository
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
		adActionRepo:    cfg.AdActionRepo,
		topicRepo:       cfg.TopicRepo,
		regionRepo:      cfg.RegionRepo,
	}, nil
}
