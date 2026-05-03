package postgres

import (
	"context"
	"database/sql"
	"errors"
	"eshkere/internal/models"
	"fmt"
)

type PartnerBlockRepository struct {
	db *sql.DB
}

func NewPartnerBlockRepository(db *sql.DB) *PartnerBlockRepository {
	return &PartnerBlockRepository{db: db}
}

const (
	insertPartnerBlock = `INSERT INTO eshkere.partner_block (
		partner_site_id, name, block_type, status, embed_token, cpm_strategy, amp_mode, size_mode, border_mode, corner_mode, theme, interscroller_mode, interscroller_background_color, revenue_share_bps, self_ad_settings, created_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW()) RETURNING id`

	selectPartnerBlockByID = `SELECT
		id, partner_site_id, name, block_type, status, embed_token, cpm_strategy, amp_mode, size_mode, border_mode, corner_mode, theme, interscroller_mode, interscroller_background_color, revenue_share_bps, self_ad_settings, created_at, updated_at
	FROM eshkere.partner_block
	WHERE id = $1`

	selectPartnerBlocksBySiteID = `SELECT
		id, partner_site_id, name, block_type, status, embed_token, cpm_strategy, amp_mode, size_mode, border_mode, corner_mode, theme, interscroller_mode, interscroller_background_color, revenue_share_bps, self_ad_settings, created_at, updated_at
	FROM eshkere.partner_block
	WHERE partner_site_id = $1
	ORDER BY created_at DESC, id DESC`

	updatePartnerBlock = `UPDATE eshkere.partner_block SET
		name = $1, status = $2, cpm_strategy = $3, amp_mode = $4, size_mode = $5, border_mode = $6, corner_mode = $7, theme = $8, interscroller_mode = $9, interscroller_background_color = $10, revenue_share_bps = $11, self_ad_settings = $12
	WHERE id = $13`

	deletePartnerBlock             = `DELETE FROM eshkere.partner_block WHERE id = $1`
	selectPartnerBlockByEmbedToken = `SELECT
		id, partner_site_id, name, block_type, status, embed_token, cpm_strategy, amp_mode, size_mode, border_mode, corner_mode, theme, interscroller_mode, interscroller_background_color, revenue_share_bps, self_ad_settings, created_at, updated_at
	FROM eshkere.partner_block
	WHERE embed_token = $1`
)

func (r *PartnerBlockRepository) Create(ctx context.Context, block *models.PartnerBlock) error {
	if block == nil {
		return fmt.Errorf("partner block cannot be nil")
	}
	err := r.db.QueryRowContext(ctx, insertPartnerBlock,
		block.PartnerSiteID, block.Name, block.BlockType, block.Status, block.EmbedToken,
		block.CPMStrategy, block.AmpMode, block.SizeMode, block.BorderMode, block.CornerMode,
		block.Theme, block.InterscrollerMode, block.InterscrollerBackgroundColor, block.RevenueShareBPS, block.SelfAdSettings,
	).Scan(&block.ID)
	if err != nil {
		return fmt.Errorf("insert partner block: %w", err)
	}
	return nil
}

func (r *PartnerBlockRepository) GetByID(ctx context.Context, blockID int) (*models.PartnerBlock, error) {
	var block models.PartnerBlock
	err := r.db.QueryRowContext(ctx, selectPartnerBlockByID, blockID).Scan(
		&block.ID, &block.PartnerSiteID, &block.Name, &block.BlockType, &block.Status, &block.EmbedToken,
		&block.CPMStrategy, &block.AmpMode, &block.SizeMode, &block.BorderMode, &block.CornerMode,
		&block.Theme, &block.InterscrollerMode, &block.InterscrollerBackgroundColor, &block.RevenueShareBPS, &block.SelfAdSettings,
		&block.CreatedAt, &block.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("partner block not found: %w", err)
		}
		return nil, fmt.Errorf("get partner block by id: %w", err)
	}
	return &block, nil
}

func (r *PartnerBlockRepository) ListBySiteID(ctx context.Context, siteID int) ([]*models.PartnerBlock, error) {
	rows, err := r.db.QueryContext(ctx, selectPartnerBlocksBySiteID, siteID)
	if err != nil {
		return nil, fmt.Errorf("list partner blocks: %w", err)
	}
	defer rows.Close()

	blocks := make([]*models.PartnerBlock, 0)
	for rows.Next() {
		var block models.PartnerBlock
		if err := rows.Scan(
			&block.ID, &block.PartnerSiteID, &block.Name, &block.BlockType, &block.Status, &block.EmbedToken,
			&block.CPMStrategy, &block.AmpMode, &block.SizeMode, &block.BorderMode, &block.CornerMode,
			&block.Theme, &block.InterscrollerMode, &block.InterscrollerBackgroundColor, &block.RevenueShareBPS, &block.SelfAdSettings,
			&block.CreatedAt, &block.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan partner block: %w", err)
		}
		blocks = append(blocks, &block)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate partner blocks: %w", err)
	}
	return blocks, nil
}

func (r *PartnerBlockRepository) Update(ctx context.Context, block *models.PartnerBlock) error {
	if block == nil {
		return fmt.Errorf("partner block cannot be nil")
	}
	_, err := r.db.ExecContext(ctx, updatePartnerBlock,
		block.Name, block.Status, block.CPMStrategy, block.AmpMode, block.SizeMode, block.BorderMode, block.CornerMode,
		block.Theme, block.InterscrollerMode, block.InterscrollerBackgroundColor, block.RevenueShareBPS, block.SelfAdSettings, block.ID,
	)
	if err != nil {
		return fmt.Errorf("update partner block: %w", err)
	}
	return nil
}

func (r *PartnerBlockRepository) Delete(ctx context.Context, blockID int) error {
	result, err := r.db.ExecContext(ctx, deletePartnerBlock, blockID)
	if err != nil {
		return fmt.Errorf("delete partner block: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete partner block rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("partner block not found")
	}
	return nil
}

func (r *PartnerBlockRepository) GetByEmbedToken(ctx context.Context, token string) (*models.PartnerBlock, error) {
	var block models.PartnerBlock
	err := r.db.QueryRowContext(ctx, selectPartnerBlockByEmbedToken, token).Scan(
		&block.ID, &block.PartnerSiteID, &block.Name, &block.BlockType, &block.Status, &block.EmbedToken,
		&block.CPMStrategy, &block.AmpMode, &block.SizeMode, &block.BorderMode, &block.CornerMode,
		&block.Theme, &block.InterscrollerMode, &block.InterscrollerBackgroundColor, &block.RevenueShareBPS, &block.SelfAdSettings,
		&block.CreatedAt, &block.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("partner block not found: %w", err)
		}
		return nil, fmt.Errorf("get partner block by token: %w", err)
	}
	return &block, nil
}
