package service

import (
	"context"
	"encoding/json"
	"errors"
	"eshkere/internal/models"
)

type AdvertiserRepository interface {
	CreateProfile(ctx context.Context, id int64, name string) error
	GetByID(ctx context.Context, id int) (*models.Advertiser, error)
	Update(ctx context.Context, a *models.Advertiser) error
}

type PartnerRepository interface {
	CreateProfile(ctx context.Context, partner *models.Partner) error
	GetByID(ctx context.Context, id int) (*models.Partner, error)
	Update(ctx context.Context, partner *models.Partner) error
}

type PartnerSiteRepository interface {
	Create(ctx context.Context, site *models.PartnerSite) error
	GetByID(ctx context.Context, siteID int) (*models.PartnerSite, error)
	ListByPartnerID(ctx context.Context, partnerID int) ([]*models.PartnerSite, error)
	Update(ctx context.Context, site *models.PartnerSite) error
	Delete(ctx context.Context, siteID int) error
	ExistsByDomain(ctx context.Context, domain string) (bool, error)
}

type PartnerBlockRepository interface {
	Create(ctx context.Context, block *models.PartnerBlock) error
	GetByID(ctx context.Context, blockID int) (*models.PartnerBlock, error)
	ListBySiteID(ctx context.Context, siteID int) ([]*models.PartnerBlock, error)
	Update(ctx context.Context, block *models.PartnerBlock) error
	Delete(ctx context.Context, blockID int) error
	GetByEmbedToken(ctx context.Context, token string) (*models.PartnerBlock, error)
}

type PartnerBlockGeoRuleRepository interface {
	ListByBlockID(ctx context.Context, blockID int) ([]*models.PartnerBlockGeoRule, error)
	ReplaceByBlockID(ctx context.Context, blockID int, rules []*models.PartnerBlockGeoRule) error
}

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
	GetRandomWorking(ctx context.Context) (*models.Ad, error)
	UpdateImage(ctx context.Context, adID int, imageKey string) error
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

type AppealStorage interface {
	UploadAppealImage(ctx context.Context, appealID int, data []byte, ext string, contentType string) (string, error)
	DeleteAppealImage(ctx context.Context, appealID int, imageKey string) error
	GetAppealImageURL(imageKey string) string
}

type AdStorage interface {
	UploadAdImage(ctx context.Context, data []byte, ext string, contentType string) (string, error)
	DeleteAdImage(ctx context.Context, imageKey string) error
	GetAdImageURL(imagKey string) string
}

type AdActionRepository interface{}

type TopicRepository interface{}

type RegionRepository interface{}

type AppealRepository interface {
	Create(ctx context.Context, appeal *models.Appeal) error
	GetByID(ctx context.Context, appealID int) (*models.Appeal, error)
	ListByAdvertiserID(ctx context.Context, advertiserID int) ([]*models.Appeal, error)
	UpdateImage(ctx context.Context, appealID int, imageKey string) error
}

type Config struct {
	AdvertiserRepo          AdvertiserRepository
	PartnerRepo             PartnerRepository
	PartnerSiteRepo         PartnerSiteRepository
	PartnerBlockRepo        PartnerBlockRepository
	PartnerBlockGeoRuleRepo PartnerBlockGeoRuleRepository
	AdCampaignRepo          AdCampaignRepository
	AdGroupRepo             AdGroupRepository
	AdRepo                  AdRepository
	FeedLinkRepo            FeedLinkRepository
	AvatarStorage           AvatarStorage
	AppealStorage           AppealStorage
	AdStorage               AdStorage
	AdActionRepo            AdActionRepository
	TopicRepo               TopicRepository
	RegionRepo              RegionRepository
	AppealRepo              AppealRepository
}

type Service struct {
	advertiserRepo          AdvertiserRepository
	partnerRepo             PartnerRepository
	partnerSiteRepo         PartnerSiteRepository
	partnerBlockRepo        PartnerBlockRepository
	partnerBlockGeoRuleRepo PartnerBlockGeoRuleRepository
	adCampaignRepo          AdCampaignRepository
	adGroupRepo             AdGroupRepository
	adRepo                  AdRepository
	feedLinkRepo            FeedLinkRepository
	avatarStorage           AvatarStorage
	appealStorage           AppealStorage
	adStorage               AdStorage
	adActionRepo            AdActionRepository
	topicRepo               TopicRepository
	regionRepo              RegionRepository
	appealRepo              AppealRepository
}

func NewService(cfg *Config) (*Service, error) {
	if cfg == nil {
		return nil, errors.New("service config is nil")
	}

	return &Service{
		advertiserRepo:          cfg.AdvertiserRepo,
		partnerRepo:             cfg.PartnerRepo,
		partnerSiteRepo:         cfg.PartnerSiteRepo,
		partnerBlockRepo:        cfg.PartnerBlockRepo,
		partnerBlockGeoRuleRepo: cfg.PartnerBlockGeoRuleRepo,
		adCampaignRepo:          cfg.AdCampaignRepo,
		adGroupRepo:             cfg.AdGroupRepo,
		adRepo:                  cfg.AdRepo,
		feedLinkRepo:            cfg.FeedLinkRepo,
		avatarStorage:           cfg.AvatarStorage,
		appealStorage:           cfg.AppealStorage,
		adStorage:               cfg.AdStorage,
		adActionRepo:            cfg.AdActionRepo,
		topicRepo:               cfg.TopicRepo,
		regionRepo:              cfg.RegionRepo,
		appealRepo:              cfg.AppealRepo,
	}, nil
}

type DictionaryItem struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type BlockTypeDictionaryItem struct {
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Platforms   []string `json:"platforms"`
}

type GeoTreeNode struct {
	Code     string         `json:"code"`
	Name     string         `json:"name"`
	Children []*GeoTreeNode `json:"children"`
}

type PartnerBlockSettings struct {
	OnlyConfigured bool                          `json:"only_configured"`
	GlobalCPMV     *int64                        `json:"global_cpmv"`
	Rules          []*models.PartnerBlockGeoRule `json:"rules"`
}

var defaultSelfAdSettings = json.RawMessage(`{"reserved":true}`)
