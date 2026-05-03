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
	DailyBudget  int64        `db:"daily_budget"`
	CPMPrice     int64        `db:"cpm_price"`
	MainAction   string       `db:"main_action"`
	CreatedAt    time.Time    `db:"created_at"`
	UpdatedAt    sql.NullTime `db:"updated_at"`
}
