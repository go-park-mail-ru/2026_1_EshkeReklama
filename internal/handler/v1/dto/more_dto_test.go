package dto

import (
	"net/url"
	"testing"
	"time"

	"eshkere/internal/models"
)

func TestAdditionalDTOConverters(t *testing.T) {
	name := "campaign"
	budget := int64(50)
	price := int64(12)
	action := "buy"
	campaignUpdate := (&UpdateAdCampaignRequest{
		Name:        &name,
		DailyBudget: &budget,
		CPMPrice:    &price,
		MainAction:  &action,
	}).ToInput(17)
	if campaignUpdate.ID != 17 || *campaignUpdate.Name != "campaign" || *campaignUpdate.MainAction != "buy" {
		t.Fatalf("unexpected campaign update input: %+v", campaignUpdate)
	}

	topic := "finance"
	region := "msk"
	groupName := "group"
	ageFrom := 21
	ageTo := 45
	gender := "woman"
	groupUpdate := (&UpdateAdGroupRequest{
		Topic:   &topic,
		Region:  &region,
		Name:    &groupName,
		AgeFrom: &ageFrom,
		AgeTo:   &ageTo,
		Gender:  &gender,
	}).ToInput(19)
	if groupUpdate.ID != 19 || *groupUpdate.Name != "group" || string(*groupUpdate.Gender) != "woman" {
		t.Fatalf("unexpected group update input: %+v", groupUpdate)
	}

	adminResp := ToListAdminAdsResponse([]*models.Ad{{ID: 3, Title: "ad"}})
	if len(adminResp.Ads) != 1 || adminResp.Ads[0].ID != 3 {
		t.Fatalf("unexpected admin ads response: %+v", adminResp)
	}

	formReq := NewCreateAppealRequestFromForm(url.Values{
		"category":    {"question"},
		"title":       {"Need help"},
		"description": {"body"},
		"name":        {"Ivan"},
		"email":       {"ivan@example.com"},
	})
	if formReq.Category != "question" || formReq.Title != "Need help" || formReq.Email != "ivan@example.com" {
		t.Fatalf("unexpected appeal form request: %+v", formReq)
	}

	lastName := "Ivanov"
	firstName := "Ivan"
	updatePartner := (&UpdatePartnerProfileRequest{
		LastName:  &lastName,
		FirstName: &firstName,
	}).ToInput(23)
	if updatePartner.PartnerID != 23 || *updatePartner.LastName != "Ivanov" || *updatePartner.FirstName != "Ivan" {
		t.Fatalf("unexpected partner update input: %+v", updatePartner)
	}

	partnerDate := time.Date(2026, 2, 3, 0, 0, 0, 0, time.UTC)
	partnerCreated := time.Date(2026, 2, 4, 5, 6, 7, 0, time.UTC)
	partnerProfile := PartnerToProfile(&models.Partner{
		ID:                     4,
		LastName:               "Ivanov",
		FirstName:              "Ivan",
		MiddleName:             "Ivanovich",
		BirthDate:              partnerDate,
		CountryCode:            "RU",
		RegistrationRegionCode: "MSK",
		CooperationForm:        models.CooperationFormSelfEmployed,
		PayoutCurrency:         models.PayoutCurrencyRUB,
		Balance:                99,
		CreatedAt:              partnerCreated,
	}, "partner@example.com", "+7999")
	if partnerProfile.ID != 4 || partnerProfile.BirthDate != "2026-02-03" || partnerProfile.Email != "partner@example.com" {
		t.Fatalf("unexpected partner profile: %+v", partnerProfile)
	}

	createSite := (&CreatePartnerSiteRequest{Domain: "example.com", SiteName: "Example"}).ToInput(8)
	if createSite.PartnerID != 8 || createSite.Domain != "example.com" {
		t.Fatalf("unexpected create site input: %+v", createSite)
	}

	status := "active"
	domain := "new.example.com"
	siteName := "Renamed"
	updateSite := (&UpdatePartnerSiteRequest{Domain: &domain, SiteName: &siteName, Status: &status}).ToInput(8, 13)
	if updateSite.PartnerID != 8 || updateSite.ID != 13 || string(*updateSite.Status) != "active" {
		t.Fatalf("unexpected update site input: %+v", updateSite)
	}
}
