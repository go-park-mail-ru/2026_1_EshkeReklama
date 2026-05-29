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
	advRepo := NewMockAdvertiserRepository(ctrl)

	svc, _ := NewService(&Config{AdCampaignRepo: cRepo, AdGroupRepo: gRepo, AdRepo: aRepo, AdvertiserRepo: advRepo})

	advRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&models.Advertiser{ID: 1, Tariff: models.TariffTypeBasic}, nil)
	cRepo.EXPECT().CountActiveByAdvertiserID(gomock.Any(), 1).Return(0, nil)
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
	cRepo.EXPECT().GetByID(gomock.Any(), 9).Return(&models.AdCampaign{ID: 9, AdvertiserID: 1}, nil)
	cRepo.EXPECT().Delete(gomock.Any(), 9).Return(nil)
	if err := svc.DeleteAdCampaign(context.Background(), 1, 9); err != nil {
		t.Fatalf("DeleteAdCampaign: %v", err)
	}

	cRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&models.AdCampaign{ID: 1, AdvertiserID: 1}, nil)
	gRepo.EXPECT().Create(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, g *models.AdGroup) error {
			if g.AdCampaignID != 1 || g.Name != "group" || g.Gender != "any" {
				t.Fatalf("unexpected group: %+v", g)
			}
			return nil
		})
	if _, err := svc.CreateAdGroup(context.Background(), 1, &serviceinput.CreateAdGroup{
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
	cRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&models.AdCampaign{ID: 1, AdvertiserID: 1}, nil)
	gRepo.EXPECT().ListByCampaignID(gomock.Any(), 1).Return([]*models.AdGroup{}, nil)
	if _, err := svc.ListAdGroups(context.Background(), 1, 1); err != nil {
		t.Fatalf("ListAdGroups: %v", err)
	}
	gRepo.EXPECT().GetByID(gomock.Any(), 3).Return(&models.AdGroup{ID: 3, AdCampaignID: 1}, nil)
	cRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&models.AdCampaign{ID: 1, AdvertiserID: 1}, nil)
	gRepo.EXPECT().Delete(gomock.Any(), 3).Return(nil)
	cRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&models.AdCampaign{ID: 1, AdvertiserID: 1}, nil)
	aRepo.EXPECT().ListByAdCampaignID(gomock.Any(), 1).Return([]*models.Ad{}, nil)
	cRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	if err := svc.DeleteAdGroup(context.Background(), 1, 3); err != nil {
		t.Fatalf("DeleteAdGroup: %v", err)
	}

	gRepo.EXPECT().GetByID(gomock.Any(), 2).Return(&models.AdGroup{ID: 2, AdCampaignID: 1}, nil)
	cRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&models.AdCampaign{ID: 1, AdvertiserID: 1}, nil)
	aRepo.EXPECT().ListByAdGroupID(gomock.Any(), 2).Return([]*models.Ad{}, nil)
	if _, err := svc.ListAds(context.Background(), 1, 2); err != nil {
		t.Fatalf("ListAds: %v", err)
	}
	aRepo.EXPECT().GetByID(gomock.Any(), 4).Return(&models.Ad{ID: 4, AdGroupID: 2}, nil)
	gRepo.EXPECT().GetByID(gomock.Any(), 2).Return(&models.AdGroup{ID: 2, AdCampaignID: 1}, nil)
	cRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&models.AdCampaign{ID: 1, AdvertiserID: 1}, nil)
	aRepo.EXPECT().Delete(gomock.Any(), 4).Return(nil)
	cRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&models.AdCampaign{ID: 1, AdvertiserID: 1}, nil)
	aRepo.EXPECT().ListByAdCampaignID(gomock.Any(), 1).Return([]*models.Ad{}, nil)
	cRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	if err := svc.DeleteAd(context.Background(), 1, 4); err != nil {
		t.Fatalf("DeleteAd: %v", err)
	}
}
