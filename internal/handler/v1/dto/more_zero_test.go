package dto

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"eshkere/internal/models"
)

func TestPartnerAndSiteDTOConversions(t *testing.T) {
	register := (&PartnerRegisterRequest{
		LastName:               "Ivanov",
		FirstName:              "Ivan",
		MiddleName:             "Ivanovich",
		BirthDate:              "2000-01-01",
		Email:                  "user@example.com",
		Phone:                  "9001234567",
		CountryCode:            "RU",
		RegistrationRegionCode: "RU-MOW",
		CooperationForm:        "self_employed",
		PayoutCurrency:         "RUB",
	}).ToInput(3)
	if register.ID != 3 || register.FirstName != "Ivan" {
		t.Fatalf("unexpected register input: %#v", register)
	}

	status := "active"
	domain := "example.com"
	siteName := "Example"
	updateSite := (&UpdatePartnerSiteRequest{Domain: &domain, SiteName: &siteName, Status: &status}).ToInput(7, 11)
	if updateSite.PartnerID != 7 || updateSite.ID != 11 || updateSite.Status == nil || *updateSite.Status != models.PartnerSiteStatusActive {
		t.Fatalf("unexpected update site input: %#v", updateSite)
	}

	createdAt := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)
	site := &models.PartnerSite{
		ID:        11,
		PartnerID: 7,
		Domain:    "example.com",
		SiteName:  "Example",
		Status:    models.PartnerSiteStatusActive,
		CreatedAt: createdAt,
		UpdatedAt: sql.NullTime{Time: updatedAt, Valid: true},
	}
	resp := ToPartnerSiteResponse(site)
	if resp == nil || resp.Status != "active" || resp.UpdatedAt == "" {
		t.Fatalf("unexpected site response: %#v", resp)
	}
	if ToPartnerSiteResponse(nil) != nil {
		t.Fatal("expected nil response for nil site")
	}
	if len(ToPartnerSiteResponses([]*models.PartnerSite{site})) != 1 {
		t.Fatal("expected one site response")
	}

	partner := &models.Partner{
		ID:                     7,
		LastName:               "Ivanov",
		FirstName:              "Ivan",
		MiddleName:             "Ivanovich",
		BirthDate:              createdAt,
		CountryCode:            "RU",
		RegistrationRegionCode: "RU-MOW",
		CooperationForm:        models.CooperationFormSelfEmployed,
		PayoutCurrency:         models.PayoutCurrencyRUB,
		Balance:                100,
		CreatedAt:              createdAt,
	}
	profile := PartnerToProfile(partner, "user@example.com", "9001234567")
	if profile.Email != "user@example.com" || profile.Phone != "9001234567" || profile.ID != 7 {
		t.Fatalf("unexpected partner profile: %#v", profile)
	}
}

