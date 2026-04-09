package postgres

import (
	"database/sql"
)

type AdRepository struct {
	db *sql.DB
}

func NewAdRepository(db *sql.DB) *AdRepository {
	return &AdRepository{db: db}
}
