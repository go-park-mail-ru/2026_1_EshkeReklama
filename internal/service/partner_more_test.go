package service

import (
	"context"
	"errors"
	errs "eshkere/internal/errors"
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
	"testing"
	"time"
)

type stubPartnerRepo struct {
	createFunc func(context.Context, *models.Partner) error
	getFunc    func(context.Context, int) (*models.Partner, error)
	updateFunc func(context.Context, *models.Partner) error
}

func (s *stubPartnerRepo) CreateProfile(ctx context.Context, partner *models.Partner) error {
	return s.createFunc(ctx, partner)
}
func (s *stubPartnerRepo) GetByID(ctx context.Context, id int) (*models.Partner, error) {
	return s.getFunc(ctx, id)
}
func (s *stubPartnerRepo) Update(ctx context.Context, partner *models.Partner) error {
	return s.updateFunc(ctx, partner)
}
func (s *stubPartnerRepo) SettleDailyEarnings(context.Context, time.Time) (int64, error) {
	return 0, nil
}

type stubPartnerSiteRepo struct {
	createFunc func(context.Context, *models.PartnerSite) error
	getFunc    func(context.Context, int) (*models.PartnerSite, error)
	listFunc   func(context.Context, int) ([]*models.PartnerSite, error)
	updateFunc func(context.Context, *models.PartnerSite) error
	deleteFunc func(context.Context, int) error
	existsFunc func(context.Context, string) (bool, error)
}

func (s *stubPartnerSiteRepo) Create(ctx context.Context, site *models.PartnerSite) error {
	return s.createFunc(ctx, site)
}
func (s *stubPartnerSiteRepo) GetByID(ctx context.Context, siteID int) (*models.PartnerSite, error) {
	return s.getFunc(ctx, siteID)
}
func (s *stubPartnerSiteRepo) ListByPartnerID(ctx context.Context, partnerID int) ([]*models.PartnerSite, error) {
	return s.listFunc(ctx, partnerID)
}
func (s *stubPartnerSiteRepo) Update(ctx context.Context, site *models.PartnerSite) error {
	return s.updateFunc(ctx, site)
}
func (s *stubPartnerSiteRepo) Delete(ctx context.Context, siteID int) error {
	return s.deleteFunc(ctx, siteID)
}
func (s *stubPartnerSiteRepo) ExistsByDomain(ctx context.Context, domain string) (bool, error) {
	return s.existsFunc(ctx, domain)
}

type stubPartnerBlockRepo struct {
	createFunc func(context.Context, *models.PartnerBlock) error
	getFunc    func(context.Context, int) (*models.PartnerBlock, error)
	listFunc   func(context.Context, int) ([]*models.PartnerBlock, error)
	updateFunc func(context.Context, *models.PartnerBlock) error
	deleteFunc func(context.Context, int) error
}

func (s *stubPartnerBlockRepo) Create(ctx context.Context, block *models.PartnerBlock) error {
	return s.createFunc(ctx, block)
}
func (s *stubPartnerBlockRepo) GetByID(ctx context.Context, blockID int) (*models.PartnerBlock, error) {
	return s.getFunc(ctx, blockID)
}
func (s *stubPartnerBlockRepo) ListBySiteID(ctx context.Context, siteID int) ([]*models.PartnerBlock, error) {
	return s.listFunc(ctx, siteID)
}
func (s *stubPartnerBlockRepo) Update(ctx context.Context, block *models.PartnerBlock) error {
	return s.updateFunc(ctx, block)
}
func (s *stubPartnerBlockRepo) Delete(ctx context.Context, blockID int) error {
	return s.deleteFunc(ctx, blockID)
}
func (s *stubPartnerBlockRepo) GetByEmbedToken(context.Context, string) (*models.PartnerBlock, error) {
	return nil, nil
}

