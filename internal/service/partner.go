package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	errs "eshkere/internal/errors"
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
	"fmt"
	"net/url"
	"strings"
	"time"
)

func (s *Service) CreatePartnerProfile(ctx context.Context, in *serviceinput.CreatePartnerProfile) error {
	if in == nil {
		return fmt.Errorf("%w: nil partner profile input", errs.BadRequestError)
	}

	birthDate, err := time.Parse("2006-01-02", strings.TrimSpace(in.BirthDate))
	if err != nil {
		return fmt.Errorf("%w: invalid birth_date", errs.BadRequestError)
	}

	partner := &models.Partner{
		ID:                     int(in.ID),
		LastName:               strings.TrimSpace(in.LastName),
		FirstName:              strings.TrimSpace(in.FirstName),
		MiddleName:             strings.TrimSpace(in.MiddleName),
		BirthDate:              birthDate,
		CountryCode:            strings.TrimSpace(in.CountryCode),
		RegistrationRegionCode: strings.TrimSpace(in.RegistrationRegionCode),
		CooperationForm:        models.CooperationForm(strings.TrimSpace(in.CooperationForm)),
		PayoutCurrency:         models.PayoutCurrency(strings.TrimSpace(in.PayoutCurrency)),
	}

	if partner.ID <= 0 || partner.LastName == "" || partner.FirstName == "" || partner.CountryCode == "" ||
		partner.RegistrationRegionCode == "" || partner.CooperationForm == "" || partner.PayoutCurrency == "" {
		return fmt.Errorf("%w: invalid partner profile fields", errs.BadRequestError)
	}

	return s.partnerRepo.CreateProfile(ctx, partner)
}

func (s *Service) GetPartnerByID(ctx context.Context, id int) (*models.Partner, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid partner id", errs.BadRequestError)
	}
	return s.partnerRepo.GetByID(ctx, id)
}

func (s *Service) UpdatePartnerProfile(ctx context.Context, in *serviceinput.UpdatePartnerProfile) (*models.Partner, error) {
	if in == nil || in.PartnerID <= 0 {
		return nil, fmt.Errorf("%w: invalid partner profile update", errs.BadRequestError)
	}

	partner, err := s.partnerRepo.GetByID(ctx, in.PartnerID)
	if err != nil {
		return nil, err
	}

	if in.LastName != nil {
		partner.LastName = strings.TrimSpace(*in.LastName)
	}
	if in.FirstName != nil {
		partner.FirstName = strings.TrimSpace(*in.FirstName)
	}
	if in.MiddleName != nil {
		partner.MiddleName = strings.TrimSpace(*in.MiddleName)
	}
	if in.BirthDate != nil {
		parsed, err := time.Parse("2006-01-02", strings.TrimSpace(*in.BirthDate))
		if err != nil {
			return nil, fmt.Errorf("%w: invalid birth_date", errs.BadRequestError)
		}
		partner.BirthDate = parsed
	}
	if in.CountryCode != nil {
		partner.CountryCode = strings.TrimSpace(*in.CountryCode)
	}
	if in.RegistrationRegionCode != nil {
		partner.RegistrationRegionCode = strings.TrimSpace(*in.RegistrationRegionCode)
	}
	if in.CooperationForm != nil {
		partner.CooperationForm = models.CooperationForm(strings.TrimSpace(*in.CooperationForm))
	}
	if in.PayoutCurrency != nil {
		partner.PayoutCurrency = models.PayoutCurrency(strings.TrimSpace(*in.PayoutCurrency))
	}

	if err := s.partnerRepo.Update(ctx, partner); err != nil {
		return nil, err
	}
	return partner, nil
}

func normalizeDomain(raw string) (string, error) {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" {
		return "", fmt.Errorf("empty domain")
	}
	if !strings.Contains(value, "://") {
		value = "https://" + value
	}
	u, err := url.Parse(value)
	if err != nil {
		return "", err
	}
	host := strings.ToLower(strings.TrimSpace(u.Hostname()))
	host = strings.TrimSuffix(host, ".")
	if host == "" || !strings.Contains(host, ".") {
		return "", fmt.Errorf("invalid domain")
	}
	return host, nil
}

