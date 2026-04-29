package models

import (
	"database/sql"
	"time"
)

type TariffType string

const (
	TariffTypeNoob    TariffType = "noob"
	TariffTypePro     TariffType = "pro"
	TariffTypeCheater TariffType = "cheater"
)

type Advertiser struct {
	ID        int            `db:"id"`
	Name      string         `db:"name"`
	Surname   string         `db:"surname"`
	AvatarURL sql.NullString `db:"avatar_url"`
	Balance   int64          `db:"balance"`
	Company   string         `db:"company"`
	City      string         `db:"city"`
	Tariff    TariffType     `db:"tariff"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt sql.NullTime   `db:"updated_at"`
}