type stubGeoRuleRepo struct {
	listFunc    func(context.Context, int) ([]*models.PartnerBlockGeoRule, error)
	replaceFunc func(context.Context, int, []*models.PartnerBlockGeoRule) error
}

func (s *stubGeoRuleRepo) ListByBlockID(ctx context.Context, blockID int) ([]*models.PartnerBlockGeoRule, error) {
	return s.listFunc(ctx, blockID)
}
func (s *stubGeoRuleRepo) ReplaceByBlockID(ctx context.Context, blockID int, rules []*models.PartnerBlockGeoRule) error {
	return s.replaceFunc(ctx, blockID, rules)
}

func TestPartnerHelpersAndServiceSetup(t *testing.T) {
	if _, err := NewService(nil); err == nil {
		t.Fatal("expected error for nil config")
	}
	svc, err := NewService(&Config{})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	client := &fakeProfileClient{}
	svc.SetProfileClient(client)
	if svc.profileClient != client {
		t.Fatal("expected profile client to be set")
	}

	if domain, err := normalizeDomain(" Example.com "); err != nil || domain != "example.com" {
		t.Fatalf("normalizeDomain: %q %v", domain, err)
	}
	if _, err := normalizeDomain("localhost"); err == nil {
		t.Fatal("expected invalid domain")
	}

	token := generatePartnerEmbedToken("pb")
	if len(token) == 0 || token[:3] != "pb_" {
		t.Fatalf("unexpected token: %s", token)
	}
	if got := defaultPlatformsForBlockType(models.PartnerBlockTypeBanner); len(got) != 3 {
		t.Fatalf("unexpected banner platforms: %v", got)
	}
	if got := defaultPlatformsForBlockType(models.PartnerBlockTypeTopAd); len(got) != 1 || got[0] != "mobile" {
		t.Fatalf("unexpected top_ad platforms: %v", got)
	}
	block := defaultPartnerBlock(models.PartnerBlockTypeBanner)
	if block.Status != models.PartnerBlockStatusInactive || block.RevenueShareBPS != 7000 || len(block.SelfAdSettings) == 0 {
		t.Fatalf("unexpected default partner block: %#v", block)
	}
}

func TestPartnerProfileFlows(t *testing.T) {
	var created *models.Partner
	partnerRepo := &stubPartnerRepo{
		createFunc: func(_ context.Context, partner *models.Partner) error {
			created = partner
			return nil
		},
		getFunc: func(_ context.Context, id int) (*models.Partner, error) {
			return &models.Partner{
				ID:                     id,
				LastName:               "Old",
				FirstName:              "Name",
				MiddleName:             "M",
				BirthDate:              time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
				CountryCode:            "RU",
				RegistrationRegionCode: "RU-MOW",
				CooperationForm:        models.CooperationFormSelfEmployed,
				PayoutCurrency:         models.PayoutCurrencyRUB,
			}, nil
		},
		updateFunc: func(context.Context, *models.Partner) error { return nil },
	}
	svc := &Service{partnerRepo: partnerRepo}

	err := svc.CreatePartnerProfile(context.Background(), &serviceinput.CreatePartnerProfile{
		ID:                     5,
		LastName:               " Ivanov ",
		FirstName:              " Ivan ",
		MiddleName:             " Ivanovich ",
		BirthDate:              "2000-01-02",
		CountryCode:            "RU",
		RegistrationRegionCode: "RU-MOW",
		CooperationForm:        string(models.CooperationFormSelfEmployed),
		PayoutCurrency:         string(models.PayoutCurrencyRUB),
	})
	if err != nil || created == nil || created.FirstName != "Ivan" {
		t.Fatalf("create partner profile: created=%#v err=%v", created, err)
	}
	if err := svc.CreatePartnerProfile(context.Background(), nil); !errors.Is(err, errs.BadRequestError) {
		t.Fatalf("expected bad request for nil input, got %v", err)
	}
	if _, err := svc.GetPartnerByID(context.Background(), 0); !errors.Is(err, errs.BadRequestError) {
		t.Fatalf("expected bad request for bad id, got %v", err)
	}

	firstName := "New"
	birthDate := "2001-02-03"
	updated, err := svc.UpdatePartnerProfile(context.Background(), &serviceinput.UpdatePartnerProfile{
		PartnerID: 5,
		FirstName: &firstName,
		BirthDate: &birthDate,
	})
	if err != nil || updated.FirstName != "New" || updated.BirthDate.Format("2006-01-02") != "2001-02-03" {
		t.Fatalf("update partner profile: %#v %v", updated, err)
	}
	if _, err := svc.UpdatePartnerProfile(context.Background(), nil); !errors.Is(err, errs.BadRequestError) {
		t.Fatalf("expected bad request for nil update, got %v", err)
	}
	badBirth := "bad"
	if _, err := svc.UpdatePartnerProfile(context.Background(), &serviceinput.UpdatePartnerProfile{PartnerID: 5, BirthDate: &badBirth}); !errors.Is(err, errs.BadRequestError) {
		t.Fatalf("expected bad request for bad birth date, got %v", err)
	}
}

