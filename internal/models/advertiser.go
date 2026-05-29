package models

import (
	"database/sql"
	"time"
)

type TariffType string
type AdvertiserRole string

const (
	TariffTypeBasic   TariffType = "basic"
	TariffTypePro     TariffType = "pro"
	TariffTypeCheater TariffType = "cheater"

	// Deprecated: используй TariffTypeBasic
	TariffTypeNoob TariffType = TariffTypeBasic
)

const (
	MaxCampaignsBasic   = 5
	MaxCampaignsPro     = 20
	MaxCampaignsCheater = 9999

	SubscriptionDuration = 30 * 24 * time.Hour // 1 месяц
	SubscriptionPriceRub = 3900
)

const (
	AdvertiserRoleUser  AdvertiserRole = "user"
	AdvertiserRoleAdmin AdvertiserRole = "admin"
)

type Advertiser struct {
	ID                      int            `db:"id"`
	Name                    string         `db:"name"`
	Surname                 sql.NullString `db:"surname"`
	AvatarURL               sql.NullString `db:"avatar_url"`
	Balance                 int64          `db:"balance"`
	Company                 sql.NullString `db:"company"`
	City                    sql.NullString `db:"city"`
	Tariff                  TariffType     `db:"tariff"`
	TariffExpiresAt         sql.NullTime   `db:"tariff_expires_at"`
	Role                    AdvertiserRole `db:"role"`
	SavedPaymentMethodID    sql.NullString `db:"saved_payment_method_id"`
	SavedPaymentMethodTitle sql.NullString `db:"saved_payment_method_title"`
	CreatedAt               time.Time      `db:"created_at"`
	UpdatedAt               sql.NullTime   `db:"updated_at"`
}

// IsProActive возвращает true, если у рекламодателя активна Pro-подписка.
func (a *Advertiser) IsProActive() bool {
	if a.Tariff == TariffTypeCheater {
		return true
	}
	if a.Tariff != TariffTypePro {
		return false
	}
	if !a.TariffExpiresAt.Valid {
		return false
	}
	return time.Now().Before(a.TariffExpiresAt.Time)
}

// MaxCampaigns возвращает максимальное количество активных кампаний для тарифа.
func (a *Advertiser) MaxCampaigns() int {
	switch a.Tariff {
	case TariffTypePro:
		if a.IsProActive() {
			return MaxCampaignsPro
		}
		return MaxCampaignsBasic
	case TariffTypeCheater:
		return MaxCampaignsCheater
	default:
		return MaxCampaignsBasic
	}
}
