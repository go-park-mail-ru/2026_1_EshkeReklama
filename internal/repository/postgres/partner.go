package postgres

import (
	"database/sql"
)

type PartnerRepository struct {
	db *sql.DB
}

func NewPartnerRepository(db *sql.DB) *PartnerRepository {
	return &PartnerRepository{db: db}
}
