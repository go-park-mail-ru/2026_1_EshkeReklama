package postgres

import (
	"database/sql"
)

type AdvertiserRepository struct {
	db *sql.DB
}

func NewAdvertiserRepository(db *sql.DB) *AdvertiserRepository {
	return &AdvertiserRepository{db: db}
}
