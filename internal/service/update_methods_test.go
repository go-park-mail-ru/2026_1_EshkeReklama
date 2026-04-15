package service

import (
	"context"
	"database/sql"
	"testing"

	"eshkere/internal/handler/dto"
	"eshkere/internal/models"

	"go.uber.org/mock/gomock"
)

func TestUpdateAdCampaign_UpdatesProvidedFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockAdCampaignRepository(ctrl)
	svc, _ := NewService(&Config{AdCampaignRepo: repo})

	current := &models.AdCampaign{ID: 1, AdvertiserID: 7, Status: models.AdStatusWorking, Name: "old", DailyBudget: 10}
	repo.EXPECT().GetByID(gomock.Any(), 1).Return(current, nil)
	repo.EXPECT().Update(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, c *models.AdCampaign) error {
			if c.Name != "new" || c.DailyBudget != 99 || c.Status != models.AdStatusRejected {
				t.Fatalf("unexpected updated: %+v", c)
			}
			if c.UpdatedAt.Valid != true {
				t.Fatalf("expected UpdatedAt set")
			}
			return nil
		})

	name := "new"
	budget := int64(99)
	status := models.AdStatusRejected
	if err := svc.UpdateAdCampaign(context.Background(), 1, dto.UpdateAdCampaignRequest{Name: &name, DailyBudget: &budget, Status: &status}); err != nil {
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
	if err := svc.UpdateAdGroup(context.Background(), 1, dto.UpdateAdGroupRequest{Name: &name, AgeFrom: &ageFrom}); err != nil {
		t.Fatalf("UpdateAdGroup: %v", err)
	}
}

func TestCreateAd_SetsModerationStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockAdRepository(ctrl)
	svc, _ := NewService(&Config{AdRepo: repo})

	ad := &models.Ad{AdGroupID: 2, Title: "t"}
	repo.EXPECT().Create(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, a *models.Ad) error {
			if a.Status != models.AdStatusModeration {
				t.Fatalf("expected moderation, got %v", a.Status)
			}
			return nil
		})

	if _, err := svc.CreateAd(context.Background(), ad); err != nil {
		t.Fatalf("CreateAd: %v", err)
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
	if err := svc.UpdateAd(context.Background(), 9, dto.UpdateAdRequest{Title: &title}); err != nil {
		t.Fatalf("UpdateAd: %v", err)
	}
}

func TestGenerateFeedLink_InvalidArgs(t *testing.T) {
	svc, _ := NewService(&Config{})
	if _, err := svc.GenerateFeedLink(context.Background(), 0); err == nil {
		t.Fatalf("expected error")
	}
}

func TestUpdateAdvertiserProfile_InvalidPhone(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	advRepo := NewMockAdvertiserRepository(ctrl)
	svc, _ := NewService(&Config{AdvertiserRepo: advRepo})

	advRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&models.Advertiser{ID: 1}, nil)

	_, err := svc.UpdateAdvertiserProfile(context.Background(), 1, "", "", "bad")
	if err == nil {
		t.Fatalf("expected error")
	}
	_ = sql.ErrNoRows // keep sql imported via other tests in package
}
