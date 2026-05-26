package models

import (
	"database/sql"
	"time"
)

type GenderType string

const (
	GenderMan   GenderType = "man"
	GenderWoman GenderType = "woman"
	GenderAny   GenderType = "any"
)

type AdGroup struct {
	ID           int          `db:"id"`
	AdCampaignID int          `db:"ad_campaign_id"`
	TopicID      int          `db:"topic_id"`
	RegionID     int          `db:"region_id"`
	Name         string       `db:"name"`
	AgeFrom      int          `db:"age_from"`
	AgeTo        int          `db:"age_to"`
	Gender       GenderType   `db:"gender"`
	CreatedAt    time.Time    `db:"created_at"`
	UpdatedAt    sql.NullTime `db:"updated_at"`
}