func TestPartnerSiteFlows(t *testing.T) {
	var createdSite *models.PartnerSite
	siteRepo := &stubPartnerSiteRepo{
		createFunc: func(_ context.Context, site *models.PartnerSite) error {
			createdSite = site
			return nil
		},
		getFunc: func(_ context.Context, siteID int) (*models.PartnerSite, error) {
			return &models.PartnerSite{
				ID:        siteID,
				PartnerID: 8,
				Domain:    "example.com",
				SiteName:  "Example",
				Status:    models.PartnerSiteStatusDraft,
			}, nil
		},
		listFunc: func(context.Context, int) ([]*models.PartnerSite, error) {
			return []*models.PartnerSite{{ID: 1}, {ID: 2}}, nil
		},
		updateFunc: func(context.Context, *models.PartnerSite) error { return nil },
		deleteFunc: func(context.Context, int) error { return nil },
		existsFunc: func(_ context.Context, domain string) (bool, error) {
			return domain == "taken.example.com", nil
		},
	}
	svc := &Service{partnerSiteRepo: siteRepo}

	site, err := svc.CreatePartnerSite(context.Background(), &serviceinput.CreatePartnerSite{
		PartnerID: 8,
		Domain:    "https://Example.com",
		SiteName:  " Example ",
	})
	if err != nil || site.Domain != "example.com" || createdSite == nil {
		t.Fatalf("create partner site: %#v %v", site, err)
	}
	if _, err := svc.CreatePartnerSite(context.Background(), &serviceinput.CreatePartnerSite{PartnerID: 8, Domain: "taken.example.com", SiteName: "Name"}); !errors.Is(err, errs.AlreadyExistsError) {
		t.Fatalf("expected already exists, got %v", err)
	}
	if _, err := svc.CreatePartnerSite(context.Background(), &serviceinput.CreatePartnerSite{PartnerID: 8, Domain: "bad", SiteName: "Name"}); !errors.Is(err, errs.BadRequestError) {
		t.Fatalf("expected bad request, got %v", err)
	}

	newDomain := "new.example.com"
	newStatus := models.PartnerSiteStatusActive
	updated, err := svc.UpdatePartnerSite(context.Background(), &serviceinput.UpdatePartnerSite{
		ID:        1,
		PartnerID: 8,
		Domain:    &newDomain,
		Status:    &newStatus,
	})
	if err != nil || updated.Domain != "new.example.com" || updated.Status != models.PartnerSiteStatusActive {
		t.Fatalf("update partner site: %#v %v", updated, err)
	}
	if _, err := svc.UpdatePartnerSite(context.Background(), &serviceinput.UpdatePartnerSite{ID: 1, PartnerID: 7}); !errors.Is(err, errs.NotFoundError) {
		t.Fatalf("expected not found for foreign site, got %v", err)
	}
	invalidStatus := models.PartnerSiteStatus("broken")
	if _, err := svc.UpdatePartnerSite(context.Background(), &serviceinput.UpdatePartnerSite{ID: 1, PartnerID: 8, Status: &invalidStatus}); !errors.Is(err, errs.BadRequestError) {
		t.Fatalf("expected bad request for invalid status, got %v", err)
	}
	if sites, err := svc.ListPartnerSites(context.Background(), 8); err != nil || len(sites) != 2 {
		t.Fatalf("list partner sites: %v %v", sites, err)
	}
	if _, err := svc.GetPartnerSite(context.Background(), 0); !errors.Is(err, errs.BadRequestError) {
		t.Fatalf("expected bad request for invalid site id, got %v", err)
	}
	if err := svc.DeletePartnerSite(context.Background(), 8, 1); err != nil {
		t.Fatalf("delete partner site: %v", err)
	}
	if err := svc.DeletePartnerSite(context.Background(), 9, 1); !errors.Is(err, errs.NotFoundError) {
		t.Fatalf("expected not found on delete, got %v", err)
	}
}