func (s *Service) CreatePartnerSite(ctx context.Context, in *serviceinput.CreatePartnerSite) (*models.PartnerSite, error) {
	if in == nil || in.PartnerID <= 0 {
		return nil, fmt.Errorf("%w: invalid create partner site input", errs.BadRequestError)
	}

	domain, err := normalizeDomain(in.Domain)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid domain", errs.BadRequestError)
	}

	exists, err := s.partnerSiteRepo.ExistsByDomain(ctx, domain)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("%w: domain already exists", errs.AlreadyExistsError)
	}

	site := &models.PartnerSite{
		PartnerID: in.PartnerID,
		Domain:    domain,
		SiteName:  strings.TrimSpace(in.SiteName),
		Status:    models.PartnerSiteStatusDraft,
	}
	if site.SiteName == "" {
		return nil, fmt.Errorf("%w: empty site_name", errs.BadRequestError)
	}
	if err := s.partnerSiteRepo.Create(ctx, site); err != nil {
		return nil, err
	}
	return site, nil
}

func (s *Service) GetPartnerSite(ctx context.Context, siteID int) (*models.PartnerSite, error) {
	if siteID <= 0 {
		return nil, fmt.Errorf("%w: invalid site id", errs.BadRequestError)
	}
	return s.partnerSiteRepo.GetByID(ctx, siteID)
}

func (s *Service) ListPartnerSites(ctx context.Context, partnerID int) ([]*models.PartnerSite, error) {
	if partnerID <= 0 {
		return nil, fmt.Errorf("%w: invalid partner id", errs.BadRequestError)
	}
	return s.partnerSiteRepo.ListByPartnerID(ctx, partnerID)
}

func (s *Service) UpdatePartnerSite(ctx context.Context, in *serviceinput.UpdatePartnerSite) (*models.PartnerSite, error) {
	if in == nil || in.ID <= 0 || in.PartnerID <= 0 {
		return nil, fmt.Errorf("%w: invalid update partner site input", errs.BadRequestError)
	}

	site, err := s.partnerSiteRepo.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if site.PartnerID != in.PartnerID {
		return nil, errs.NotFoundError
	}

	if in.Domain != nil {
		domain, err := normalizeDomain(*in.Domain)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid domain", errs.BadRequestError)
		}
		if domain != site.Domain {
			exists, err := s.partnerSiteRepo.ExistsByDomain(ctx, domain)
			if err != nil {
				return nil, err
			}
			if exists {
				return nil, fmt.Errorf("%w: domain already exists", errs.AlreadyExistsError)
			}
		}
		site.Domain = domain
	}
	if in.SiteName != nil {
		site.SiteName = strings.TrimSpace(*in.SiteName)
	}
	if in.Status != nil {
		if !in.Status.IsValid() {
			return nil, fmt.Errorf("%w: invalid site status", errs.BadRequestError)
		}
		site.Status = *in.Status
	}

	if err := s.partnerSiteRepo.Update(ctx, site); err != nil {
		return nil, err
	}
	return site, nil
}

func (s *Service) DeletePartnerSite(ctx context.Context, partnerID, siteID int) error {
	site, err := s.partnerSiteRepo.GetByID(ctx, siteID)
	if err != nil {
		return err
	}
	if site.PartnerID != partnerID {
		return errs.NotFoundError
	}
	return s.partnerSiteRepo.Delete(ctx, siteID)
}

func generatePartnerEmbedToken(prefix string) string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return prefix + "_fallback"
	}
	return prefix + "_" + hex.EncodeToString(buf)
}

func defaultPlatformsForBlockType(blockType models.PartnerBlockType) []string {
	switch blockType {
	case models.PartnerBlockTypeBanner:
		return []string{"desktop", "mobile", "amp"}
	case models.PartnerBlockTypeTopAd:
		return []string{"mobile"}
	default:
		return []string{"desktop", "mobile"}
	}
}

func defaultPartnerBlock(blockType models.PartnerBlockType) *models.PartnerBlock {
	background := sql.NullString{}
	return &models.PartnerBlock{
		BlockType:                    blockType,
		Status:                       models.PartnerBlockStatusInactive,
		EmbedToken:                   generatePartnerEmbedToken("pb"),
		CPMStrategy:                  models.CPMStrategyMaxIncome,
		AmpMode:                      models.AmpModeDisabled,
		SizeMode:                     models.SizeModeAdaptive,
		BorderMode:                   models.BorderModeAuto,
		CornerMode:                   models.CornerModeAuto,
		Theme:                        models.ThemeModeLight,
		InterscrollerMode:            models.InterscrollerModeAuto,
		InterscrollerBackgroundColor: background,
		RevenueShareBPS:              7000,
		SelfAdSettings:               append(json.RawMessage(nil), defaultSelfAdSettings...),
	}
}

