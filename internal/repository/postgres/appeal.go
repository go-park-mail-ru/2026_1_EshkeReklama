package postgres

import (
	"context"
	"database/sql"
	"errors"
	errs "eshkere/internal/errors"
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
	"eshkere/pkg/logger"
	"fmt"
	"strings"
)

type AppealRepository struct {
	db *sql.DB
}

func NewAppealRepository(db *sql.DB) *AppealRepository {
	return &AppealRepository{db: db}
}

const (
	insertAppeal = `INSERT INTO eshkere.appeal (
		advertiser_id, status, category, title, description, image_url, name, email)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	insertAppealStatusHistory = `INSERT INTO eshkere.appeal_status_history (appeal_id, status) VALUES ($1, $2)`

	updateAppealImage = `UPDATE eshkere.appeal
	SET image_url = $1
	WHERE id = $2`

	updateAppealStatus = `UPDATE eshkere.appeal
	SET status = $1
	WHERE id = $2`

	selectAppealByID = `SELECT
		id, advertiser_id, status, category, title, description, image_url, name, email, created_at, updated_at
	FROM eshkere.appeal
	WHERE id = $1`

	selectAppealsByAdvertiserID = `SELECT
		id, advertiser_id, status, category, title, description, image_url, name, email, created_at, updated_at
	FROM eshkere.appeal
	WHERE advertiser_id = $1
	ORDER BY created_at DESC, id DESC`

	selectAppealMessagesByAppealID = `SELECT id, appeal_id, author, text, created_at
	FROM eshkere.appeal_message
	WHERE appeal_id = $1
	ORDER BY created_at ASC, id ASC`

	selectAppealStatusHistoryByAppealID = `SELECT id, appeal_id, status, created_at
	FROM eshkere.appeal_status_history
	WHERE appeal_id = $1
	ORDER BY created_at ASC, id ASC`

	insertAppealMessage = `INSERT INTO eshkere.appeal_message (appeal_id, author, text)
	VALUES ($1, $2, $3) RETURNING id, created_at`
)

func (r *AppealRepository) Create(ctx context.Context, appeal *models.Appeal) error {
	if appeal == nil {
		return fmt.Errorf("appeal cannot be nil")
	}

	logger.GetLoggerFromCtx(ctx).Debug("db: create appeal")

	err := r.db.QueryRowContext(ctx, insertAppeal,
		appeal.AdvertiserID, appeal.Status, appeal.Category, appeal.Title, appeal.Description, appeal.ImageURL, appeal.Name, appeal.Email,
	).Scan(&appeal.ID)
	if err != nil {
		return fmt.Errorf("insert appeal: %w", err)
	}

	if _, err := r.db.ExecContext(ctx, insertAppealStatusHistory, appeal.ID, appeal.Status); err != nil {
		return fmt.Errorf("insert appeal status history: %w", err)
	}

	return nil
}

func (r *AppealRepository) UpdateImage(ctx context.Context, appealID int, imageKey string) error {
	logger.GetLoggerFromCtx(ctx).Debugf("db: update appeal image by id: %d", appealID)

	if _, err := r.db.ExecContext(ctx, updateAppealImage, imageKey, appealID); err != nil {
		return fmt.Errorf("update appeal image: %w", err)
	}

	return nil
}

func (r *AppealRepository) GetByID(ctx context.Context, appealID int) (*models.Appeal, error) {
	logger.GetLoggerFromCtx(ctx).Debugf("db: get appeal by id: %d", appealID)
	var appeal models.Appeal

	err := r.db.QueryRowContext(ctx, selectAppealByID, appealID).Scan(
		&appeal.ID,
		&appeal.AdvertiserID,
		&appeal.Status,
		&appeal.Category,
		&appeal.Title,
		&appeal.Description,
		&appeal.ImageURL,
		&appeal.Name,
		&appeal.Email,
		&appeal.CreatedAt,
		&appeal.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.NotFoundError
		}
		return nil, fmt.Errorf("get appeal by id: %w", err)
	}

	return &appeal, nil
}

func (r *AppealRepository) ListByAdvertiserID(ctx context.Context, advertiserID int) ([]*models.Appeal, error) {
	logger.GetLoggerFromCtx(ctx).Debugf("db: get appeals by advertiserID: %d", advertiserID)

	rows, err := r.db.QueryContext(ctx, selectAppealsByAdvertiserID, advertiserID)
	if err != nil {
		return nil, fmt.Errorf("list appeals by advertiserID: %w", err)
	}
	defer rows.Close()

	appeals := make([]*models.Appeal, 0)
	for rows.Next() {
		var appeal models.Appeal
		if err = rows.Scan(
			&appeal.ID,
			&appeal.AdvertiserID,
			&appeal.Status,
			&appeal.Category,
			&appeal.Title,
			&appeal.Description,
			&appeal.ImageURL,
			&appeal.Name,
			&appeal.Email,
			&appeal.CreatedAt,
			&appeal.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan appeal: %w", err)
		}

		appeals = append(appeals, &appeal)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate appeals rows: %w", err)
	}

	return appeals, nil
}

func (r *AppealRepository) ListMessages(ctx context.Context, appealID int) ([]*models.AppealMessage, error) {
	logger.GetLoggerFromCtx(ctx).Debugf("db: list appeal messages: %d", appealID)

	rows, err := r.db.QueryContext(ctx, selectAppealMessagesByAppealID, appealID)
	if err != nil {
		return nil, fmt.Errorf("list appeal messages: %w", err)
	}
	defer rows.Close()

	out := make([]*models.AppealMessage, 0)
	for rows.Next() {
		var msg models.AppealMessage
		if err := rows.Scan(&msg.ID, &msg.AppealID, &msg.Author, &msg.Text, &msg.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan appeal message: %w", err)
		}
		out = append(out, &msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate appeal messages: %w", err)
	}
	return out, nil
}

func (r *AppealRepository) AddMessage(ctx context.Context, msg *models.AppealMessage) error {
	if msg == nil {
		return fmt.Errorf("message cannot be nil")
	}
	logger.GetLoggerFromCtx(ctx).Debugf("db: add appeal message: %d", msg.AppealID)

	if err := r.db.QueryRowContext(ctx, insertAppealMessage, msg.AppealID, msg.Author, msg.Text).Scan(&msg.ID, &msg.CreatedAt); err != nil {
		return fmt.Errorf("insert appeal message: %w", err)
	}
	return nil
}

func (r *AppealRepository) AdminList(ctx context.Context, filter *serviceinput.AdminListAppealsFilter) ([]*models.Appeal, error) {
	if filter == nil {
		return nil, fmt.Errorf("filter cannot be nil")
	}

	logger.GetLoggerFromCtx(ctx).Debug("db: admin list appeals")

	where := make([]string, 0, 3)
	args := make([]any, 0, 5)

	argN := 1
	if filter.AdvertiserID != nil {
		where = append(where, fmt.Sprintf("advertiser_id = $%d", argN))
		args = append(args, *filter.AdvertiserID)
		argN++
	}
	if filter.Status != nil {
		where = append(where, fmt.Sprintf("status = $%d", argN))
		args = append(args, *filter.Status)
		argN++
	}
	if filter.Category != nil {
		where = append(where, fmt.Sprintf("category = $%d", argN))
		args = append(args, *filter.Category)
		argN++
	}

	query := `SELECT id, advertiser_id, status, category, title, description, image_url, name, email, created_at, updated_at
FROM eshkere.appeal`
	if len(where) > 0 {
		query += "\nWHERE " + strings.Join(where, " AND ")
	}
	query += fmt.Sprintf("\nORDER BY created_at DESC, id DESC\nLIMIT $%d OFFSET $%d", argN, argN+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("admin list appeals: %w", err)
	}
	defer rows.Close()

	appeals := make([]*models.Appeal, 0)
	for rows.Next() {
		var appeal models.Appeal
		if err = rows.Scan(
			&appeal.ID,
			&appeal.AdvertiserID,
			&appeal.Status,
			&appeal.Category,
			&appeal.Title,
			&appeal.Description,
			&appeal.ImageURL,
			&appeal.Name,
			&appeal.Email,
			&appeal.CreatedAt,
			&appeal.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan appeal: %w", err)
		}
		appeals = append(appeals, &appeal)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate appeals rows: %w", err)
	}

	return appeals, nil
}

func (r *AppealRepository) AdminListMessages(ctx context.Context, appealID int) ([]*models.AppealMessage, error) {
	return r.ListMessages(ctx, appealID)
}

func (r *AppealRepository) AdminListStatusHistory(ctx context.Context, appealID int) ([]*models.AppealStatusHistory, error) {
	logger.GetLoggerFromCtx(ctx).Debugf("db: admin list appeal status history: %d", appealID)

	rows, err := r.db.QueryContext(ctx, selectAppealStatusHistoryByAppealID, appealID)
	if err != nil {
		return nil, fmt.Errorf("list appeal status history: %w", err)
	}
	defer rows.Close()

	out := make([]*models.AppealStatusHistory, 0)
	for rows.Next() {
		var h models.AppealStatusHistory
		if err := rows.Scan(&h.ID, &h.AppealID, &h.Status, &h.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan appeal status history: %w", err)
		}
		out = append(out, &h)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate appeal status history: %w", err)
	}
	return out, nil
}

func (r *AppealRepository) AdminAddMessage(ctx context.Context, msg *models.AppealMessage) error {
	return r.AddMessage(ctx, msg)
}

func (r *AppealRepository) AdminUpdateStatus(ctx context.Context, appealID int, status models.AppealStatus) error {
	logger.GetLoggerFromCtx(ctx).Debugf("db: admin update appeal status: %d", appealID)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx, updateAppealStatus, status, appealID)
	if err != nil {
		return fmt.Errorf("update appeal status: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return errs.NotFoundError
	}

	if _, err := tx.ExecContext(ctx, insertAppealStatusHistory, appealID, status); err != nil {
		return fmt.Errorf("insert status history: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
