package service

import (
	"errors"
)

type AdvertiserRepository interface{}

type PartnerRepository interface{}

type PartnerSiteRepository interface{}

type AdCampaignRepository interface{}

type AdGroupRepository interface{}

type AdRepository interface{}

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
