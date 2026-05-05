package service

import (
	"context"
	"testing"
	"time"

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
	}, nil)

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
		store.record.AdID != 9 || store.record.TopicID != 12 || store.record.TargetURL != "https://target.example" ||
		store.ttl != adRequestTTL {
		t.Fatalf("unexpected saved record: record=%+v ttl=%s", store.record, store.ttl)
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
		TargetURL: "https://target.example",
	}}
	profile := &fakeProfileClient{}
	svc, _ := NewService(&Config{AdRequestStore: store, ProfileClient: profile})

	targetURL, err := svc.ClickAd(context.Background(), "req")
	if err != nil {
		t.Fatalf("ClickAd: %v", err)
	}
	if targetURL != "https://target.example" {
		t.Fatalf("unexpected target url: %s", targetURL)
	}
	if profile.visitorID != "visitor-1" || profile.topicID != 12 || profile.eventType != EventTypeClick {
		t.Fatalf("unexpected tracked event: %+v", profile)
	}
}

type fakeAdRequestStore struct {
	record *AdRequestRecord
	ttl    time.Duration
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

type fakeProfileClient struct {
	visitorID string
	topicID   int
	eventType string
}

func (c *fakeProfileClient) GetProfile(context.Context, string) ([]TopicScore, bool, error) {
	return nil, false, nil
}

func (c *fakeProfileClient) TrackEvent(_ context.Context, visitorID string, topicID int, eventType string) error {
	c.visitorID = visitorID
	c.topicID = topicID
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
