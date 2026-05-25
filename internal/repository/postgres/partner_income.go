package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"eshkere/internal/service"
)

type PartnerIncomeRepository struct {
	db *sql.DB
}

func NewPartnerIncomeRepository(db *sql.DB) *PartnerIncomeRepository {
	return &PartnerIncomeRepository{db: db}
}

const selectPartnerIncome = `SELECT
	e.earning_date,
	ps.id,
	COALESCE(ps.site_name, ''),
	ps.domain,
	pb.id,
	pb.name,
	COALESCE(SUM(e.impressions), 0),
	COALESCE(SUM(e.reward), 0)
FROM eshkere.partner_block_daily_earning e
JOIN eshkere.partner_block pb ON pb.id = e.partner_block_id
JOIN eshkere.partner_site ps ON ps.id = pb.partner_site_id
WHERE ps.partner_id = $1
	AND e.earning_date >= $2
	AND e.earning_date <= $3
GROUP BY e.earning_date, ps.id, ps.site_name, ps.domain, pb.id, pb.name
ORDER BY e.earning_date DESC, ps.id ASC, pb.id ASC`

func (r *PartnerIncomeRepository) ListPartnerIncome(ctx context.Context, partnerID int, from, to time.Time) ([]service.PartnerIncomeRow, error) {
	rows, err := r.db.QueryContext(ctx, selectPartnerIncome, partnerID, from, to)
	if err != nil {
		return nil, fmt.Errorf("query partner income: %w", err)
	}
	defer rows.Close()

	incomeRows := make([]service.PartnerIncomeRow, 0)
	for rows.Next() {
		var row service.PartnerIncomeRow
		if err = rows.Scan(
			&row.Date,
			&row.SiteID,
			&row.SiteName,
			&row.Domain,
			&row.BlockID,
			&row.BlockName,
			&row.Impressions,
			&row.Reward,
		); err != nil {
			return nil, fmt.Errorf("scan partner income: %w", err)
		}
		incomeRows = append(incomeRows, row)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate partner income: %w", err)
	}

	return incomeRows, nil
}
