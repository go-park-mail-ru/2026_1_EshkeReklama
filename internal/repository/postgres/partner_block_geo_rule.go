package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"eshkere/internal/models"
)

type PartnerBlockGeoRuleRepository struct {
	db *sql.DB
}

func NewPartnerBlockGeoRuleRepository(db *sql.DB) *PartnerBlockGeoRuleRepository {
	return &PartnerBlockGeoRuleRepository{db: db}
}

const (
	selectPartnerBlockGeoRulesByBlockID = `SELECT
		id, partner_block_id, geo_code, is_enabled, cpmv, created_at, updated_at
	FROM eshkere.partner_block_geo_rule
	WHERE partner_block_id = $1
	ORDER BY id ASC`

	deletePartnerBlockGeoRulesByBlockID = `DELETE FROM eshkere.partner_block_geo_rule WHERE partner_block_id = $1`
	insertPartnerBlockGeoRule           = `INSERT INTO eshkere.partner_block_geo_rule (
		partner_block_id, geo_code, is_enabled, cpmv, created_at
	) VALUES ($1, $2, $3, $4, NOW())`
)

func (r *PartnerBlockGeoRuleRepository) ListByBlockID(ctx context.Context, blockID int) ([]*models.PartnerBlockGeoRule, error) {
	rows, err := r.db.QueryContext(ctx, selectPartnerBlockGeoRulesByBlockID, blockID)
	if err != nil {
		return nil, fmt.Errorf("list partner block geo rules: %w", err)
	}
	defer rows.Close()

	rules := make([]*models.PartnerBlockGeoRule, 0)
	for rows.Next() {
		var rule models.PartnerBlockGeoRule
		if err := rows.Scan(&rule.ID, &rule.PartnerBlockID, &rule.GeoCode, &rule.IsEnabled, &rule.CPMV, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan partner block geo rule: %w", err)
		}
		rules = append(rules, &rule)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate partner block geo rules: %w", err)
	}
	return rules, nil
}

func (r *PartnerBlockGeoRuleRepository) ReplaceByBlockID(ctx context.Context, blockID int, rules []*models.PartnerBlockGeoRule) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin replace partner block geo rules: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, deletePartnerBlockGeoRulesByBlockID, blockID); err != nil {
		return fmt.Errorf("delete partner block geo rules: %w", err)
	}

	for _, rule := range rules {
		if rule == nil {
			continue
		}
		if _, err := tx.ExecContext(ctx, insertPartnerBlockGeoRule, blockID, rule.GeoCode, rule.IsEnabled, rule.CPMV); err != nil {
			return fmt.Errorf("insert partner block geo rule: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit replace partner block geo rules: %w", err)
	}
	return nil
}
