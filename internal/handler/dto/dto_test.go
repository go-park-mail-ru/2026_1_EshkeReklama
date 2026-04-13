package dto

import (
	"testing"
	"time"

	"eshkere/internal/models"
)

func TestToModel_AdCampaign(t *testing.T) {
	req := CreateAdCampaignRequest{Name: "n", DailyBudget: 10}
	m := req.ToModel(7)
	if m.AdvertiserID != 7 || m.Name != "n" || m.DailyBudget != 10 {
		t.Fatalf("unexpected model: %+v", m)
	}
	if m.Status != models.AdStatusModeration {
		t.Fatalf("expected moderation status")
	}
}

func TestToAdResponse(t *testing.T) {
	ad := &models.Ad{ID: 1, Status: models.AdStatusWorking, Title: "t", ShortDesc: "s", ImageURL: "i", TargetURL: "u"}
	resp := ToAdResponse(ad)
	if resp.ID != 1 || resp.Title != "t" || resp.TargetURL != "u" {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

func TestAdvertiserToProfile_NilAndNonNil(t *testing.T) {
	empty := AdvertiserToProfile(nil)
	if empty.ID != 0 {
		t.Fatalf("expected empty")
	}

	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	adv := &models.Advertiser{ID: 1, Name: "n", Email: "e", Phone: "p", Balance: 10, CreatedAt: now}
	p := AdvertiserToProfile(adv)
	if p.ID != 1 || p.CreatedAt != now.Format(time.RFC3339) {
		t.Fatalf("unexpected profile: %+v", p)
	}
}

func TestToModel_AdGroup(t *testing.T) {
	req := CreateAdGroupRequest{TopicID: 1, RegionID: 2, Name: "g", AgeFrom: 18, AgeTo: 25, Gender: "any"}
	m := req.ToModel(10)
	if m.AdCampaignID != 10 || m.TopicID != 1 || m.Gender != "any" {
		t.Fatalf("unexpected model: %+v", m)
	}
}

func TestToAdCampaignResponse(t *testing.T) {
	c := &models.AdCampaign{ID: 1, Status: models.AdStatusWorking, Name: "n", DailyBudget: 10}
	resp := ToAdCampaignResponse(c)
	if resp.ID != 1 || resp.Name != "n" {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

func TestToAdGroupResponse(t *testing.T) {
	g := &models.AdGroup{ID: 1, TopicID: 2, RegionID: 3, Name: "n", AgeFrom: 1, AgeTo: 2, Gender: "any"}
	resp := ToAdGroupResponse(g)
	if resp.ID != 1 || resp.TopicID != 2 {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

func TestCreateAdRequest_ToModel(t *testing.T) {
	req := CreateAdRequest{Title: "t", ShortDesc: "s", ImageURL: "i", TargetURL: "u"}
	ad, err := req.ToModel()
	if err != nil {
		t.Fatalf("ToModel: %v", err)
	}
	if ad.Title != "t" || ad.TargetURL != "u" {
		t.Fatalf("unexpected ad: %+v", ad)
	}
}