func (s *Service) CreatePartnerBlock(ctx context.Context, partnerID int, in *serviceinput.CreatePartnerBlock) (*models.PartnerBlock, error) {
	if in == nil || in.PartnerSiteID <= 0 {
		return nil, fmt.Errorf("%w: invalid create partner block input", errs.BadRequestError)
	}
	site, err := s.partnerSiteRepo.GetByID(ctx, in.PartnerSiteID)
	if err != nil {
		return nil, err
	}
	if site.PartnerID != partnerID {
		return nil, errs.NotFoundError
	}

	block := defaultPartnerBlock(in.BlockType)
	block.PartnerSiteID = in.PartnerSiteID
	block.Name = strings.TrimSpace(in.Name)
	if block.Name == "" {
		return nil, fmt.Errorf("%w: empty block name", errs.BadRequestError)
	}
	if err := s.partnerBlockRepo.Create(ctx, block); err != nil {
		return nil, err
	}
	return block, nil
}

func (s *Service) GetPartnerBlock(ctx context.Context, partnerID, siteID, blockID int) (*models.PartnerBlock, []*models.PartnerBlockGeoRule, error) {
	site, err := s.partnerSiteRepo.GetByID(ctx, siteID)
	if err != nil {
		return nil, nil, err
	}
	if site.PartnerID != partnerID {
		return nil, nil, errs.NotFoundError
	}

	block, err := s.partnerBlockRepo.GetByID(ctx, blockID)
	if err != nil {
		return nil, nil, err
	}
	if block.PartnerSiteID != siteID {
		return nil, nil, errs.NotFoundError
	}

	rules, err := s.partnerBlockGeoRuleRepo.ListByBlockID(ctx, blockID)
	if err != nil {
		return nil, nil, err
	}
	return block, rules, nil
}

func (s *Service) ListPartnerBlocks(ctx context.Context, partnerID, siteID int) ([]*models.PartnerBlock, error) {
	site, err := s.partnerSiteRepo.GetByID(ctx, siteID)
	if err != nil {
		return nil, err
	}
	if site.PartnerID != partnerID {
		return nil, errs.NotFoundError
	}
	return s.partnerBlockRepo.ListBySiteID(ctx, siteID)
}

func (s *Service) UpdatePartnerBlockMeta(ctx context.Context, partnerID, siteID int, in *serviceinput.UpdatePartnerBlockMeta) (*models.PartnerBlock, error) {
	block, _, err := s.GetPartnerBlock(ctx, partnerID, siteID, in.ID)
	if err != nil {
		return nil, err
	}
	if in.Name != nil {
		block.Name = strings.TrimSpace(*in.Name)
	}
	if in.Status != nil {
		if !in.Status.IsValid() {
			return nil, fmt.Errorf("%w: invalid block status", errs.BadRequestError)
		}
		block.Status = *in.Status
	}
	if err := s.partnerBlockRepo.Update(ctx, block); err != nil {
		return nil, err
	}
	return block, nil
}