func TestPartnerBlockFlows(t *testing.T) {
	siteRepo := &stubPartnerSiteRepo{
		getFunc: func(_ context.Context, siteID int) (*models.PartnerSite, error) {
			return &models.PartnerSite{ID: siteID, PartnerID: 8}, nil
		},
		createFunc: func(context.Context, *models.PartnerSite) error { return nil },
		listFunc:   func(context.Context, int) ([]*models.PartnerSite, error) { return nil, nil },
		updateFunc: func(context.Context, *models.PartnerSite) error { return nil },
		deleteFunc: func(context.Context, int) error { return nil },
		existsFunc: func(context.Context, string) (bool, error) { return false, nil },
	}
	var createdBlock *models.PartnerBlock
	blockRepo := &stubPartnerBlockRepo{
		createFunc: func(_ context.Context, block *models.PartnerBlock) error {
			createdBlock = block
			return nil
		},
		getFunc: func(_ context.Context, blockID int) (*models.PartnerBlock, error) {
			return &models.PartnerBlock{ID: blockID, PartnerSiteID: 1, Name: "Block", Status: models.PartnerBlockStatusInactive}, nil
		},
		listFunc: func(context.Context, int) ([]*models.PartnerBlock, error) {
			return []*models.PartnerBlock{{ID: 1}}, nil
		},
		updateFunc: func(context.Context, *models.PartnerBlock) error { return nil },
		deleteFunc: func(context.Context, int) error { return nil },
	}
	geoRepo := &stubGeoRuleRepo{
		listFunc: func(context.Context, int) ([]*models.PartnerBlockGeoRule, error) {
			return []*models.PartnerBlockGeoRule{{GeoCode: "all"}}, nil
		},
		replaceFunc: func(context.Context, int, []*models.PartnerBlockGeoRule) error { return nil },
	}
	svc := &Service{
		partnerSiteRepo:         siteRepo,
		partnerBlockRepo:        blockRepo,
		partnerBlockGeoRuleRepo: geoRepo,
	}

	block, err := svc.CreatePartnerBlock(context.Background(), 8, &serviceinput.CreatePartnerBlock{
		PartnerSiteID: 1,
		Name:          " Main ",
		BlockType:     models.PartnerBlockTypeBanner,
	})
	if err != nil || block.Name != "Main" || createdBlock == nil {
		t.Fatalf("create partner block: %#v %v", block, err)
	}
	if _, err := svc.CreatePartnerBlock(context.Background(), 9, &serviceinput.CreatePartnerBlock{PartnerSiteID: 1, Name: "x", BlockType: models.PartnerBlockTypeBanner}); !errors.Is(err, errs.NotFoundError) {
		t.Fatalf("expected not found, got %v", err)
	}
	if _, _, err := svc.GetPartnerBlock(context.Background(), 8, 1, 2); err != nil {
		t.Fatalf("get partner block: %v", err)
	}
	if blocks, err := svc.ListPartnerBlocks(context.Background(), 8, 1); err != nil || len(blocks) != 1 {
		t.Fatalf("list partner blocks: %v %v", blocks, err)
	}

	name := "Updated"
	status := models.PartnerBlockStatusActive
	updated, err := svc.UpdatePartnerBlockMeta(context.Background(), 8, 1, &serviceinput.UpdatePartnerBlockMeta{ID: 2, Name: &name, Status: &status})
	if err != nil || updated.Name != "Updated" || updated.Status != models.PartnerBlockStatusActive {
		t.Fatalf("update partner block meta: %#v %v", updated, err)
	}
	invalidStatus := models.PartnerBlockStatus("bad")
	if _, err := svc.UpdatePartnerBlockMeta(context.Background(), 8, 1, &serviceinput.UpdatePartnerBlockMeta{ID: 2, Status: &invalidStatus}); !errors.Is(err, errs.BadRequestError) {
		t.Fatalf("expected bad request for invalid block status, got %v", err)
	}
}

