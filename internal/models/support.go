package models

import (
	"database/sql"
	"time"
)

type SupportMessageAuthor string

const (
	SupportMessageAuthorAdvertiser SupportMessageAuthor = "advertiser"
	SupportMessageAuthorModerator  SupportMessageAuthor = "moderator"
	SupportMessageAuthorSystem     SupportMessageAuthor = "system"
)

type SupportThread struct {
	ID           int          `db:"id"`
	CampaignID   int          `db:"campaign_id"`
	AdvertiserID int          `db:"advertiser_id"`
	CreatedAt    time.Time    `db:"created_at"`
	UpdatedAt    sql.NullTime `db:"updated_at"`
}

type SupportMessage struct {
	ID         int                  `db:"id"`
	ThreadID   int                  `db:"thread_id"`
	AuthorType SupportMessageAuthor `db:"author_type"`
	AuthorID   sql.NullInt64        `db:"author_id"`
	Text       string               `db:"text"`
	CreatedAt  time.Time            `db:"created_at"`
}

type SupportThreadSummary struct {
	Thread       *SupportThread
	LastMessage  *SupportMessage
	LastActivity time.Time
}
