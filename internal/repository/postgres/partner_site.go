package postgres

import (
	"database/sql"
)

type PartnerSiteRepository struct {
	db *sql.DB
}

func NewPartnerSiteRepository(db *sql.DB) *PartnerSiteRepository {
	return &PartnerSiteRepository{db: db}
}
