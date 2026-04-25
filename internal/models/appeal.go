package models

import (
	"database/sql"
	"time"
)

type AppealStatus string
type AppealCategory string

const (
	AppealStatusOpen       AppealStatus = "open"
	AppealStatusInProgress AppealStatus = "in_progress"
	AppealStatusClosed     AppealStatus = "closed"

	AppealCategoryBug        AppealCategory = "bug"
	AppealCategorySuggestion AppealCategory = "suggestion"
	AppealCategoryComplaint  AppealCategory = "complaint"
	AppealCategoryQuestion   AppealCategory = "question"
)

type Appeal struct {
	ID           int            `db:"id"`
	AdvertiserID sql.NullInt64  `db:"advertiser_id"`
	Status       AppealStatus   `db:"status"`
	Category     AppealCategory `db:"category"`
	Title        string         `db:"title"`
	Description  string         `db:"description"`
	ImageURL     string         `db:"image_url"`
	Name         string         `db:"name"`
	Email        string         `db:"email"`
	CreatedAt    time.Time      `db:"created_at"`
	UpdatedAt    sql.NullTime   `db:"updated_at"`
}