func TestPartnerBlockAdvancedFlowsAndDictionaries(t *testing.T) {
	siteRepo := &stubPartnerSiteRepo{
		getFunc: func(_ context.Context, siteID int) (*models.PartnerSite, error) {
			return &models.PartnerSite{ID: siteID, PartnerID: 8}, nil
		},
		createFunc: func(context.Context, *models.PartnerSite) error { return nil },
		listFunc:   func(context.Context, int) ([]*models.PartnerSite, error) { return nil, nil },
		updateFunc: func(context.Context, *models.PartnerSite) error { return nil },
		deleteFunc: func(context.Context, int) error { return nil },
		existsFunc: func(context.Context, string) (bool, error) { return false, nil },
	}
	blockRepo := &stubPartnerBlockRepo{
		createFunc: func(context.Context, *models.PartnerBlock) error { return nil },
		getFunc: func(_ context.Context, blockID int) (*models.PartnerBlock, error) {
			return &models.PartnerBlock{
				ID:            blockID,
				PartnerSiteID: 1,
				BlockType:     models.PartnerBlockTypeBanner,
				Status:        models.PartnerBlockStatusInactive,
				Name:          "Block",
				EmbedToken:    "token-1",
			}, nil
		},
		listFunc:   func(context.Context, int) ([]*models.PartnerBlock, error) { return nil, nil },
		updateFunc: func(context.Context, *models.PartnerBlock) error { return nil },
		deleteFunc: func(context.Context, int) error { return nil },
	}
	var replacedRules []*models.PartnerBlockGeoRule
	geoRepo := &stubGeoRuleRepo{
		listFunc: func(context.Context, int) ([]*models.PartnerBlockGeoRule, error) {
			return []*models.PartnerBlockGeoRule{{GeoCode: "all"}}, nil
		},
		replaceFunc: func(_ context.Context, _ int, rules []*models.PartnerBlockGeoRule) error {
			replacedRules = rules
			return nil
		},
	}
	partnerRepo := &stubPartnerRepo{
		createFunc: func(context.Context, *models.Partner) error { return nil },
		getFunc:    func(context.Context, int) (*models.Partner, error) { return nil, nil },
		updateFunc: func(context.Context, *models.Partner) error { return nil },
	}
	svc := &Service{
		partnerRepo:             partnerRepo,
		partnerSiteRepo:         siteRepo,
		partnerBlockRepo:        blockRepo,
		partnerBlockGeoRuleRepo: geoRepo,
	}

	theme := models.ThemeModeDark
	amp := models.AmpModeEnabled
	interscroller := models.InterscrollerModeEnabled
	revenueShare := 8000
	color := " #000 "
	updatedBlock, err := svc.UpdatePartnerBlockGeneralSettings(context.Background(), 8, 1, &serviceinput.UpdatePartnerBlockGeneral{
		ID:                           2,
		Theme:                        &theme,
		AmpMode:                      &amp,
		InterscrollerMode:            &interscroller,
		InterscrollerBackgroundColor: &color,
		RevenueShareBPS:              &revenueShare,
	})
	if err != nil || updatedBlock.Theme != models.ThemeModeDark || updatedBlock.RevenueShareBPS != 8000 || !updatedBlock.InterscrollerBackgroundColor.Valid {
		t.Fatalf("update partner block general settings: %#v %v", updatedBlock, err)
	}

	badShare := 10001
	if _, err := svc.UpdatePartnerBlockGeneralSettings(context.Background(), 8, 1, &serviceinput.UpdatePartnerBlockGeneral{ID: 2, RevenueShareBPS: &badShare}); !errors.Is(err, errs.BadRequestError) {
		t.Fatalf("expected bad request for revenue share, got %v", err)
	}

	topAdRepo := &stubPartnerBlockRepo{
		createFunc: func(context.Context, *models.PartnerBlock) error { return nil },
		getFunc: func(_ context.Context, blockID int) (*models.PartnerBlock, error) {
			return &models.PartnerBlock{ID: blockID, PartnerSiteID: 1, BlockType: models.PartnerBlockTypeTopAd}, nil
		},
		listFunc:   func(context.Context, int) ([]*models.PartnerBlock, error) { return nil, nil },
		updateFunc: func(context.Context, *models.PartnerBlock) error { return nil },
		deleteFunc: func(context.Context, int) error { return nil },
	}
	svc.partnerBlockRepo = topAdRepo
	if _, err := svc.UpdatePartnerBlockGeneralSettings(context.Background(), 8, 1, &serviceinput.UpdatePartnerBlockGeneral{ID: 2, AmpMode: &amp}); !errors.Is(err, errs.BusinessLogicError) {
		t.Fatalf("expected business logic error for unsupported amp, got %v", err)
	}

	svc.partnerBlockRepo = blockRepo
	cpmv := int64(15)
	rules, err := svc.UpdatePartnerBlockGeographySettings(context.Background(), 8, 1, &serviceinput.UpdatePartnerBlockGeography{
		ID:         2,
		GlobalCPMV: &cpmv,
		Rules: []serviceinput.PartnerBlockGeoRuleInput{
			{GeoCode: "RU-MOW", IsEnabled: true, CPMV: &cpmv},
		},
	})
	if err != nil || len(rules) != 1 || len(replacedRules) != 2 {
		t.Fatalf("update geography settings: rules=%#v replaced=%#v err=%v", rules, replacedRules, err)
	}

	selfAd, err := svc.UpdatePartnerBlockSelfAdSettings(context.Background(), 8, 1, &serviceinput.UpdatePartnerBlockSelfAd{
		ID:       2,
		Settings: []byte(`{"reserved":false}`),
	})
	if err != nil || string(selfAd.SelfAdSettings) != `{"reserved":false}` {
		t.Fatalf("update self ad settings: %#v %v", selfAd, err)
	}
	selfAd, err = svc.UpdatePartnerBlockSelfAdSettings(context.Background(), 8, 1, &serviceinput.UpdatePartnerBlockSelfAd{ID: 2})
	if err != nil || len(selfAd.SelfAdSettings) == 0 {
		t.Fatalf("update self ad settings default: %#v %v", selfAd, err)
	}

	if err := svc.DeletePartnerBlock(context.Background(), 8, 1, 2); err != nil {
		t.Fatalf("delete partner block: %v", err)
	}
	settleDate := time.Date(2026, 5, 5, 15, 4, 0, 0, time.UTC)
	partnerRepo.getFunc = func(context.Context, int) (*models.Partner, error) { return nil, nil }
	partnerRepo.updateFunc = func(context.Context, *models.Partner) error { return nil }
	partnerRepo.createFunc = func(context.Context, *models.Partner) error { return nil }
	partnerRepoSettleCalled := false
	svc.partnerRepo = &stubPartnerRepo{
		createFunc: func(context.Context, *models.Partner) error { return nil },
		getFunc:    func(context.Context, int) (*models.Partner, error) { return nil, nil },
		updateFunc: func(context.Context, *models.Partner) error { return nil },
	}
	svc.partnerRepo = &stubPartnerRepo{
		createFunc: func(context.Context, *models.Partner) error { return nil },
		getFunc:    func(context.Context, int) (*models.Partner, error) { return nil, nil },
		updateFunc: func(context.Context, *models.Partner) error { return nil },
	}
	svc.partnerRepo = &stubPartnerRepo{
		createFunc: func(context.Context, *models.Partner) error { return nil },
		getFunc:    func(context.Context, int) (*models.Partner, error) { return nil, nil },
		updateFunc: func(context.Context, *models.Partner) error { return nil },
	}
	svc.partnerRepo = &stubPartnerRepo{
		createFunc: func(context.Context, *models.Partner) error { return nil },
		getFunc:    func(context.Context, int) (*models.Partner, error) { return nil, nil },
		updateFunc: func(context.Context, *models.Partner) error { return nil },
	}
	svc.partnerRepo = &stubPartnerRepo{
		createFunc: func(context.Context, *models.Partner) error { return nil },
		getFunc:    func(context.Context, int) (*models.Partner, error) { return nil, nil },
		updateFunc: func(context.Context, *models.Partner) error { return nil },
	}
	svc.partnerRepo = &stubPartnerRepo{
		createFunc: func(context.Context, *models.Partner) error { return nil },
		getFunc:    func(context.Context, int) (*models.Partner, error) { return nil, nil },
		updateFunc: func(context.Context, *models.Partner) error { return nil },
	}
	svc.partnerRepo = &stubPartnerRepo{
		createFunc: func(context.Context, *models.Partner) error { return nil },
		getFunc:    func(context.Context, int) (*models.Partner, error) { return nil, nil },
		updateFunc: func(context.Context, *models.Partner) error { return nil },
	}
	svc.partnerRepo = &stubPartnerRepo{
		createFunc: func(context.Context, *models.Partner) error { return nil },
		getFunc:    func(context.Context, int) (*models.Partner, error) { return nil, nil },
		updateFunc: func(context.Context, *models.Partner) error { return nil },
	}
	svc.partnerRepo = &stubPartnerRepo{
		createFunc: func(context.Context, *models.Partner) error { return nil },
		getFunc:    func(context.Context, int) (*models.Partner, error) { return nil, nil },
		updateFunc: func(context.Context, *models.Partner) error { return nil },
	}
	svc.partnerRepo = &stubPartnerRepo{
		createFunc: func(context.Context, *models.Partner) error { return nil },
		getFunc:    func(context.Context, int) (*models.Partner, error) { return nil, nil },
		updateFunc: func(context.Context, *models.Partner) error { return nil },
	}
	svc.partnerRepo = &stubPartnerRepoWithSettle{stubPartnerRepo: stubPartnerRepo{}, settleFunc: func(_ context.Context, date time.Time) (int64, error) {
		partnerRepoSettleCalled = true
		if !date.Equal(time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("unexpected settle date: %v", date)
		}
		return 55, nil
	}}
	if settled, err := svc.SettlePartnerDailyEarnings(context.Background(), settleDate); err != nil || settled != 55 || !partnerRepoSettleCalled {
		t.Fatalf("settle partner daily earnings: %d %v", settled, err)
	}

	token, html, err := svc.GetPartnerBlockEmbedCode(context.Background(), 8, 1, 2, "https://example.com/")
	if err != nil || token != "token-1" || html == "" {
		t.Fatalf("get embed code: %q %q %v", token, html, err)
	}

	if len(svc.ListPartnerCountries(context.Background())) == 0 ||
		len(svc.ListPartnerRegistrationRegions(context.Background(), "RU")) == 0 ||
		len(svc.ListPartnerRegistrationRegions(context.Background(), "XX")) != 0 ||
		len(svc.ListPartnerCooperationForms(context.Background())) == 0 ||
		len(svc.ListPartnerPayoutCurrencies(context.Background())) == 0 ||
		len(svc.ListPartnerBlockTypes(context.Background())) == 0 ||
		len(svc.GetPartnerGeoTree(context.Background())) == 0 {
		t.Fatal("expected non-empty partner dictionaries")
	}
	if !supportsAMP(models.PartnerBlockTypeBanner) || supportsAMP(models.PartnerBlockTypeTopAd) {
		t.Fatal("unexpected supportsAMP behavior")
	}
}

