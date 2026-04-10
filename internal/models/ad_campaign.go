package models

import (
	"database/sql"
	"time"
)

type AdCampaign struct {
	ID           int          `db:"id"`
	AdvertiserID int          `db:"advertiser_id"`
	Status       AdStatus     `db:"status"`
	Name         string       `db:"name"`
	DailyBudget  float64      `db:"daily_budget"`
	CreatedAt    time.Time    `db:"created_at"`
	UpdatedAt    sql.NullTime `db:"updated_at"`
}
