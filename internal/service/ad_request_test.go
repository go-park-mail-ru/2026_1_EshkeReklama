package service

import (
	"context"
	"testing"
	"time"

	"eshkere/internal/analytics"
	"eshkere/internal/models"

	"go.uber.org/mock/gomock"
)

func TestImpressionPrice(t *testing.T) {
	tests := []struct {
		name string
		cpm  int64
		want int64
	}{
		{name: "zero", cpm: 0, want: 0},
		{name: "one kopek cpm", cpm: 1, want: 1},
		{name: "exact division", cpm: 20000, want: 20},
		{name: "round up", cpm: 20001, want: 21},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ImpressionPrice(tt.cpm); got != tt.want {
				t.Fatalf("ImpressionPrice(%d)=%d want %d", tt.cpm, got, tt.want)
			}
		})
	}
}

func TestPersonalizedWeight(t *testing.T) {
	if got := personalizedWeight(100, 0); got != 100 {
		t.Fatalf("score 0 weight=%d want 100", got)
	}
	if got := personalizedWeight(100, 10); got != 200 {
		t.Fatalf("score 10 weight=%d want 200", got)
	}
	if got := personalizedWeight(100, 50); got != 300 {
		t.Fatalf("capped score weight=%d want 300", got)
	}
}

func TestAdSelectionWeight(t *testing.T) {
	if got := adSelectionWeight(0, 0); got != 100 {
		t.Fatalf("score 0 ad weight=%d want 100", got)
	}
	if got := adSelectionWeight(5, 0); got != 150 {
		t.Fatalf("score 5 ad weight=%d want 150", got)
	}
	if got := adSelectionWeight(0, 5); got != 150 {
		t.Fatalf("region score 5 ad weight=%d want 150", got)
	}
	if got := adSelectionWeight(5, 5); got != 225 {
		t.Fatalf("topic and region score 5 ad weight=%d want 225", got)
	}
	if got := adSelectionWeight(50, 50); got != 900 {
		t.Fatalf("capped topic and region score ad weight=%d want 900", got)
	}
}

func TestGroupEligibleCandidates(t *testing.T) {
	candidates := groupEligibleCandidates([]*models.AdCandidate{
		{
			CampaignID:        1,
			AdvertiserID:      10,
			DailyBudget:       1000,
			CPMPrice:          20000,
			SpentToday:        100,
			AdvertiserBalance: 1000,
			Ad:                &models.Ad{ID: 1},
		},
		{
			CampaignID:        1,
			AdvertiserID:      10,
			DailyBudget:       1000,
			CPMPrice:          20000,
			SpentToday:        100,
			AdvertiserBalance: 1000,
			Ad:                &models.Ad{ID: 2},
		},
		{
			CampaignID:        2,
			AdvertiserID:      11,
			DailyBudget:       10,
			CPMPrice:          20000,
			SpentToday:        0,
			AdvertiserBalance: 1000,
			Ad:                &models.Ad{ID: 3},
		},
	}, nil, nil)

	if len(candidates) != 1 {
		t.Fatalf("expected one eligible campaign, got %d", len(candidates))
	}
	if candidates[0].campaignID != 1 || candidates[0].weight != 900 || len(candidates[0].ads) != 2 {
		t.Fatalf("unexpected grouped candidate: %+v", candidates[0])
	}
}

func TestRequestAdReservesImpression(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	blockRepo := NewMockPartnerBlockRepository(ctrl)
	adRepo := NewMockAdRepository(ctrl)
	svc, _ := NewService(&Config{PartnerBlockRepo: blockRepo, AdRepo: adRepo})

	blockRepo.EXPECT().
		GetByEmbedToken(gomock.Any(), "pb").
		Return(&models.PartnerBlock{ID: 7, EmbedToken: "pb", RevenueShareBPS: 7000}, nil)

	adRepo.EXPECT().
		ListAdCandidates(gomock.Any(), gomock.Any()).
		Return([]*models.AdCandidate{
			{
				CampaignID:        3,
				AdvertiserID:      4,
				DailyBudget:       1000,
				CPMPrice:          20000,
				SpentToday:        0,
				AdvertiserBalance: 1000,
				Ad:                &models.Ad{ID: 9},
			},
		}, nil)

	adRepo.EXPECT().
		ReserveImpression(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, reservation models.ImpressionReservation) (bool, error) {
			if reservation.CampaignID != 3 ||
				reservation.AdvertiserID != 4 ||
				reservation.PartnerBlockID != 7 ||
				reservation.Price != 20 ||
				reservation.PartnerReward != 14 ||
				reservation.PlatformRevenue != 6 {
				t.Fatalf("unexpected reservation: %+v", reservation)
			}
			return true, nil
		})

	result, err := svc.RequestAd(context.Background(), "pb", "")
	if err != nil {
		t.Fatalf("RequestAd: %v", err)
	}
	if result.Ad.ID != 9 || result.RequestID == "" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestRequestAdSavesClickContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	blockRepo := NewMockPartnerBlockRepository(ctrl)
	adRepo := NewMockAdRepository(ctrl)
	store := &fakeAdRequestStore{}
	svc, _ := NewService(&Config{PartnerBlockRepo: blockRepo, AdRepo: adRepo, AdRequestStore: store})

	blockRepo.EXPECT().
		GetByEmbedToken(gomock.Any(), "pb").
		Return(&models.PartnerBlock{ID: 7, EmbedToken: "pb", RevenueShareBPS: 0}, nil)
	adRepo.EXPECT().
		ListAdCandidates(gomock.Any(), gomock.Any()).
		Return([]*models.AdCandidate{
			{
				CampaignID:        3,
				AdvertiserID:      4,
				TopicID:           12,
				RegionID:          6,
				DailyBudget:       1000,
				CPMPrice:          20000,
				SpentToday:        0,
				AdvertiserBalance: 1000,
				Ad:                &models.Ad{ID: 9, TargetURL: "https://target.example"},
			},
		}, nil)
	adRepo.EXPECT().ReserveImpression(gomock.Any(), gomock.Any()).Return(true, nil)

	result, err := svc.RequestAd(context.Background(), "pb", "visitor-1")
	if err != nil {
		t.Fatalf("RequestAd: %v", err)
	}
	if store.record == nil || store.record.RequestID != result.RequestID || store.record.VisitorID != "visitor-1" ||
		store.record.AdID != 9 || store.record.TopicID != 12 || store.record.RegionID != 6 ||
		store.record.TargetURL != "https://target.example" ||
		store.ttl != adRequestTTL {
		t.Fatalf("unexpected saved record: record=%+v ttl=%s", store.record, store.ttl)
	}
}