func (s *Service) UpdatePartnerBlockGeneralSettings(ctx context.Context, partnerID, siteID int, in *serviceinput.UpdatePartnerBlockGeneral) (*models.PartnerBlock, error) {
	block, _, err := s.GetPartnerBlock(ctx, partnerID, siteID, in.ID)
	if err != nil {
		return nil, err
	}
	if in.CPMStrategy != nil {
		block.CPMStrategy = *in.CPMStrategy
	}
	if in.AmpMode != nil {
		if *in.AmpMode == models.AmpModeEnabled && !supportsAMP(block.BlockType) {
			return nil, fmt.Errorf("%w: amp not supported for block type", errs.BusinessLogicError)
		}
		block.AmpMode = *in.AmpMode
	}
	if in.SizeMode != nil {
		block.SizeMode = *in.SizeMode
	}
	if in.BorderMode != nil {
		block.BorderMode = *in.BorderMode
	}
	if in.CornerMode != nil {
		block.CornerMode = *in.CornerMode
	}
	if in.Theme != nil {
		block.Theme = *in.Theme
	}
	if in.InterscrollerMode != nil {
		block.InterscrollerMode = *in.InterscrollerMode
	}
	if in.InterscrollerBackgroundColor != nil {
		color := strings.TrimSpace(*in.InterscrollerBackgroundColor)
		if color == "" {
			block.InterscrollerBackgroundColor = sql.NullString{}
		} else {
			block.InterscrollerBackgroundColor = sql.NullString{String: color, Valid: true}
		}
	}
	if in.RevenueShareBPS != nil {
		if *in.RevenueShareBPS < 0 || *in.RevenueShareBPS > 10000 {
			return nil, fmt.Errorf("%w: invalid revenue share", errs.BadRequestError)
		}
		block.RevenueShareBPS = *in.RevenueShareBPS
	}
	if err := s.partnerBlockRepo.Update(ctx, block); err != nil {
		return nil, err
	}
	return block, nil
}

func supportsAMP(blockType models.PartnerBlockType) bool {
	return blockType == models.PartnerBlockTypeBanner
}

func (s *Service) UpdatePartnerBlockGeographySettings(ctx context.Context, partnerID, siteID int, in *serviceinput.UpdatePartnerBlockGeography) ([]*models.PartnerBlockGeoRule, error) {
	block, _, err := s.GetPartnerBlock(ctx, partnerID, siteID, in.ID)
	if err != nil {
		return nil, err
	}

	rules := make([]*models.PartnerBlockGeoRule, 0, len(in.Rules))
	if in.GlobalCPMV != nil {
		rules = append(rules, &models.PartnerBlockGeoRule{
			PartnerBlockID: block.ID,
			GeoCode:        "all",
			IsEnabled:      true,
			CPMV:           sql.NullInt64{Int64: *in.GlobalCPMV, Valid: true},
		})
	}
	for _, item := range in.Rules {
		rule := &models.PartnerBlockGeoRule{
			PartnerBlockID: block.ID,
			GeoCode:        strings.TrimSpace(item.GeoCode),
			IsEnabled:      item.IsEnabled,
			CPMV:           sql.NullInt64{},
		}
		if item.CPMV != nil {
			rule.CPMV = sql.NullInt64{Int64: *item.CPMV, Valid: true}
		}
		rules = append(rules, rule)
	}
	if err := s.partnerBlockGeoRuleRepo.ReplaceByBlockID(ctx, block.ID, rules); err != nil {
		return nil, err
	}
	return s.partnerBlockGeoRuleRepo.ListByBlockID(ctx, block.ID)
}

func (s *Service) UpdatePartnerBlockSelfAdSettings(ctx context.Context, partnerID, siteID int, in *serviceinput.UpdatePartnerBlockSelfAd) (*models.PartnerBlock, error) {
	block, _, err := s.GetPartnerBlock(ctx, partnerID, siteID, in.ID)
	if err != nil {
		return nil, err
	}
	if len(in.Settings) == 0 {
		block.SelfAdSettings = append(json.RawMessage(nil), defaultSelfAdSettings...)
	} else {
		block.SelfAdSettings = append(json.RawMessage(nil), in.Settings...)
	}
	if err := s.partnerBlockRepo.Update(ctx, block); err != nil {
		return nil, err
	}
	return block, nil
}

func (s *Service) DeletePartnerBlock(ctx context.Context, partnerID, siteID, blockID int) error {
	block, _, err := s.GetPartnerBlock(ctx, partnerID, siteID, blockID)
	if err != nil {
		return err
	}
	return s.partnerBlockRepo.Delete(ctx, block.ID)
}

func (s *Service) SettlePartnerDailyEarnings(ctx context.Context, earningDate time.Time) (int64, error) {
	return s.partnerRepo.SettleDailyEarnings(ctx, dateOnly(earningDate.UTC()))
}

func (s *Service) GetPartnerBlockEmbedCode(ctx context.Context, partnerID, siteID, blockID int, baseURL string) (string, string, string, error) {
	block, _, err := s.GetPartnerBlock(ctx, partnerID, siteID, blockID)
	if err != nil {
		return "", "", "", err
	}
	base := strings.TrimRight(baseURL, "/")
	iframeURL := fmt.Sprintf("%s/public/partner/blocks/%s/frame", base, block.EmbedToken)
	htmlSnippet := fmt.Sprintf(`<iframe src="%s" width="300" height="250" style="border:0;overflow:hidden" loading="lazy" referrerpolicy="strict-origin-when-cross-origin"></iframe>`, iframeURL)
	return block.EmbedToken, iframeURL, htmlSnippet, nil
}

