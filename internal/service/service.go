package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"eshkere/internal/analytics"
	"eshkere/internal/models"
	"eshkere/internal/yookassa"
)

type AdvertiserRepository interface {
	CreateProfile(ctx context.Context, id int64, name string) error
	GetByID(ctx context.Context, id int) (*models.Advertiser, error)
	Update(ctx context.Context, a *models.Advertiser) error
}

type PaymentTransactionRepository interface {
	Create(ctx context.Context, tx *models.PaymentTransaction) error
	GetByID(ctx context.Context, id string) (*models.PaymentTransaction, error)
	Complete(ctx context.Context, paymentID string, status models.PaymentTransactionStatus, paymentMethodID string, paymentMethodTitle string) (*models.PaymentCompletionResult, error)
}

type AdvertiserAutopaySettingsRepository interface {
	GetByAdvertiserID(ctx context.Context, advertiserID int) (*models.AdvertiserAutopaySettings, error)
	Upsert(ctx context.Context, settings *models.AdvertiserAutopaySettings) error
	ListEligible(ctx context.Context) ([]models.AdvertiserAutopayCandidate, error)
}

type AdvertiserNotificationSettingsRepository interface {
	GetByAdvertiserID(ctx context.Context, advertiserID int) (*models.AdvertiserNotificationSettings, error)
	Upsert(ctx context.Context, settings *models.AdvertiserNotificationSettings) error
	ListEnabled(ctx context.Context) ([]models.AdvertiserNotificationSettings, error)
}

type YookassaClient interface {
	Enabled() bool
	CreateRedirectPayment(ctx context.Context, amountRub int64, description string, metadata map[string]string) (*yookassa.Payment, error)
	CreateAutopayPayment(ctx context.Context, amountRub int64, paymentMethodID string, description string, metadata map[string]string) (*yookassa.Payment, error)
	GetPayment(ctx context.Context, paymentID string) (*yookassa.Payment, error)
}

type PartnerRepository interface {
	CreateProfile(ctx context.Context, partner *models.Partner) error
	GetByID(ctx context.Context, id int) (*models.Partner, error)
	Update(ctx context.Context, partner *models.Partner) error
	SettleDailyEarnings(ctx context.Context, earningDate time.Time) (int64, error)
}

