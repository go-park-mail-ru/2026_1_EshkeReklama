package service

import (
	"context"
	"testing"

	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"

	"go.uber.org/mock/gomock"
)

func TestService_PassthroughMethods(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cRepo := NewMockAdCampaignRepository(ctrl)
	gRepo := NewMockAdGroupRepository(ctrl)
	aRepo := NewMockAdRepository(ctrl)

	svc, _ := NewService(&Config{AdCampaignRepo: cRepo, AdGroupRepo: gRepo, AdRepo: aRepo})

	cRepo.EXPECT().Create(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, c *models.AdCampaign) error {
			if c.AdvertiserID != 1 || c.Name != "camp" || c.Status != models.AdStatusModeration {
				t.Fatalf("unexpected campaign: %+v", c)
			}
			return nil
		})
	if _, err := svc.CreateAdCampaign(context.Background(), &serviceinput.CreateAdCampaign{
		AdvertiserID: 1,
		Name:         "camp",
	}); err != nil {
		t.Fatalf("CreateAdCampaign: %v", err)
	}
	cRepo.EXPECT().ListByAdvertiserID(gomock.Any(), 1).Return([]*models.AdCampaign{}, nil)
	if _, err := svc.ListAdCampaigns(context.Background(), 1); err != nil {
		t.Fatalf("ListAdCampaigns: %v", err)
	}
	cRepo.EXPECT().Delete(gomock.Any(), 9).Return(nil)
	if err := svc.DeleteAdCampaign(context.Background(), 9); err != nil {
		t.Fatalf("DeleteAdCampaign: %v", err)
	}

	gRepo.EXPECT().Create(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, g *models.AdGroup) error {
			if g.AdCampaignID != 1 || g.Name != "group" || g.Gender != "any" {
				t.Fatalf("unexpected group: %+v", g)
			}
			return nil
		})
	if _, err := svc.CreateAdGroup(context.Background(), &serviceinput.CreateAdGroup{
		AdCampaignID: 1,
		TopicID:      1,
		RegionID:     2,
		Name:         "group",
		AgeFrom:      18,
		AgeTo:        25,
		Gender:       "any",
	}); err != nil {
		t.Fatalf("CreateAdGroup: %v", err)
	}
	gRepo.EXPECT().ListByCampaignID(gomock.Any(), 1).Return([]*models.AdGroup{}, nil)
	if _, err := svc.ListAdGroups(context.Background(), 1); err != nil {
		t.Fatalf("ListAdGroups: %v", err)
	}
	gRepo.EXPECT().Delete(gomock.Any(), 3).Return(nil)
	if err := svc.DeleteAdGroup(context.Background(), 3); err != nil {
		t.Fatalf("DeleteAdGroup: %v", err)
	}

	aRepo.EXPECT().ListByAdGroupID(gomock.Any(), 2).Return([]*models.Ad{}, nil)
	if _, err := svc.ListAds(context.Background(), 2); err != nil {
		t.Fatalf("ListAds: %v", err)
	}
	aRepo.EXPECT().Delete(gomock.Any(), 4).Return(nil)
	if err := svc.DeleteAd(context.Background(), 4); err != nil {
		t.Fatalf("DeleteAd: %v", err)
	}
}
