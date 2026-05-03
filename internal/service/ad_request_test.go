package service

import (
	"context"
	"testing"

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
	})

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

	result, err := svc.RequestAd(context.Background(), "pb")
	if err != nil {
		t.Fatalf("RequestAd: %v", err)
	}
	if result.Ad.ID != 9 || result.RequestID == "" {
		t.Fatalf("unexpected result: %+v", result)
	}
}