func TestPartnerBlockDTOConversions(t *testing.T) {
	createReq := (&CreatePartnerBlockRequest{BlockType: "banner", Name: "Main"}).ToInput(5)
	if createReq.PartnerSiteID != 5 || createReq.BlockType != models.PartnerBlockTypeBanner {
		t.Fatalf("unexpected create input: %#v", createReq)
	}

	name := "Updated"
	status := "active"
	metaInput := (&UpdatePartnerBlockMetaRequest{Name: &name, Status: &status}).ToInput(8)
	if metaInput.ID != 8 || metaInput.Status == nil || *metaInput.Status != models.PartnerBlockStatusActive {
		t.Fatalf("unexpected meta input: %#v", metaInput)
	}

	theme := "dark"
	revenueShare := 7000
	generalInput := (&UpdatePartnerBlockGeneralRequest{Theme: &theme, RevenueShareBPS: &revenueShare}).ToInput(8)
	if generalInput.ID != 8 || generalInput.Theme == nil || *generalInput.Theme != models.ThemeModeDark || *generalInput.RevenueShareBPS != 7000 {
		t.Fatalf("unexpected general input: %#v", generalInput)
	}

	cpmv := int64(10)
	geoInput := (&UpdatePartnerBlockGeographyRequest{
		OnlyConfigured: true,
		GlobalCPMV:     &cpmv,
		Rules:          []UpdatePartnerBlockGeographyRuleRequest{{GeoCode: "RU-MOW", IsEnabled: true, CPMV: &cpmv}},
	}).ToInput(8)
	if geoInput.ID != 8 || len(geoInput.Rules) != 1 || geoInput.Rules[0].GeoCode != "RU-MOW" {
		t.Fatalf("unexpected geo input: %#v", geoInput)
	}

	selfAdInput := (&UpdatePartnerBlockSelfAdRequest{Reserved: true}).ToInput(8)
	if selfAdInput.ID != 8 || !json.Valid(selfAdInput.Settings) {
		t.Fatalf("unexpected self ad input: %#v", selfAdInput)
	}

	block := &models.PartnerBlock{
		ID:                           8,
		PartnerSiteID:                5,
		Name:                         "Main",
		BlockType:                    models.PartnerBlockTypeBanner,
		Status:                       models.PartnerBlockStatusActive,
		CPMStrategy:                  models.CPMStrategyMaxIncome,
		AmpMode:                      models.AmpModeDisabled,
		SizeMode:                     models.SizeModeAdaptive,
		BorderMode:                   models.BorderModeAuto,
		CornerMode:                   models.CornerModeRounded,
		Theme:                        models.ThemeModeLight,
		InterscrollerMode:            models.InterscrollerModeAuto,
		InterscrollerBackgroundColor: sql.NullString{String: "#fff", Valid: true},
		RevenueShareBPS:              7000,
		SelfAdSettings:               json.RawMessage(`{"reserved":false}`),
		CreatedAt:                    time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC),
		UpdatedAt:                    sql.NullTime{Time: time.Date(2026, 5, 5, 13, 0, 0, 0, time.UTC), Valid: true},
	}
	rules := []*models.PartnerBlockGeoRule{{GeoCode: "all", CPMV: sql.NullInt64{Int64: 15, Valid: true}, IsEnabled: true}}

	resp := ToPartnerBlockResponse(block)
	if resp == nil || resp.BlockType != "banner" || resp.UpdatedAt == "" {
		t.Fatalf("unexpected block response: %#v", resp)
	}
	if ToPartnerBlockResponse(nil) != nil {
		t.Fatal("expected nil block response")
	}

	details := ToPartnerBlockDetailsResponse(block, rules, false, &cpmv, []string{"desktop"})
	if details == nil || details.GeographySettings.GlobalCPMV == nil || *details.GeographySettings.GlobalCPMV != cpmv {
		t.Fatalf("unexpected details: %#v", details)
	}
	if toNullableString(sql.NullString{}) != nil || toNullableInt64(sql.NullInt64{}) != nil {
		t.Fatal("expected nil nullable conversions")
	}
}

func TestFormDTOHelpers(t *testing.T) {
	create := NewCreateAdRequestFromForm(map[string][]string{
		"title":      {"Title"},
		"short_desc": {"Short"},
		"target_url": {"https://example.com"},
	})
	if create.Title != "Title" || create.TargetURL != "https://example.com" {
		t.Fatalf("unexpected create ad request: %#v", create)
	}

	update := NewUpdateAdRequestFromForm(map[string][]string{
		"title":      {"Title"},
		"status":     {"working"},
		"short_desc": {"Short"},
		"target_url": {"https://example.com"},
	})
	if update.Title == nil || *update.Title != "Title" || update.Status == nil || *update.Status != "working" {
		t.Fatalf("unexpected update ad request: %#v", update)
	}
	if optionalStringFromForm(map[string][]string{}, "missing") != nil {
		t.Fatal("expected nil optional form value")
	}
	if update.ToInput(4).ID != 4 {
		t.Fatal("expected ad update input to include id")
	}

	appeals := ToAppealResponses([]*models.Appeal{{ID: 1}, {ID: 2}})
	if len(appeals) != 2 {
		t.Fatalf("unexpected appeal response count: %d", len(appeals))
	}
	if ToListAppealsResponse(7, []*models.Appeal{{ID: 1}}).AdvertiserID != 7 {
		t.Fatal("expected advertiser id in appeals list response")
	}
}
