package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"eshkere/internal/models"
	"eshkere/pkg/logger"
)

type SupportRepository struct {
	db *sql.DB
}

func NewSupportRepository(db *sql.DB) *SupportRepository {
	return &SupportRepository{db: db}
}

const (
	insertSupportThread = `INSERT INTO eshkere.support_thread (
		campaign_id, advertiser_id, updated_at
	) VALUES ($1, $2, NOW()) RETURNING id, created_at, updated_at`

	selectSupportThreadByID = `SELECT
		id, campaign_id, advertiser_id, created_at, updated_at
	FROM eshkere.support_thread
	WHERE id = $1`

	selectSupportThreadByCampaignID = `SELECT
		id, campaign_id, advertiser_id, created_at, updated_at
	FROM eshkere.support_thread
	WHERE campaign_id = $1`

	insertSupportMessage = `INSERT INTO eshkere.support_message (
		thread_id, author_type, author_id, text
	) VALUES ($1, $2, $3, $4)
	RETURNING id, created_at`

	updateSupportThreadActivity = `UPDATE eshkere.support_thread
	SET updated_at = NOW()
	WHERE id = $1`

	selectSupportThreads = `SELECT
		st.id, st.campaign_id, st.advertiser_id, st.created_at, st.updated_at,
		sm.id, sm.thread_id, sm.author_type, sm.author_id, sm.text, sm.created_at,
		COALESCE(st.updated_at, st.created_at) AS last_activity
	FROM eshkere.support_thread st
	LEFT JOIN LATERAL (
		SELECT id, thread_id, author_type, author_id, text, created_at
		FROM eshkere.support_message
		WHERE thread_id = st.id
		ORDER BY created_at DESC, id DESC
		LIMIT 1
	) sm ON true
	ORDER BY COALESCE(st.updated_at, st.created_at) DESC, st.id DESC`
)

func (r *SupportRepository) CreateThread(ctx context.Context, thread *models.SupportThread) error {
	if thread == nil {
		return fmt.Errorf("support thread cannot be nil")
	}

	logger.GetLoggerFromCtx(ctx).Debug("db: create support thread")

	if err := r.db.QueryRowContext(ctx, insertSupportThread, thread.CampaignID, thread.AdvertiserID).
		Scan(&thread.ID, &thread.CreatedAt, &thread.UpdatedAt); err != nil {
		return fmt.Errorf("insert support thread: %w", err)
	}

	return nil
}

func (r *SupportRepository) GetThreadByID(ctx context.Context, threadID int) (*models.SupportThread, error) {
	logger.GetLoggerFromCtx(ctx).Debugf("db: get support thread by id: %d", threadID)

	var thread models.SupportThread
	if err := r.db.QueryRowContext(ctx, selectSupportThreadByID, threadID).
		Scan(&thread.ID, &thread.CampaignID, &thread.AdvertiserID, &thread.CreatedAt, &thread.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("support thread not found: %w", err)
		}
		return nil, fmt.Errorf("get support thread by id: %w", err)
	}

	return &thread, nil
}

func (r *SupportRepository) GetThreadByCampaignID(ctx context.Context, campaignID int) (*models.SupportThread, error) {
	logger.GetLoggerFromCtx(ctx).Debugf("db: get support thread by campaign id: %d", campaignID)

	var thread models.SupportThread
	if err := r.db.QueryRowContext(ctx, selectSupportThreadByCampaignID, campaignID).
		Scan(&thread.ID, &thread.CampaignID, &thread.AdvertiserID, &thread.CreatedAt, &thread.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("support thread not found: %w", err)
		}
		return nil, fmt.Errorf("get support thread by campaign id: %w", err)
	}

	return &thread, nil
}

func (r *SupportRepository) ListMessages(ctx context.Context, threadID, limit int, beforeID *int) ([]*models.SupportMessage, error) {
	logger.GetLoggerFromCtx(ctx).Debugf("db: list support messages by thread id: %d", threadID)

	args := []any{threadID}
	var b strings.Builder
	b.WriteString(`SELECT
		id, thread_id, author_type, author_id, text, created_at
	FROM eshkere.support_message
	WHERE thread_id = $1`)
	if beforeID != nil {
		args = append(args, *beforeID)
		fmt.Fprintf(&b, " AND id < $%d", len(args))
	}
	if limit <= 0 {
		limit = 50
	}
	args = append(args, limit)
	fmt.Fprintf(&b, " ORDER BY created_at DESC, id DESC LIMIT $%d", len(args))

	rows, err := r.db.QueryContext(ctx, b.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("list support messages: %w", err)
	}
	defer rows.Close()

	messages := make([]*models.SupportMessage, 0)
	for rows.Next() {
		var msg models.SupportMessage
		if err = rows.Scan(&msg.ID, &msg.ThreadID, &msg.AuthorType, &msg.AuthorID, &msg.Text, &msg.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan support message: %w", err)
		}
		messages = append(messages, &msg)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate support messages rows: %w", err)
	}

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (r *SupportRepository) CreateMessage(ctx context.Context, msg *models.SupportMessage) error {
	if msg == nil {
		return fmt.Errorf("support message cannot be nil")
	}

	logger.GetLoggerFromCtx(ctx).Debugf("db: create support message in thread: %d", msg.ThreadID)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin support message tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = tx.QueryRowContext(ctx, insertSupportMessage, msg.ThreadID, msg.AuthorType, msg.AuthorID, msg.Text).
		Scan(&msg.ID, &msg.CreatedAt); err != nil {
		return fmt.Errorf("insert support message: %w", err)
	}
	if _, err = tx.ExecContext(ctx, updateSupportThreadActivity, msg.ThreadID); err != nil {
		return fmt.Errorf("touch support thread: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit support message tx: %w", err)
	}

	return nil
}

func (r *SupportRepository) ListThreads(ctx context.Context) ([]*models.SupportThreadSummary, error) {
	logger.GetLoggerFromCtx(ctx).Debug("db: list support threads")

	rows, err := r.db.QueryContext(ctx, selectSupportThreads)
	if err != nil {
		return nil, fmt.Errorf("list support threads: %w", err)
	}
	defer rows.Close()

	threads := make([]*models.SupportThreadSummary, 0)
	for rows.Next() {
		var (
			thread       models.SupportThread
			lastMessage  models.SupportMessage
			lastID       sql.NullInt64
			lastThreadID sql.NullInt64
			authorType   sql.NullString
			authorID     sql.NullInt64
			text         sql.NullString
			createdAt    sql.NullTime
			lastActivity time.Time
		)

		if err = rows.Scan(
			&thread.ID,
			&thread.CampaignID,
			&thread.AdvertiserID,
			&thread.CreatedAt,
			&thread.UpdatedAt,
			&lastID,
			&lastThreadID,
			&authorType,
			&authorID,
			&text,
			&createdAt,
			&lastActivity,
		); err != nil {
			return nil, fmt.Errorf("scan support thread: %w", err)
		}

		summary := &models.SupportThreadSummary{
			Thread:       &thread,
			LastActivity: lastActivity,
		}
		if lastID.Valid {
			lastMessage = models.SupportMessage{
				ID:         int(lastID.Int64),
				ThreadID:   int(lastThreadID.Int64),
				AuthorType: models.SupportMessageAuthor(authorType.String),
				AuthorID:   authorID,
				Text:       text.String,
				CreatedAt:  createdAt.Time,
			}
			summary.LastMessage = &lastMessage
		}

		threads = append(threads, summary)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate support threads rows: %w", err)
	}

	return threads, nil
}
