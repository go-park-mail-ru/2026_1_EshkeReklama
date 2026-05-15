package service

import (
	"context"
	"database/sql"
	"testing"

	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"

	"go.uber.org/mock/gomock"
)

func TestUpdateAdCampaign_UpdatesProvidedFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockAdCampaignRepository(ctrl)
	svc, _ := NewService(&Config{AdCampaignRepo: repo})

	current := &models.AdCampaign{ID: 1, AdvertiserID: 7, Status: models.AdStatusWorking, Name: "old"}
	repo.EXPECT().GetByID(gomock.Any(), 1).Return(current, nil)
	repo.EXPECT().Update(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, c *models.AdCampaign) error {
			if c.Name != "new" || c.Status != models.AdStatusRejected {
				t.Fatalf("unexpected updated: %+v", c)
			}
			if c.UpdatedAt.Valid != true {
				t.Fatalf("expected UpdatedAt set")
			}
			return nil
		})

	name := "new"
	status := models.AdStatusRejected
	if err := svc.UpdateAdCampaign(context.Background(), &serviceinput.UpdateAdCampaign{
		ID:     1,
		Name:   &name,
		Status: &status,
	}); err != nil {
		t.Fatalf("UpdateAdCampaign: %v", err)
	}
}

func TestUpdateAdGroup_UpdatesProvidedFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockAdGroupRepository(ctrl)
	svc, _ := NewService(&Config{AdGroupRepo: repo})

	current := &models.AdGroup{ID: 1, AdCampaignID: 2, TopicID: 1, RegionID: 1, Name: "old", AgeFrom: 18, AgeTo: 25, Gender: "any"}
	repo.EXPECT().GetByID(gomock.Any(), 1).Return(current, nil)
	repo.EXPECT().Update(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, g *models.AdGroup) error {
			if g.Name != "new" || g.AgeFrom != 21 {
				t.Fatalf("unexpected group: %+v", g)
			}
			if g.UpdatedAt.Valid != true {
				t.Fatalf("expected UpdatedAt set")
			}
			return nil
		})

	name := "new"
	ageFrom := 21
	if err := svc.UpdateAdGroup(context.Background(), &serviceinput.UpdateAdGroup{
		ID:      1,
		Name:    &name,
		AgeFrom: &ageFrom,
	}); err != nil {
		t.Fatalf("UpdateAdGroup: %v", err)
	}
}

func TestCreateAd_SetsModerationStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockAdRepository(ctrl)
	svc, _ := NewService(&Config{AdRepo: repo})

	repo.EXPECT().Create(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, a *models.Ad) error {
			if a.AdGroupID != 2 || a.Title != "t" || a.Status != models.AdStatusModeration {
				t.Fatalf("unexpected ad: %+v", a)
			}
			return nil
		})

	if _, err := svc.CreateAd(context.Background(), &serviceinput.CreateAd{
		AdGroupID: 2,
		Title:     "t",
	}); err != nil {
		t.Fatalf("CreateAd: %v", err)
	}
}

func TestGetAdByID_DecoratesImageURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockAdRepository(ctrl)
	storage := NewMockAdStorage(ctrl)
	svc, _ := NewService(&Config{AdRepo: repo, AdStorage: storage})

	ad := &models.Ad{ID: 9, ImageURL: "images/ad.png"}
	repo.EXPECT().GetByID(gomock.Any(), 9).Return(ad, nil)
	storage.EXPECT().GetAdImageURL("images/ad.png").Return("https://cdn.example/ad.png")

	got, err := svc.GetAdByID(context.Background(), 9)
	if err != nil {
		t.Fatalf("GetAdByID: %v", err)
	}
	if got.ImageURL != "https://cdn.example/ad.png" {
		t.Fatalf("expected decorated image URL, got %q", got.ImageURL)
	}
}

func TestListModerationAds_UsesModerationStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockAdRepository(ctrl)
	svc, _ := NewService(&Config{AdRepo: repo})

	repo.EXPECT().ListByStatus(gomock.Any(), models.AdStatusModeration).Return([]*models.Ad{
		{ID: 1, Status: models.AdStatusModeration},
	}, nil)

	ads, err := svc.ListModerationAds(context.Background())
	if err != nil {
		t.Fatalf("ListModerationAds: %v", err)
	}
	if len(ads) != 1 || ads[0].Status != models.AdStatusModeration {
		t.Fatalf("unexpected ads: %+v", ads)
	}
}

func TestUpdateAd_SetsUpdatedAt(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockAdRepository(ctrl)
	svc, _ := NewService(&Config{AdRepo: repo})

	current := &models.Ad{ID: 9, Title: "old"}
	repo.EXPECT().GetByID(gomock.Any(), 9).Return(current, nil)
	repo.EXPECT().Update(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, a *models.Ad) error {
			if a.Title != "new" {
				t.Fatalf("expected title updated")
			}
			if a.UpdatedAt.Valid != true {
				t.Fatalf("expected UpdatedAt set")
			}
			return nil
		})

	title := "new"
	if err := svc.UpdateAd(context.Background(), &serviceinput.UpdateAd{ID: 9, Title: &title}); err != nil {
		t.Fatalf("UpdateAd: %v", err)
	}
}