func TestRequestAdPublishesImpression(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	blockRepo := NewMockPartnerBlockRepository(ctrl)
	adRepo := NewMockAdRepository(ctrl)
	store := &fakeAdRequestStore{}
	publisher := &fakeAdEventPublisher{}
	svc, _ := NewService(&Config{
		PartnerBlockRepo: blockRepo,
		AdRepo:           adRepo,
		AdRequestStore:   store,
		AdEventPublisher: publisher,
	})

	blockRepo.EXPECT().
		GetByEmbedToken(gomock.Any(), "pb").
		Return(&models.PartnerBlock{ID: 7, PartnerSiteID: 8, EmbedToken: "pb", RevenueShareBPS: 7000}, nil)
	adRepo.EXPECT().
		ListAdCandidates(gomock.Any(), gomock.Any()).
		Return([]*models.AdCandidate{
			{
				CampaignID:        3,
				AdvertiserID:      4,
				TopicID:           12,
				DailyBudget:       1000,
				CPMPrice:          20000,
				SpentToday:        0,
				AdvertiserBalance: 1000,
				Ad:                &models.Ad{ID: 9, AdGroupID: 5, TargetURL: "https://target.example"},
			},
		}, nil)
	adRepo.EXPECT().ReserveImpression(gomock.Any(), gomock.Any()).Return(true, nil)

	result, err := svc.RequestAd(context.Background(), "pb", "visitor-1")
	if err != nil {
		t.Fatalf("RequestAd: %v", err)
	}
	if len(publisher.events) != 1 {
		t.Fatalf("expected one event, got %d", len(publisher.events))
	}
	event := publisher.events[0]
	if event.EventID == "" || event.EventType != analytics.EventTypeImpression ||
		event.RequestID != result.RequestID || event.VisitorID != "visitor-1" ||
		event.AdvertiserID != 4 || event.CampaignID != 3 || event.AdGroupID != 5 ||
		event.AdID != 9 || event.PartnerBlockID != 7 || event.PartnerSiteID != 8 ||
		event.TopicID != 12 || event.Price != 20 || event.PartnerReward != 14 ||
		event.PlatformRevenue != 6 {
		t.Fatalf("unexpected impression event: %+v", event)
	}
}

func TestRequestAdDecoratesStoredImageURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	blockRepo := NewMockPartnerBlockRepository(ctrl)
	adRepo := NewMockAdRepository(ctrl)
	svc, _ := NewService(&Config{
		PartnerBlockRepo: blockRepo,
		AdRepo:           adRepo,
		AdStorage:        fakeAdStorage{baseURL: "https://cdn.example.com"},
	})

	blockRepo.EXPECT().
		GetByEmbedToken(gomock.Any(), "pb").
		Return(&models.PartnerBlock{ID: 7, EmbedToken: "pb", RevenueShareBPS: 0}, nil)
	adRepo.EXPECT().
		ListAdCandidates(gomock.Any(), gomock.Any()).
		Return([]*models.AdCandidate{
			{
				CampaignID:        3,
				AdvertiserID:      4,
				DailyBudget:       1000,
				CPMPrice:          20000,
				SpentToday:        0,
				AdvertiserBalance: 1000,
				Ad:                &models.Ad{ID: 9, ImageURL: "ads/banner.png"},
			},
		}, nil)
	adRepo.EXPECT().ReserveImpression(gomock.Any(), gomock.Any()).Return(true, nil)

	result, err := svc.RequestAd(context.Background(), "pb", "")
	if err != nil {
		t.Fatalf("RequestAd: %v", err)
	}
	if result.Ad.ImageURL != "https://cdn.example.com/ads/banner.png" {
		t.Fatalf("unexpected image url: %s", result.Ad.ImageURL)
	}
}

