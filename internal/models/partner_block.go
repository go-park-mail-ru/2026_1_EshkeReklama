package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

type PartnerBlockType string
type PartnerBlockStatus string
type CPMStrategy string
type AmpMode string
type SizeMode string
type BorderMode string
type CornerMode string
type ThemeMode string
type InterscrollerMode string

const (
	PartnerBlockTypeBanner     PartnerBlockType = "banner"
	PartnerBlockTypeFullscreen PartnerBlockType = "fullscreen"
	PartnerBlockTypeFloorAd    PartnerBlockType = "floor_ad"
	PartnerBlockTypeTopAd      PartnerBlockType = "top_ad"
	PartnerBlockTypeFeed       PartnerBlockType = "feed"
	PartnerBlockTypeInImage    PartnerBlockType = "in_image"

	PartnerBlockStatusDraft    PartnerBlockStatus = "draft"
	PartnerBlockStatusActive   PartnerBlockStatus = "active"
	PartnerBlockStatusArchived PartnerBlockStatus = "archived"

	CPMStrategyMaxIncome CPMStrategy = "max_income"

	AmpModeDisabled AmpMode = "disabled"
	AmpModeEnabled  AmpMode = "enabled"

	SizeModeAdaptive SizeMode = "adaptive"

	BorderModeAuto     BorderMode = "auto"
	BorderModeEnabled  BorderMode = "enabled"
	BorderModeDisabled BorderMode = "disabled"

	CornerModeAuto    CornerMode = "auto"
	CornerModeRounded CornerMode = "rounded"
	CornerModeSquare  CornerMode = "square"

	ThemeModeLight ThemeMode = "light"
	ThemeModeDark  ThemeMode = "dark"

	InterscrollerModeAuto     InterscrollerMode = "auto"
	InterscrollerModeEnabled  InterscrollerMode = "enabled"
	InterscrollerModeDisabled InterscrollerMode = "disabled"
)

type PartnerBlock struct {
	ID                           int                `db:"id"`
	PartnerSiteID                int                `db:"partner_site_id"`
	Name                         string             `db:"name"`
	BlockType                    PartnerBlockType   `db:"block_type"`
	Status                       PartnerBlockStatus `db:"status"`
	EmbedToken                   string             `db:"embed_token"`
	CPMStrategy                  CPMStrategy        `db:"cpm_strategy"`
	AmpMode                      AmpMode            `db:"amp_mode"`
	SizeMode                     SizeMode           `db:"size_mode"`
	BorderMode                   BorderMode         `db:"border_mode"`
	CornerMode                   CornerMode         `db:"corner_mode"`
	Theme                        ThemeMode          `db:"theme"`
	InterscrollerMode            InterscrollerMode  `db:"interscroller_mode"`
	InterscrollerBackgroundColor sql.NullString     `db:"interscroller_background_color"`
	RevenueShareBPS              int                `db:"revenue_share_bps"`
	SelfAdSettings               json.RawMessage    `db:"self_ad_settings"`
	CreatedAt                    time.Time          `db:"created_at"`
	UpdatedAt                    sql.NullTime       `db:"updated_at"`
}
