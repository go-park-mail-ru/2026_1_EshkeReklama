package dto

import (
	"database/sql"
	"encoding/json"
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
	"time"
)

type CreatePartnerBlockRequest struct {
	BlockType string `json:"block_type" validate:"required,oneof=banner fullscreen floor_ad top_ad feed in_image"`
	Name      string `json:"name" validate:"required"`
}

func (r *CreatePartnerBlockRequest) ToInput(siteID int) *serviceinput.CreatePartnerBlock {
	return &serviceinput.CreatePartnerBlock{
		PartnerSiteID: siteID,
		Name:          r.Name,
		BlockType:     models.PartnerBlockType(r.BlockType),
	}
}

type UpdatePartnerBlockMetaRequest struct {
	Name *string `json:"name"`
}

func (r *UpdatePartnerBlockMetaRequest) ToInput(blockID int) *serviceinput.UpdatePartnerBlockMeta {
	return &serviceinput.UpdatePartnerBlockMeta{ID: blockID, Name: r.Name}
}

type UpdatePartnerBlockGeneralRequest struct {
	CPMStrategy                  *string `json:"cpm_strategy" validate:"omitempty,oneof=max_income"`
	AmpMode                      *string `json:"amp_mode" validate:"omitempty,oneof=disabled enabled"`
	SizeMode                     *string `json:"size_mode" validate:"omitempty,oneof=adaptive"`
	BorderMode                   *string `json:"border_mode" validate:"omitempty,oneof=auto enabled disabled"`
	CornerMode                   *string `json:"corner_mode" validate:"omitempty,oneof=auto rounded square"`
	Theme                        *string `json:"theme" validate:"omitempty,oneof=light dark"`
	InterscrollerMode            *string `json:"interscroller_mode" validate:"omitempty,oneof=auto enabled disabled"`
	InterscrollerBackgroundColor *string `json:"interscroller_background_color"`
}

func (r *UpdatePartnerBlockGeneralRequest) ToInput(blockID int) *serviceinput.UpdatePartnerBlockGeneral {
	input := &serviceinput.UpdatePartnerBlockGeneral{ID: blockID, InterscrollerBackgroundColor: r.InterscrollerBackgroundColor}
	if r.CPMStrategy != nil {
		v := models.CPMStrategy(*r.CPMStrategy)
		input.CPMStrategy = &v
	}
	if r.AmpMode != nil {
		v := models.AmpMode(*r.AmpMode)
		input.AmpMode = &v
	}
	if r.SizeMode != nil {
		v := models.SizeMode(*r.SizeMode)
		input.SizeMode = &v
	}
	if r.BorderMode != nil {
		v := models.BorderMode(*r.BorderMode)
		input.BorderMode = &v
	}
	if r.CornerMode != nil {
		v := models.CornerMode(*r.CornerMode)
		input.CornerMode = &v
	}
	if r.Theme != nil {
		v := models.ThemeMode(*r.Theme)
		input.Theme = &v
	}
	if r.InterscrollerMode != nil {
		v := models.InterscrollerMode(*r.InterscrollerMode)
		input.InterscrollerMode = &v
	}
	return input
}

type UpdatePartnerBlockGeographyRuleRequest struct {
	GeoCode   string `json:"geo_code" validate:"required"`
	IsEnabled bool   `json:"is_enabled"`
	CPMV      *int64 `json:"cpmv"`
}

type UpdatePartnerBlockGeographyRequest struct {
	OnlyConfigured bool                                     `json:"only_configured"`
	GlobalCPMV     *int64                                   `json:"global_cpmv"`
	Rules          []UpdatePartnerBlockGeographyRuleRequest `json:"rules"`
}

func (r *UpdatePartnerBlockGeographyRequest) ToInput(blockID int) *serviceinput.UpdatePartnerBlockGeography {
	rules := make([]serviceinput.PartnerBlockGeoRuleInput, 0, len(r.Rules))
	for _, rule := range r.Rules {
		rules = append(rules, serviceinput.PartnerBlockGeoRuleInput{
			GeoCode:   rule.GeoCode,
			IsEnabled: rule.IsEnabled,
			CPMV:      rule.CPMV,
		})
	}
	return &serviceinput.UpdatePartnerBlockGeography{
		ID:             blockID,
		OnlyConfigured: r.OnlyConfigured,
		GlobalCPMV:     r.GlobalCPMV,
		Rules:          rules,
	}
}

type UpdatePartnerBlockSelfAdRequest struct {
	Reserved bool `json:"reserved"`
}

func (r *UpdatePartnerBlockSelfAdRequest) ToInput(blockID int) *serviceinput.UpdatePartnerBlockSelfAd {
	payload, _ := json.Marshal(map[string]bool{"reserved": r.Reserved})
	return &serviceinput.UpdatePartnerBlockSelfAd{ID: blockID, Settings: payload}
}

type PartnerBlockResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	BlockType string `json:"block_type"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

func ToPartnerBlockResponse(block *models.PartnerBlock) *PartnerBlockResponse {
	if block == nil {
		return nil
	}
	resp := &PartnerBlockResponse{
		ID:        block.ID,
		Name:      block.Name,
		BlockType: string(block.BlockType),
		Status:    string(block.Status),
		CreatedAt: block.CreatedAt.Format(time.RFC3339),
	}
	if block.UpdatedAt.Valid {
		resp.UpdatedAt = block.UpdatedAt.Time.Format(time.RFC3339)
	}
	return resp
}

type ListPartnerBlocksResponse struct {
	SiteID int                     `json:"site_id"`
	Blocks []*PartnerBlockResponse `json:"blocks"`
}

type PartnerBlockGeneralSettingsResponse struct {
	CPMStrategy                  string  `json:"cpm_strategy"`
	AmpMode                      string  `json:"amp_mode"`
	SizeMode                     string  `json:"size_mode"`
	BorderMode                   string  `json:"border_mode"`
	CornerMode                   string  `json:"corner_mode"`
	Theme                        string  `json:"theme"`
	InterscrollerMode            string  `json:"interscroller_mode"`
	InterscrollerBackgroundColor *string `json:"interscroller_background_color"`
}

type PartnerBlockGeoRuleResponse struct {
	GeoCode   string `json:"geo_code"`
	IsEnabled bool   `json:"is_enabled"`
	CPMV      *int64 `json:"cpmv"`
}

type PartnerBlockGeographySettingsResponse struct {
	OnlyConfigured bool                           `json:"only_configured"`
	GlobalCPMV     *int64                         `json:"global_cpmv"`
	Rules          []*PartnerBlockGeoRuleResponse `json:"rules"`
}

type PartnerBlockSelfAdSettingsResponse struct {
	Reserved bool `json:"reserved"`
}

type PartnerBlockDetailsResponse struct {
	ID                 int                                   `json:"id"`
	SiteID             int                                   `json:"site_id"`
	Name               string                                `json:"name"`
	BlockType          string                                `json:"block_type"`
	Status             string                                `json:"status"`
	SupportedPlatforms []string                              `json:"supported_platforms"`
	GeneralSettings    PartnerBlockGeneralSettingsResponse   `json:"general_settings"`
	GeographySettings  PartnerBlockGeographySettingsResponse `json:"geography_settings"`
	SelfAdSettings     PartnerBlockSelfAdSettingsResponse    `json:"self_ad_settings"`
	CreatedAt          string                                `json:"created_at"`
	UpdatedAt          string                                `json:"updated_at,omitempty"`
}

func toNullableString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	v := value.String
	return &v
}

func toNullableInt64(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	v := value.Int64
	return &v
}

func ToPartnerBlockDetailsResponse(block *models.PartnerBlock, rules []*models.PartnerBlockGeoRule, onlyConfigured bool, globalCPMV *int64, supportedPlatforms []string) *PartnerBlockDetailsResponse {
	selfAd := PartnerBlockSelfAdSettingsResponse{Reserved: true}
	if len(block.SelfAdSettings) > 0 {
		_ = json.Unmarshal(block.SelfAdSettings, &selfAd)
	}
	respRules := make([]*PartnerBlockGeoRuleResponse, 0, len(rules))
	for _, rule := range rules {
		respRules = append(respRules, &PartnerBlockGeoRuleResponse{
			GeoCode:   rule.GeoCode,
			IsEnabled: rule.IsEnabled,
			CPMV:      toNullableInt64(rule.CPMV),
		})
	}
	resp := &PartnerBlockDetailsResponse{
		ID:                 block.ID,
		SiteID:             block.PartnerSiteID,
		Name:               block.Name,
		BlockType:          string(block.BlockType),
		Status:             string(block.Status),
		SupportedPlatforms: supportedPlatforms,
		GeneralSettings: PartnerBlockGeneralSettingsResponse{
			CPMStrategy:                  string(block.CPMStrategy),
			AmpMode:                      string(block.AmpMode),
			SizeMode:                     string(block.SizeMode),
			BorderMode:                   string(block.BorderMode),
			CornerMode:                   string(block.CornerMode),
			Theme:                        string(block.Theme),
			InterscrollerMode:            string(block.InterscrollerMode),
			InterscrollerBackgroundColor: toNullableString(block.InterscrollerBackgroundColor),
		},
		GeographySettings: PartnerBlockGeographySettingsResponse{
			OnlyConfigured: onlyConfigured,
			GlobalCPMV:     globalCPMV,
			Rules:          respRules,
		},
		SelfAdSettings: selfAd,
		CreatedAt:      block.CreatedAt.Format(time.RFC3339),
	}
	if block.UpdatedAt.Valid {
		resp.UpdatedAt = block.UpdatedAt.Time.Format(time.RFC3339)
	}
	return resp
}

type PartnerBlockEmbedResponse struct {
	BlockID     int    `json:"block_id"`
	EmbedToken  string `json:"embed_token"`
	ScriptURL   string `json:"script_url"`
	IframeURL   string `json:"iframe_url"`
	HTMLSnippet string `json:"html_snippet"`
}