type stubPartnerRepoWithSettle struct {
	stubPartnerRepo
	settleFunc func(context.Context, time.Time) (int64, error)
}

func (s *stubPartnerRepoWithSettle) CreateProfile(ctx context.Context, partner *models.Partner) error {
	return s.stubPartnerRepo.CreateProfile(ctx, partner)
}
func (s *stubPartnerRepoWithSettle) GetByID(ctx context.Context, id int) (*models.Partner, error) {
	return s.stubPartnerRepo.GetByID(ctx, id)
}
func (s *stubPartnerRepoWithSettle) Update(ctx context.Context, partner *models.Partner) error {
	return s.stubPartnerRepo.Update(ctx, partner)
}
func (s *stubPartnerRepoWithSettle) SettleDailyEarnings(ctx context.Context, date time.Time) (int64, error) {
	return s.settleFunc(ctx, date)
}

func TestPartnerRepoErrorsBubble(t *testing.T) {
	wantErr := errors.New("repo down")
	svc := &Service{
		partnerRepo: &stubPartnerRepo{
			createFunc: func(context.Context, *models.Partner) error { return wantErr },
			getFunc:    func(context.Context, int) (*models.Partner, error) { return nil, wantErr },
			updateFunc: func(context.Context, *models.Partner) error { return wantErr },
		},
		partnerSiteRepo: &stubPartnerSiteRepo{
			createFunc: func(context.Context, *models.PartnerSite) error { return wantErr },
			getFunc:    func(context.Context, int) (*models.PartnerSite, error) { return nil, wantErr },
			listFunc:   func(context.Context, int) ([]*models.PartnerSite, error) { return nil, wantErr },
			updateFunc: func(context.Context, *models.PartnerSite) error { return wantErr },
			deleteFunc: func(context.Context, int) error { return wantErr },
			existsFunc: func(context.Context, string) (bool, error) { return false, wantErr },
		},
	}
	if err := svc.CreatePartnerProfile(context.Background(), &serviceinput.CreatePartnerProfile{
		ID:                     1,
		LastName:               "A",
		FirstName:              "B",
		BirthDate:              "2000-01-01",
		CountryCode:            "RU",
		RegistrationRegionCode: "RU-MOW",
		CooperationForm:        string(models.CooperationFormSelfEmployed),
		PayoutCurrency:         string(models.PayoutCurrencyRUB),
	}); !errors.Is(err, wantErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
	if _, err := svc.GetPartnerByID(context.Background(), 1); !errors.Is(err, wantErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
	if _, err := svc.ListPartnerSites(context.Background(), 1); !errors.Is(err, wantErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
	if _, err := svc.CreatePartnerSite(context.Background(), &serviceinput.CreatePartnerSite{PartnerID: 1, Domain: "example.com", SiteName: "Name"}); !errors.Is(err, wantErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
