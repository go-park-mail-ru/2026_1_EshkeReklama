package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"eshkere/internal/models"
)

type PartnerSiteRepository struct {
	db *sql.DB
}

func NewPartnerSiteRepository(db *sql.DB) *PartnerSiteRepository {
	return &PartnerSiteRepository{db: db}
}

const (
	insertPartnerSite = `INSERT INTO eshkere.partner_site (
		partner_id, domain, site_name, status, created_at
	) VALUES ($1, $2, $3, $4, NOW()) RETURNING id`

	selectPartnerSiteByID = `SELECT
		id, partner_id, domain, site_name, status, created_at, updated_at
	FROM eshkere.partner_site
	WHERE id = $1`

	selectPartnerSitesByPartnerID = `SELECT
		id, partner_id, domain, site_name, status, created_at, updated_at
	FROM eshkere.partner_site
	WHERE partner_id = $1
	ORDER BY created_at DESC, id DESC`

	updatePartnerSite = `UPDATE eshkere.partner_site SET
		domain = $1, site_name = $2, status = $3
	WHERE id = $4`

	deletePartnerSite               = `DELETE FROM eshkere.partner_site WHERE id = $1`
	selectPartnerSiteExistsByDomain = `SELECT EXISTS(SELECT 1 FROM eshkere.partner_site WHERE domain = $1)`
)

func (r *PartnerSiteRepository) Create(ctx context.Context, site *models.PartnerSite) error {
	if site == nil {
		return fmt.Errorf("partner site cannot be nil")
	}
	if err := r.db.QueryRowContext(ctx, insertPartnerSite, site.PartnerID, site.Domain, site.SiteName, site.Status).Scan(&site.ID); err != nil {
		return fmt.Errorf("insert partner site: %w", err)
	}
	return nil
}

func (r *PartnerSiteRepository) GetByID(ctx context.Context, siteID int) (*models.PartnerSite, error) {
	var site models.PartnerSite
	err := r.db.QueryRowContext(ctx, selectPartnerSiteByID, siteID).Scan(
		&site.ID,
		&site.PartnerID,
		&site.Domain,
		&site.SiteName,
		&site.Status,
		&site.CreatedAt,
		&site.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("partner site not found: %w", err)
		}
		return nil, fmt.Errorf("get partner site by id: %w", err)
	}
	return &site, nil
}

func (r *PartnerSiteRepository) ListByPartnerID(ctx context.Context, partnerID int) ([]*models.PartnerSite, error) {
	rows, err := r.db.QueryContext(ctx, selectPartnerSitesByPartnerID, partnerID)
	if err != nil {
		return nil, fmt.Errorf("list partner sites: %w", err)
	}
	defer rows.Close()

	sites := make([]*models.PartnerSite, 0)
	for rows.Next() {
		var site models.PartnerSite
		if err := rows.Scan(&site.ID, &site.PartnerID, &site.Domain, &site.SiteName, &site.Status, &site.CreatedAt, &site.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan partner site: %w", err)
		}
		sites = append(sites, &site)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate partner sites: %w", err)
	}
	return sites, nil
}

func (r *PartnerSiteRepository) Update(ctx context.Context, site *models.PartnerSite) error {
	if site == nil {
		return fmt.Errorf("partner site cannot be nil")
	}
	_, err := r.db.ExecContext(ctx, updatePartnerSite, site.Domain, site.SiteName, site.Status, site.ID)
	if err != nil {
		return fmt.Errorf("update partner site: %w", err)
	}
	return nil
}

func (r *PartnerSiteRepository) Delete(ctx context.Context, siteID int) error {
	result, err := r.db.ExecContext(ctx, deletePartnerSite, siteID)
	if err != nil {
		return fmt.Errorf("delete partner site: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete partner site rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("partner site not found")
	}
	return nil
}

func (r *PartnerSiteRepository) ExistsByDomain(ctx context.Context, domain string) (bool, error) {
	var exists bool
	if err := r.db.QueryRowContext(ctx, selectPartnerSiteExistsByDomain, domain).Scan(&exists); err != nil {
		return false, fmt.Errorf("exists partner site by domain: %w", err)
	}
	return exists, nil
}
