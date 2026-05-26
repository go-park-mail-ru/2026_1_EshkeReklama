package models

import (
	"database/sql"
	"time"
)

type PartnerBlockGeoRule struct {
	ID             int           `db:"id"`
	PartnerBlockID int           `db:"partner_block_id"`
	GeoCode        string        `db:"geo_code"`
	IsEnabled      bool          `db:"is_enabled"`
	CPMV           sql.NullInt64 `db:"cpmv"`
	CreatedAt      time.Time     `db:"created_at"`
	UpdatedAt      sql.NullTime  `db:"updated_at"`
}
