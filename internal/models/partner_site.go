package models

import (
	"database/sql"
	"time"
)

type PartnerSiteStatus string

const (
	PartnerSiteStatusDraft         PartnerSiteStatus = "draft"
	PartnerSiteStatusPendingReview PartnerSiteStatus = "pending_review"
	PartnerSiteStatusActive        PartnerSiteStatus = "active"
	PartnerSiteStatusRejected      PartnerSiteStatus = "rejected"
	PartnerSiteStatusBlocked       PartnerSiteStatus = "blocked"
	PartnerSiteStatusArchived      PartnerSiteStatus = "archived"
)

func (s PartnerSiteStatus) IsValid() bool {
	switch s {
	case PartnerSiteStatusDraft,
		PartnerSiteStatusPendingReview,
		PartnerSiteStatusActive,
		PartnerSiteStatusRejected,
		PartnerSiteStatusBlocked,
		PartnerSiteStatusArchived:
		return true
	default:
		return false
	}
}

type PartnerSite struct {
	ID        int               `db:"id"`
	PartnerID int               `db:"partner_id"`
	Domain    string            `db:"domain"`
	SiteName  string            `db:"site_name"`
	Status    PartnerSiteStatus `db:"status"`
	CreatedAt time.Time         `db:"created_at"`
	UpdatedAt sql.NullTime      `db:"updated_at"`
}
