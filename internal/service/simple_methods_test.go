package service

import (
	"context"
	"testing"

	"eshkere/internal/models"

	"go.uber.org/mock/gomock"
)

func TestService_PassthroughMethods(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cRepo := NewMockAdCampaignRepository(ctrl)
	gRepo := NewMockAdGroupRepository(ctrl)
	aRepo := NewMockAdRepository(ctrl)

	svc, _ := NewService(&Config{AdCampaignRepo: cRepo, AdGroupRepo: gRepo, AdRepo: aRepo})

	c := &models.AdCampaign{AdvertiserID: 1}
	cRepo.EXPECT().Create(gomock.Any(), c).Return(nil)
	if _, err := svc.CreateAdCampaign(context.Background(), c); err != nil {
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

	g := &models.AdGroup{AdCampaignID: 1}
	gRepo.EXPECT().Create(gomock.Any(), g).Return(nil)
	if _, err := svc.CreateAdGroup(context.Background(), g); err != nil {
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
