package postgres

import (
	"database/sql"
)

type AdGroupRepository struct {
	db *sql.DB
}

func NewAdGroupRepository(db *sql.DB) *AdGroupRepository {
	return &AdGroupRepository{db: db}
}
