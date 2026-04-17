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
	Title        string       `db:"title"`
	ShortDesc    string       `db:"short_desc"`
	ImageURL     string       `db:"image_url"`
	TargetURL    string       `db:"target_url"`
	CreatedAt    time.Time    `db:"created_at"`
	UpdatedAt    sql.NullTime `db:"updated_at"`
}
