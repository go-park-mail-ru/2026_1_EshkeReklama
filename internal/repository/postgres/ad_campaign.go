package postgres

import (
	"database/sql"
)

type AdCampaignRepository struct {
	db *sql.DB
}

func NewAdCampaignRepository(db *sql.DB) *AdCampaignRepository {
	return &AdCampaignRepository{db: db}
}
