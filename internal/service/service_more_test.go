package service

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"eshkere/internal/models"

	"go.uber.org/mock/gomock"
)

func TestTopUpAdvertiserBalance_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	advRepo := NewMockAdvertiserRepository(ctrl)
	svc, _ := NewService(&Config{AdvertiserRepo: advRepo})

	adv := &models.Advertiser{ID: 1, Balance: 100}
	advRepo.EXPECT().GetByID(gomock.Any(), 1).Return(adv, nil)
	advRepo.EXPECT().Update(gomock.Any(), adv).Return(nil)

	got, err := svc.TopUpAdvertiserBalance(context.Background(), 1, 50)
	if err != nil {
		t.Fatalf("TopUpAdvertiserBalance: %v", err)
	}
	if got != 150 {
		t.Fatalf("expected 150 got %d", got)
	}
}

func TestUpdateAdvertiserAvatar_UploadsAvatar(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	advRepo := NewMockAdvertiserRepository(ctrl)
	st := NewMockAvatarStorage(ctrl)

	svc, _ := NewService(&Config{AdvertiserRepo: advRepo, AvatarStorage: st})

	adv := &models.Advertiser{ID: 1}
	advRepo.EXPECT().GetByID(gomock.Any(), 1).Return(adv, nil)
	st.EXPECT().
		UploadAvatar(gomock.Any(), 1, []byte("img"), "a.png", "image/png").
		Return("https://cdn/avatar.png", nil)
	advRepo.EXPECT().Update(gomock.Any(), adv).Return(nil)
	st.EXPECT().
		GetAvatarURL("https://cdn/avatar.png").
		Return("https://cdn/avatar.png")

	updated, err := svc.UpdateAdvertiserAvatar(context.Background(), 1, []byte("img"), "a.png", "image/png")
	if err != nil {
		t.Fatalf("UpdateAdvertiserAvatar: %v", err)
	}
	if updated.AvatarURL.Valid != true {
		t.Fatalf("expected avatar url to be set")
	}
}

func TestUpdateAdvertiserAvatar_DeletesPrevious(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	advRepo := NewMockAdvertiserRepository(ctrl)
	st := NewMockAvatarStorage(ctrl)

	svc, _ := NewService(&Config{AdvertiserRepo: advRepo, AvatarStorage: st})

	adv := &models.Advertiser{
		ID:        1,
		AvatarURL: sql.NullString{String: "https://cdn/avatars/1/old.png", Valid: true},
	}
	advRepo.EXPECT().GetByID(gomock.Any(), 1).Return(adv, nil)
	st.EXPECT().
		UploadAvatar(gomock.Any(), 1, []byte("img"), "a.png", "image/png").
		Return("https://cdn/avatars/1/new.png", nil)
	advRepo.EXPECT().Update(gomock.Any(), adv).Return(nil)
	st.EXPECT().
		DeleteAvatar(gomock.Any(), 1, "https://cdn/avatars/1/old.png").
		Return(nil)
	st.EXPECT().
		GetAvatarURL("https://cdn/avatars/1/new.png").
		Return("https://cdn/avatars/1/new.png")

	_, err := svc.UpdateAdvertiserAvatar(context.Background(), 1, []byte("img"), "a.png", "image/png")
	if err != nil {
		t.Fatalf("UpdateAdvertiserAvatar: %v", err)
	}
}

func TestGenerateFeedLink_And_GetAdsByFeedToken_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	feedRepo := NewMockFeedLinkRepository(ctrl)
	adRepo := NewMockAdRepository(ctrl)
	campaignRepo := NewMockAdCampaignRepository(ctrl)

	svc, _ := NewService(&Config{FeedLinkRepo: feedRepo, AdRepo: adRepo, AdCampaignRepo: campaignRepo})

	campaignRepo.EXPECT().
		GetByID(gomock.Any(), 1).
		Return(&models.AdCampaign{ID: 1}, nil)
	feedRepo.EXPECT().
		Create(gomock.Any(), 1, gomock.Any()).
		Return(nil)

	token, err := svc.GenerateFeedLink(context.Background(), 1)
	if err != nil {
		t.Fatalf("GenerateFeedLink: %v", err)
	}
	if token == "" || token == BaseFeedURL {
		t.Fatalf("expected token")
	}

	feedRepo.EXPECT().GetCampaignIDByToken(gomock.Any(), "t").Return(1, nil)
	adRepo.EXPECT().ListByAdCampaignID(gomock.Any(), 1).Return([]*models.Ad{}, nil)

	ads, err := svc.GetAdsByFeedToken(context.Background(), "t")
	if err != nil {
		t.Fatalf("GetAdsByFeedToken: %v", err)
	}
	if len(ads) != 0 {
		t.Fatalf("expected empty ads")
	}
}

func TestGetAdsByFeedToken_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	feedRepo := NewMockFeedLinkRepository(ctrl)
	svc, _ := NewService(&Config{FeedLinkRepo: feedRepo, AdRepo: NewMockAdRepository(ctrl)})

	feedRepo.EXPECT().GetCampaignIDByToken(gomock.Any(), "t").Return(0, sql.ErrNoRows)

	_, err := svc.GetAdsByFeedToken(context.Background(), "t")
	if err == nil {
		t.Fatalf("expected error")
	}
	if err == sql.ErrNoRows {
		t.Fatalf("expected wrapped not-found error, got %v", err)
	}
}

func TestGetPartnerBlockEmbedCode_ReturnsIframeSnippet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	siteRepo := NewMockPartnerSiteRepository(ctrl)
	blockRepo := NewMockPartnerBlockRepository(ctrl)
	geoRepo := NewMockPartnerBlockGeoRuleRepository(ctrl)

	svc, _ := NewService(&Config{
		PartnerSiteRepo:         siteRepo,
		PartnerBlockRepo:        blockRepo,
		PartnerBlockGeoRuleRepo: geoRepo,
	})

	siteRepo.EXPECT().GetByID(gomock.Any(), 12).Return(&models.PartnerSite{ID: 12, PartnerID: 7}, nil)
	blockRepo.EXPECT().GetByID(gomock.Any(), 34).Return(&models.PartnerBlock{ID: 34, PartnerSiteID: 12, EmbedToken: "pb_abc123"}, nil)
	geoRepo.EXPECT().ListByBlockID(gomock.Any(), 34).Return([]*models.PartnerBlockGeoRule{}, nil)

	embedToken, iframeURL, htmlSnippet, err := svc.GetPartnerBlockEmbedCode(context.Background(), 7, 12, 34, "https://ads.example.com/")
	if err != nil {
		t.Fatalf("GetPartnerBlockEmbedCode: %v", err)
	}
	if embedToken != "pb_abc123" {
		t.Fatalf("unexpected embed token: %s", embedToken)
	}
	if iframeURL != "https://ads.example.com/public/partner/blocks/pb_abc123/frame" {
		t.Fatalf("unexpected iframe url: %s", iframeURL)
	}
	if !strings.Contains(htmlSnippet, `<iframe src="https://ads.example.com/public/partner/blocks/pb_abc123/frame"`) {
		t.Fatalf("html snippet must contain iframe url, got: %s", htmlSnippet)
	}
	if !strings.Contains(htmlSnippet, `loading="lazy"`) || !strings.Contains(htmlSnippet, `referrerpolicy="strict-origin-when-cross-origin"`) {
		t.Fatalf("html snippet must contain iframe attributes, got: %s", htmlSnippet)
	}
}
