package postgres

import (
	"database/sql"
)

type AdActionRepository struct {
	db *sql.DB
}

func NewAdActionRepository(db *sql.DB) *AdActionRepository {
	return &AdActionRepository{db: db}
}