func (s *Service) ListPartnerCountries(_ context.Context) []DictionaryItem {
	return []DictionaryItem{
		{Code: "RU", Name: "Россия"},
		{Code: "KZ", Name: "Казахстан"},
		{Code: "BY", Name: "Беларусь"},
	}
}

func (s *Service) ListPartnerRegistrationRegions(_ context.Context, countryCode string) []DictionaryItem {
	switch strings.ToUpper(strings.TrimSpace(countryCode)) {
	case "RU":
		return []DictionaryItem{
			{Code: "RU-MOW", Name: "Москва"},
			{Code: "RU-SPE", Name: "Санкт-Петербург"},
			{Code: "RU-KDA", Name: "Краснодарский край"},
		}
	default:
		return []DictionaryItem{}
	}
}

func (s *Service) ListPartnerCooperationForms(_ context.Context) []DictionaryItem {
	return []DictionaryItem{
		{Code: string(models.CooperationFormSelfEmployed), Name: "Самозанятый"},
		{Code: string(models.CooperationFormIndividualEntrepreneur), Name: "ИП"},
		{Code: string(models.CooperationFormLegalEntity), Name: "Юрлицо"},
	}
}

func (s *Service) ListPartnerPayoutCurrencies(_ context.Context) []DictionaryItem {
	return []DictionaryItem{
		{Code: string(models.PayoutCurrencyRUB), Name: "Российский рубль"},
		{Code: string(models.PayoutCurrencyUSD), Name: "US Dollar"},
		{Code: string(models.PayoutCurrencyEUR), Name: "Euro"},
	}
}

func (s *Service) ListPartnerBlockTypes(_ context.Context) []BlockTypeDictionaryItem {
	return []BlockTypeDictionaryItem{
		{Code: string(models.PartnerBlockTypeBanner), Name: "Баннер", Description: "Универсальный рекламный блок для размещения в выделенном контейнере на странице", Platforms: defaultPlatformsForBlockType(models.PartnerBlockTypeBanner)},
		{Code: string(models.PartnerBlockTypeFullscreen), Name: "Полноэкранный", Description: "Рекламный блок отображается на весь экран и перекрывает содержимое страницы", Platforms: defaultPlatformsForBlockType(models.PartnerBlockTypeFullscreen)},
		{Code: string(models.PartnerBlockTypeFloorAd), Name: "Floor Ad", Description: "Рекламный блок фиксируется в нижней части экрана поверх контента и занимает всю ширину", Platforms: defaultPlatformsForBlockType(models.PartnerBlockTypeFloorAd)},
		{Code: string(models.PartnerBlockTypeTopAd), Name: "Top Ad", Description: "Мобильный рекламный блок фиксируется в верхней части экрана поверх контента и занимает всю ширину", Platforms: defaultPlatformsForBlockType(models.PartnerBlockTypeTopAd)},
		{Code: string(models.PartnerBlockTypeFeed), Name: "Лента", Description: "Рекламный блок в формате ленты объявлений, размещается под основным контентом страницы", Platforms: defaultPlatformsForBlockType(models.PartnerBlockTypeFeed)},
		{Code: string(models.PartnerBlockTypeInImage), Name: "In-Image", Description: "Рекламный блок фиксируется поверх изображений на сайте", Platforms: defaultPlatformsForBlockType(models.PartnerBlockTypeInImage)},
	}
}

func (s *Service) GetPartnerGeoTree(_ context.Context) []*GeoTreeNode {
	return []*GeoTreeNode{
		{Code: "all", Name: "Все регионы"},
		{Code: "cis", Name: "СНГ"},
		{Code: "europe", Name: "Европа"},
		{Code: "asia", Name: "Азия"},
		{Code: "africa", Name: "Африка"},
		{Code: "north_america", Name: "Северная Америка"},
		{Code: "south_america", Name: "Южная Америка"},
		{Code: "australia_oceania", Name: "Австралия и Океания"},
	}
}