func TestClickAdTracksProfileAndReturnsTarget(t *testing.T) {
	store := &fakeAdRequestStore{record: &AdRequestRecord{
		RequestID: "req",
		VisitorID: "visitor-1",
		TopicID:   12,
		RegionID:  6,
		TargetURL: "https://target.example",
	}, clickOnce: true}
	profile := &fakeProfileClient{}
	svc, _ := NewService(&Config{AdRequestStore: store, ProfileClient: profile})

	targetURL, err := svc.ClickAd(context.Background(), "req")
	if err != nil {
		t.Fatalf("ClickAd: %v", err)
	}
	if targetURL != "https://target.example" {
		t.Fatalf("unexpected target url: %s", targetURL)
	}
	if profile.visitorID != "visitor-1" || profile.topicID != 12 || profile.regionID != 6 || profile.eventType != EventTypeClick {
		t.Fatalf("unexpected tracked event: %+v", profile)
	}
}

func TestClickAdPublishesClickOnlyOnce(t *testing.T) {
	store := &fakeAdRequestStore{record: &AdRequestRecord{
		RequestID:       "req",
		VisitorID:       "visitor-1",
		AdvertiserID:    4,
		CampaignID:      3,
		AdGroupID:       5,
		AdID:            9,
		PartnerBlockID:  7,
		PartnerSiteID:   8,
		TopicID:         12,
		RegionID:        6,
		TargetURL:       "https://target.example",
		Price:           20,
		PartnerReward:   14,
		PlatformRevenue: 6,
	}, clickOnce: true}
	profile := &fakeProfileClient{}
	publisher := &fakeAdEventPublisher{}
	svc, _ := NewService(&Config{AdRequestStore: store, ProfileClient: profile, AdEventPublisher: publisher})

	if _, err := svc.ClickAd(context.Background(), "req"); err != nil {
		t.Fatalf("first ClickAd: %v", err)
	}
	if _, err := svc.ClickAd(context.Background(), "req"); err != nil {
		t.Fatalf("second ClickAd: %v", err)
	}
	if len(publisher.events) != 1 {
		t.Fatalf("expected one click event, got %d", len(publisher.events))
	}
	event := publisher.events[0]
	if event.EventID == "" || event.EventType != analytics.EventTypeClick ||
		event.RequestID != "req" || event.VisitorID != "visitor-1" ||
		event.AdvertiserID != 4 || event.CampaignID != 3 || event.AdGroupID != 5 ||
		event.AdID != 9 || event.PartnerBlockID != 7 || event.PartnerSiteID != 8 ||
		event.TopicID != 12 || event.Price != 0 || event.PartnerReward != 0 ||
		event.PlatformRevenue != 0 {
		t.Fatalf("unexpected click event: %+v", event)
	}
	if profile.visitorID != "visitor-1" || profile.topicID != 12 || profile.regionID != 6 || profile.eventType != EventTypeClick {
		t.Fatalf("unexpected tracked event: %+v", profile)
	}
}

type fakeAdRequestStore struct {
	record    *AdRequestRecord
	ttl       time.Duration
	clickOnce bool
}

func (s *fakeAdRequestStore) Save(_ context.Context, record AdRequestRecord, ttl time.Duration) error {
	s.record = &record
	s.ttl = ttl
	return nil
}

func (s *fakeAdRequestStore) Get(_ context.Context, requestID string) (*AdRequestRecord, error) {
	if s.record == nil {
		return nil, errNotFoundForTest
	}
	return s.record, nil
}

func (s *fakeAdRequestStore) MarkClickedOnce(context.Context, string, time.Duration) (bool, error) {
	if !s.clickOnce {
		return false, nil
	}
	s.clickOnce = false
	return true, nil
}

type fakeAdEventPublisher struct {
	events []analytics.AdEvent
}

func (p *fakeAdEventPublisher) PublishAdEvent(_ context.Context, event analytics.AdEvent) error {
	p.events = append(p.events, event)
	return nil
}

type fakeProfileClient struct {
	visitorID string
	topicID   int
	regionID  int
	eventType string
}

func (c *fakeProfileClient) GetProfile(context.Context, string) (*Profile, bool, error) {
	return nil, false, nil
}

func (c *fakeProfileClient) TrackEvent(_ context.Context, visitorID string, topicID, regionID int, eventType string) error {
	c.visitorID = visitorID
	c.topicID = topicID
	c.regionID = regionID
	c.eventType = eventType
	return nil
}

var errNotFoundForTest = context.Canceled

type fakeAdStorage struct {
	baseURL string
}

func (s fakeAdStorage) UploadAdImage(context.Context, []byte, string, string) (string, error) {
	return "", nil
}

func (s fakeAdStorage) DeleteAdImage(context.Context, string) error {
	return nil
}

func (s fakeAdStorage) GetAdImageURL(imageKey string) string {
	return s.baseURL + "/" + imageKey
}
