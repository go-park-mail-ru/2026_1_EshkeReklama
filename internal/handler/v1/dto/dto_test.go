package dto

import (
	"testing"
	"time"

	"eshkere/internal/models"
)

func TestToModel_AdCampaign(t *testing.T) {
	req := CreateAdCampaignRequest{Name: "n"}
	in := req.ToInput(7)
	if in.AdvertiserID != 7 || in.Name != "n" {
		t.Fatalf("unexpected input: %+v", in)
	}
}

func TestToAdResponse(t *testing.T) {
	ad := &models.Ad{ID: 1, Status: models.AdStatusWorking, Title: "t", ShortDesc: "s", ImageURL: "i", TargetURL: "u"}
	resp := ToAdResponse(ad)
	if resp.ID != 1 || resp.Title != "t" || resp.TargetURL != "u" {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

func TestToListAdsResponse(t *testing.T) {
	ads := []*models.Ad{{ID: 1, Title: "t"}}
	resp := ToListAdsResponse(7, ads)
	if resp.GroupID != 7 || len(resp.Ads) != 1 || resp.Ads[0].ID != 1 {
		t.Fatalf("unexpected list ads response: %+v", resp)
	}
}

func TestAdvertiserToProfile_NilAndNonNil(t *testing.T) {
	empty := AdvertiserToProfile(nil)
	if empty.ID != 0 {
		t.Fatalf("expected empty")
	}

	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	adv := &models.Advertiser{ID: 1, Name: "n", Balance: 10, CreatedAt: now}
	p := AdvertiserWithContactsToProfile(adv, "e", "p")
	if p.ID != 1 || p.Email != "e" || p.Phone != "p" || p.CreatedAt != now.Format(time.RFC3339) {
		t.Fatalf("unexpected profile: %+v", p)
	}
}

func TestToModel_AdGroup(t *testing.T) {
	req := CreateAdGroupRequest{TopicID: 1, RegionID: 2, Name: "g", AgeFrom: 18, AgeTo: 25, Gender: "any"}
	in := req.ToInput(10)
	if in.AdCampaignID != 10 || in.TopicID != 1 || in.Gender != "any" {
		t.Fatalf("unexpected input: %+v", in)
	}
}

func TestToAdCampaignResponse(t *testing.T) {
	c := &models.AdCampaign{ID: 1, Status: models.AdStatusWorking, Name: "n", DailyBudget: 10}
	resp := ToAdCampaignResponse(c)
	if resp.ID != 1 || resp.Name != "n" {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

func TestToListAdCampaignsResponse(t *testing.T) {
	campaigns := []*models.AdCampaign{{ID: 1, Name: "n"}}
	resp := ToListAdCampaignsResponse(5, campaigns)
	if resp.AdvertiserID != 5 || len(resp.Campaigns) != 1 || resp.Campaigns[0].ID != 1 {
		t.Fatalf("unexpected list campaigns response: %+v", resp)
	}
}

func TestToAdGroupResponse(t *testing.T) {
	g := &models.AdGroup{ID: 1, TopicID: 2, RegionID: 3, Name: "n", AgeFrom: 1, AgeTo: 2, Gender: "any"}
	resp := ToAdGroupResponse(g)
	if resp.ID != 1 || resp.TopicID != 2 {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

func TestToListAdGroupsResponse(t *testing.T) {
	groups := []*models.AdGroup{{ID: 1, TopicID: 2}}
	resp := ToListAdGroupsResponse(9, groups)
	if resp.AdCampaignID != 9 || len(resp.Groups) != 1 || resp.Groups[0].ID != 1 {
		t.Fatalf("unexpected list groups response: %+v", resp)
	}
}

func TestCreateAdRequest_ToInput(t *testing.T) {
	req := CreateAdRequest{Title: "t", ShortDesc: "s", TargetURL: "u"}
	in := req.ToInput(3)
	if in.AdGroupID != 3 || in.Title != "t" || in.TargetURL != "u" {
		t.Fatalf("unexpected input: %+v", in)
	}
}

func TestUpdateAdvertiserProfileRequest_ToInput(t *testing.T) {
	name := "n"
	email := "e"
	phone := "p"
	req := UpdateAdvertiserProfileRequest{Name: &name, Email: &email, Phone: &phone}
	in := req.ToInput(7)
	if in.AdvertiserID != 7 || *in.Name != "n" {
		t.Fatalf("unexpected input: %+v", in)
	}
}

func TestToFeedResponse(t *testing.T) {
	ads := []*models.Ad{{ID: 1, Title: "t"}}
	resp := ToFeedResponse(ads)
	if len(resp.Ads) != 1 || resp.Ads[0].ID != 1 {
		t.Fatalf("unexpected feed response: %+v", resp)
	}
}

func TestToAppealResponse_WithImageURL(t *testing.T) {
	appeal := &models.Appeal{ID: 3, Status: models.AppealStatusOpen, Category: models.AppealCategoryBug, Title: "Crash", Description: "Steps", ImageURL: "https://cdn.example.com/file.png"}
	resp := ToAppealResponse(appeal)
	if resp.ID != 3 || resp.ImageURL != "https://cdn.example.com/file.png" {
		t.Fatalf("unexpected appeal response: %+v", resp)
	}
}

func TestCreateAppealRequest_ToInput_WithImage(t *testing.T) {
	req := CreateAppealRequest{
		Category:    "bug",
		Title:       "Crash",
		Description: "Steps",
		Name:        "Ivan",
		Email:       "ivan@example.com",
	}

	in := req.ToInput(&UploadedImage{
		Data:        []byte("png"),
		Ext:         ".png",
		ContentType: "image/png",
	})
	if in.Category != models.AppealCategoryBug || in.Title != "Crash" || in.Email != "ivan@example.com" {
		t.Fatalf("unexpected appeal input: %+v", in)
	}
	if string(in.Image) != "png" || in.ImageExt != ".png" || in.ImageType != "image/png" {
		t.Fatalf("unexpected image input: %+v", in)
	}
}
