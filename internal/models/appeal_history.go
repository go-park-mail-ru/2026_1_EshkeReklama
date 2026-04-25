package models

import "time"

type AppealMessageAuthor string

const (
	AppealMessageAuthorAdmin      AppealMessageAuthor = "admin"
	AppealMessageAuthorAdvertiser AppealMessageAuthor = "advertiser"
)

type AppealMessage struct {
	ID        int                 `db:"id"`
	AppealID  int                 `db:"appeal_id"`
	Author    AppealMessageAuthor `db:"author"`
	Text      string              `db:"text"`
	CreatedAt time.Time           `db:"created_at"`
}

type AppealStatusHistory struct {
	ID        int          `db:"id"`
	AppealID  int          `db:"appeal_id"`
	Status    AppealStatus `db:"status"`
	CreatedAt time.Time    `db:"created_at"`
}