func TestUpdateAdModerationStatus_ApproveWithEnoughBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	adRepo := NewMockAdRepository(ctrl)
	groupRepo := NewMockAdGroupRepository(ctrl)
	campaignRepo := NewMockAdCampaignRepository(ctrl)
	advertiserRepo := NewMockAdvertiserRepository(ctrl)
	svc, _ := NewService(&Config{
		AdRepo:         adRepo,
		AdGroupRepo:    groupRepo,
		AdCampaignRepo: campaignRepo,
		AdvertiserRepo: advertiserRepo,
	})

	ad := &models.Ad{ID: 9, AdGroupID: 3, Status: models.AdStatusModeration}
	group := &models.AdGroup{ID: 3, AdCampaignID: 2}
	campaign := &models.AdCampaign{ID: 2, AdvertiserID: 7, Status: models.AdStatusModeration, DailyBudget: 100, CPMPrice: 1000}
	advertiser := &models.Advertiser{ID: 7, Balance: 10}

	adRepo.EXPECT().GetByID(gomock.Any(), 9).Return(ad, nil)
	groupRepo.EXPECT().GetByID(gomock.Any(), 3).Return(group, nil)
	campaignRepo.EXPECT().GetByID(gomock.Any(), 2).Return(campaign, nil)
	advertiserRepo.EXPECT().GetByID(gomock.Any(), 7).Return(advertiser, nil)
	adRepo.EXPECT().Update(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, updated *models.Ad) error {
			if updated.Status != models.AdStatusWorking {
				t.Fatalf("expected ad working, got %s", updated.Status)
			}
			if !updated.UpdatedAt.Valid {
				t.Fatalf("expected ad UpdatedAt set")
			}
			return nil
		})
	adRepo.EXPECT().ListByAdCampaignID(gomock.Any(), 2).Return([]*models.Ad{
		{ID: 9, Status: models.AdStatusWorking},
	}, nil)
	campaignRepo.EXPECT().Update(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, updated *models.AdCampaign) error {
			if updated.Status != models.AdStatusWorking {
				t.Fatalf("expected campaign working, got %s", updated.Status)
			}
			if !updated.UpdatedAt.Valid {
				t.Fatalf("expected campaign UpdatedAt set")
			}
			return nil
		})

	err := svc.UpdateAdModerationStatus(context.Background(), &serviceinput.UpdateAdStatus{
		AdID:     9,
		Decision: serviceinput.AdModerationApprove,
	})
	if err != nil {
		t.Fatalf("UpdateAdModerationStatus: %v", err)
	}
}

func TestUpdateAdModerationStatus_ApproveWithoutEnoughBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	adRepo := NewMockAdRepository(ctrl)
	groupRepo := NewMockAdGroupRepository(ctrl)
	campaignRepo := NewMockAdCampaignRepository(ctrl)
	advertiserRepo := NewMockAdvertiserRepository(ctrl)
	svc, _ := NewService(&Config{
		AdRepo:         adRepo,
		AdGroupRepo:    groupRepo,
		AdCampaignRepo: campaignRepo,
		AdvertiserRepo: advertiserRepo,
	})

	adRepo.EXPECT().GetByID(gomock.Any(), 9).Return(&models.Ad{ID: 9, AdGroupID: 3}, nil)
	groupRepo.EXPECT().GetByID(gomock.Any(), 3).Return(&models.AdGroup{ID: 3, AdCampaignID: 2}, nil)
	campaignRepo.EXPECT().GetByID(gomock.Any(), 2).Return(&models.AdCampaign{
		ID: 2, AdvertiserID: 7, DailyBudget: 100, CPMPrice: 1000,
	}, nil)
	advertiserRepo.EXPECT().GetByID(gomock.Any(), 7).Return(&models.Advertiser{ID: 7, Balance: 0}, nil)
	adRepo.EXPECT().Update(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, updated *models.Ad) error {
			if updated.Status != models.AdStatusNotEnoughMoney {
				t.Fatalf("expected ad not_enough_money, got %s", updated.Status)
			}
			return nil
		})
	adRepo.EXPECT().ListByAdCampaignID(gomock.Any(), 2).Return([]*models.Ad{
		{ID: 9, Status: models.AdStatusNotEnoughMoney},
	}, nil)
	campaignRepo.EXPECT().Update(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, updated *models.AdCampaign) error {
			if updated.Status != models.AdStatusNotEnoughMoney {
				t.Fatalf("expected campaign not_enough_money, got %s", updated.Status)
			}
			return nil
		})

	err := svc.UpdateAdModerationStatus(context.Background(), &serviceinput.UpdateAdStatus{
		AdID:     9,
		Decision: serviceinput.AdModerationApprove,
	})
	if err != nil {
		t.Fatalf("UpdateAdModerationStatus: %v", err)
	}
}

func TestCampaignStatusFromAds_Priority(t *testing.T) {
	status := campaignStatusFromAds([]*models.Ad{
		{Status: models.AdStatusRejected},
		{Status: models.AdStatusModeration},
		{Status: models.AdStatusNotEnoughMoney},
	})
	if status != models.AdStatusModeration {
		t.Fatalf("expected moderation, got %s", status)
	}

	status = campaignStatusFromAds([]*models.Ad{
		{Status: models.AdStatusRejected},
		{Status: models.AdStatusWorking},
	})
	if status != models.AdStatusWorking {
		t.Fatalf("expected working, got %s", status)
	}
}

func TestGenerateFeedLink_InvalidArgs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	campaignRepo := NewMockAdCampaignRepository(ctrl)
	svc, _ := NewService(&Config{AdCampaignRepo: campaignRepo})

	campaignRepo.EXPECT().GetByID(gomock.Any(), 0).Return(nil, sql.ErrNoRows)

	if _, err := svc.GenerateFeedLink(context.Background(), 0); err == nil {
		t.Fatalf("expected error")
	}
}

func TestUpdateAdvertiserProfile_InvalidInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	advRepo := NewMockAdvertiserRepository(ctrl)
	svc, _ := NewService(&Config{AdvertiserRepo: advRepo})

	_, err := svc.UpdateAdvertiserProfile(context.Background(), nil)
	if err == nil {
		t.Fatalf("expected error")
	}
	_ = sql.ErrNoRows // keep sql imported via other tests in package
}
