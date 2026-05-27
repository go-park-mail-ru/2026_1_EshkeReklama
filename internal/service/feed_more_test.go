package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	errs "eshkere/internal/errors"
	"eshkere/internal/models"

	"go.uber.org/mock/gomock"
)

func TestGenerateFeedLinkAndGetAdByFeedToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	campaignRepo := NewMockAdCampaignRepository(ctrl)
	feedRepo := NewMockFeedLinkRepository(ctrl)
	adRepo := NewMockAdRepository(ctrl)
	adStorage := NewMockAdStorage(ctrl)

	svc, _ := NewService(&Config{
		AdCampaignRepo: campaignRepo,
		FeedLinkRepo:   feedRepo,
		AdRepo:         adRepo,
		AdStorage:      adStorage,
	})

	campaignRepo.EXPECT().GetByID(gomock.Any(), 0).Return(nil, sql.ErrNoRows)
	if _, err := svc.GenerateFeedLink(context.Background(), 0); !errors.Is(err, errs.BadRequestError) {
		t.Fatalf("expected bad request for invalid campaign, got %v", err)
	}

	campaignRepo.EXPECT().GetByID(gomock.Any(), 7).Return(&models.AdCampaign{ID: 7}, nil)
	feedRepo.EXPECT().Create(gomock.Any(), 7, gomock.Any()).Return(nil)
	link, err := svc.GenerateFeedLink(context.Background(), 7)
	if err != nil || len(link) <= len(BaseFeedURL) || link[:len(BaseFeedURL)] != BaseFeedURL {
		t.Fatalf("GenerateFeedLink mismatch link=%s err=%v", link, err)
	}

	if _, err := svc.GetAdByFeedToken(context.Background(), ""); !errors.Is(err, errs.ErrInvalidAdvertiserArg) {
		t.Fatalf("expected invalid token error, got %v", err)
	}
	feedRepo.EXPECT().GetCampaignIDByToken(gomock.Any(), "missing").Return(0, sql.ErrNoRows)
	if _, err := svc.GetAdByFeedToken(context.Background(), "missing"); !errors.Is(err, errs.NotFoundError) {
		t.Fatalf("expected not found for missing token, got %v", err)
	}

	feedRepo.EXPECT().GetCampaignIDByToken(gomock.Any(), "broken").Return(9, nil)
	adRepo.EXPECT().ListByAdCampaignID(gomock.Any(), 9).Return(nil, errors.New("db down"))
	if ad, err := svc.GetAdByFeedToken(context.Background(), "broken"); err != nil || ad == nil || ad.ID != 0 {
		t.Fatalf("expected empty ad on repo error, got ad=%+v err=%v", ad, err)
	}

	feedRepo.EXPECT().GetCampaignIDByToken(gomock.Any(), "ok").Return(9, nil)
	ad := &models.Ad{ID: 5, ImageURL: "images/ad.png"}
	adRepo.EXPECT().ListByAdCampaignID(gomock.Any(), 9).Return([]*models.Ad{ad}, nil)
	adStorage.EXPECT().GetAdImageURL("images/ad.png").Return("https://cdn.example/ad.png")
	got, err := svc.GetAdByFeedToken(context.Background(), "ok")
	if err != nil || got.ID != 5 || got.ImageURL != "https://cdn.example/ad.png" {
		t.Fatalf("GetAdByFeedToken mismatch got=%+v err=%v", got, err)
	}

	token, err := generateToken(16)
	if err != nil || len(token) != 32 {
		t.Fatalf("generateToken mismatch token=%q err=%v", token, err)
	}
}