type PartnerIncomeRepository interface {
	ListPartnerIncome(ctx context.Context, partnerID int, from, to time.Time) ([]PartnerIncomeRow, error)
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
	ListByStatus(ctx context.Context, status models.AdStatus) ([]*models.Ad, error)
	UpdateStatusByCampaignID(ctx context.Context, campaignID int, status models.AdStatus) error
	GetRandomWorking(ctx context.Context) (*models.Ad, error)
	ListAdCandidates(ctx context.Context, spendDate time.Time) ([]*models.AdCandidate, error)
	ReserveImpression(ctx context.Context, reservation models.ImpressionReservation) (bool, error)
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

type TopicScore struct {
	TopicID int
	Score   float64
}

type ProfileClient interface {
	GetProfile(ctx context.Context, visitorID string) ([]TopicScore, bool, error)
	TrackEvent(ctx context.Context, visitorID string, topicID int, eventType string) error
}

type AdEventPublisher interface {
	PublishAdEvent(ctx context.Context, event analytics.AdEvent) error
}

type StatsFilter struct {
	CampaignID int
	AdGroupID  int
	AdID       int
	From       time.Time
	To         time.Time
}

type StatsTotals struct {
	Impressions     int64
	Clicks          int64
	Spend           int64
	PartnerReward   int64
	PlatformRevenue int64
}

type StatsTimelinePoint struct {
	Date   time.Time
	Totals StatsTotals
}

type StatsBreakdownRow struct {
	ID     int
	Totals StatsTotals
}

type PartnerIncomeRow struct {
	Date        time.Time
	SiteID      int
	SiteName    string
	Domain      string
	BlockID     int
	BlockName   string
	Impressions int64
	Reward      int64
}

type PartnerIncomeStats struct {
	From        time.Time
	To          time.Time
	Impressions int64
	Reward      int64
	ECPM        float64
	Rows        []PartnerIncomeRow
}

type StatsReader interface {
	Totals(ctx context.Context, filter StatsFilter) (StatsTotals, error)
	Timeline(ctx context.Context, filter StatsFilter) ([]StatsTimelinePoint, error)
	Breakdown(ctx context.Context, filter StatsFilter, dimension string) ([]StatsBreakdownRow, error)
}

type AdRequestRecord struct {
	RequestID       string
	VisitorID       string
	AdvertiserID    int
	CampaignID      int
	AdGroupID       int
	AdID            int
	PartnerBlockID  int
	PartnerSiteID   int
	TopicID         int
	TargetURL       string
	Price           int64
	PartnerReward   int64
	PlatformRevenue int64
}

type AdRequestStore interface {
	Save(ctx context.Context, record AdRequestRecord, ttl time.Duration) error
	Get(ctx context.Context, requestID string) (*AdRequestRecord, error)
	MarkClickedOnce(ctx context.Context, requestID string, ttl time.Duration) (bool, error)
}

type Config struct {
	AdvertiserRepo           AdvertiserRepository
	PaymentTransactionRepo   PaymentTransactionRepository
	AutopaySettingsRepo      AdvertiserAutopaySettingsRepository
	NotificationSettingsRepo AdvertiserNotificationSettingsRepository
	PartnerRepo              PartnerRepository
	PartnerSiteRepo          PartnerSiteRepository
	PartnerBlockRepo         PartnerBlockRepository
	PartnerBlockGeoRuleRepo  PartnerBlockGeoRuleRepository
	PartnerIncomeRepo        PartnerIncomeRepository
	AdCampaignRepo           AdCampaignRepository
	AdGroupRepo              AdGroupRepository
	AdRepo                   AdRepository
	FeedLinkRepo             FeedLinkRepository
	AvatarStorage            AvatarStorage
	AppealStorage            AppealStorage
	AdStorage                AdStorage
	AdActionRepo             AdActionRepository
	TopicRepo                TopicRepository
	RegionRepo               RegionRepository
	AppealRepo               AppealRepository
	ProfileClient            ProfileClient
	AdRequestStore           AdRequestStore
	AdEventPublisher         AdEventPublisher
	YookassaClient           YookassaClient
	StatsReader              StatsReader
}

type Service struct {
	advertiserRepo           AdvertiserRepository
	paymentTransactionRepo   PaymentTransactionRepository
	autopaySettingsRepo      AdvertiserAutopaySettingsRepository
	notificationSettingsRepo AdvertiserNotificationSettingsRepository
	partnerRepo              PartnerRepository
	partnerSiteRepo          PartnerSiteRepository
	partnerBlockRepo         PartnerBlockRepository
	partnerBlockGeoRuleRepo  PartnerBlockGeoRuleRepository
	partnerIncomeRepo        PartnerIncomeRepository
	adCampaignRepo           AdCampaignRepository
	adGroupRepo              AdGroupRepository
	adRepo                   AdRepository
	feedLinkRepo             FeedLinkRepository
	avatarStorage            AvatarStorage
	appealStorage            AppealStorage
	adStorage                AdStorage
	adActionRepo             AdActionRepository
	topicRepo                TopicRepository
	regionRepo               RegionRepository
	appealRepo               AppealRepository
	profileClient            ProfileClient
	adRequestStore           AdRequestStore
	adEventPublisher         AdEventPublisher
	yookassaClient           YookassaClient
	statsReader              StatsReader
}

func NewService(cfg *Config) (*Service, error) {
	if cfg == nil {
		return nil, errors.New("service config is nil")
	}

	return &Service{
		advertiserRepo:           cfg.AdvertiserRepo,
		paymentTransactionRepo:   cfg.PaymentTransactionRepo,
		autopaySettingsRepo:      cfg.AutopaySettingsRepo,
		notificationSettingsRepo: cfg.NotificationSettingsRepo,
		partnerRepo:              cfg.PartnerRepo,
		partnerSiteRepo:          cfg.PartnerSiteRepo,
		partnerBlockRepo:         cfg.PartnerBlockRepo,
		partnerBlockGeoRuleRepo:  cfg.PartnerBlockGeoRuleRepo,
		partnerIncomeRepo:        cfg.PartnerIncomeRepo,
		adCampaignRepo:           cfg.AdCampaignRepo,
		adGroupRepo:              cfg.AdGroupRepo,
		adRepo:                   cfg.AdRepo,
		feedLinkRepo:             cfg.FeedLinkRepo,
		avatarStorage:            cfg.AvatarStorage,
		appealStorage:            cfg.AppealStorage,
		adStorage:                cfg.AdStorage,
		adActionRepo:             cfg.AdActionRepo,
		topicRepo:                cfg.TopicRepo,
		regionRepo:               cfg.RegionRepo,
		appealRepo:               cfg.AppealRepo,
		profileClient:            cfg.ProfileClient,
		adRequestStore:           cfg.AdRequestStore,
		adEventPublisher:         cfg.AdEventPublisher,
		yookassaClient:           cfg.YookassaClient,
		statsReader:              cfg.StatsReader,
	}, nil
}

func (s *Service) SetProfileClient(client ProfileClient) {
	s.profileClient = client
}

type DictionaryItem struct {
	Code string
	Name string
}

type BlockTypeDictionaryItem struct {
	Code        string
	Name        string
	Description string
	Platforms   []string
}

type GeoTreeNode struct {
	Code     string
	Name     string
	Children []*GeoTreeNode
}

var defaultSelfAdSettings = json.RawMessage(`{"reserved":true}`)
