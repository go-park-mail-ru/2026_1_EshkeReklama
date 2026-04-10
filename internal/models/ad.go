package models

import (
	"database/sql"
	"time"
)

type AdStatus string

const (
	AdStatusTurnedOff     AdStatus = "turned_off"
	AdStatusModeration    AdStatus = "moderation"
	AdStatusWorking       AdStatus = "working"
	AdStatusRejected      AdStatus = "rejected"
	AdStatusNotEnoughMoney AdStatus = "not_enough_money"
)

type Ad struct {
	ID        int          `db:"id"`
	AdGroupID int          `db:"ad_group_id"`
	Status    AdStatus     `db:"status"`
	Title     string       `db:"title"`
	ShortDesc string       `db:"short_desc"`
	ImageURL  string       `db:"image_url"`
	TargetURL string       `db:"target_url"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt sql.NullTime `db:"updated_at"`
}
