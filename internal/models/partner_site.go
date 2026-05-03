package models

import (
	"database/sql"
	"time"
)

type PartnerSite struct {
	ID        int          `db:"id"`
	PartnerID int          `db:"partner_id"`
	Domain    string       `db:"domain"`
	SiteName  string       `db:"site_name"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt sql.NullTime `db:"updated_at"`
}
