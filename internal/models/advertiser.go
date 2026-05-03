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
	Surname   sql.NullString `db:"surname"`
	AvatarURL sql.NullString `db:"avatar_url"`
	Balance   int64          `db:"balance"`
	Company   sql.NullString `db:"company"`
	City      sql.NullString `db:"city"`
	Tariff    TariffType     `db:"tariff"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt sql.NullTime   `db:"updated_at"`
}
