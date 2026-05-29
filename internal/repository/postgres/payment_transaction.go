package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	errs "eshkere/internal/errors"
	"eshkere/internal/models"
	"eshkere/pkg/logger"
)

type PaymentTransactionRepository struct {
	db *sql.DB
}

func NewPaymentTransactionRepository(db *sql.DB) *PaymentTransactionRepository {
	return &PaymentTransactionRepository{db: db}
}

const (
	insertPaymentTransaction = `INSERT INTO eshkere.payment_transaction
		(id, advertiser_id, amount, status, payment_type, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())`

	selectPaymentTransactionByID = `SELECT
		id, advertiser_id, amount, status, payment_type, created_at, updated_at
	FROM eshkere.payment_transaction
	WHERE id = $1`

	selectPaymentTransactionForUpdate = `SELECT advertiser_id, amount, status, payment_type
	FROM eshkere.payment_transaction
	WHERE id = $1
	FOR UPDATE`

	updatePaymentTransactionStatus = `UPDATE eshkere.payment_transaction
	SET status = $2, updated_at = $3
	WHERE id = $1`

	updateAdvertiserPaymentMethod = `UPDATE eshkere.advertiser
	SET saved_payment_method_id = $1, saved_payment_method_title = $2, updated_at = $3
	WHERE id = $4`

	incrementAdvertiserBalance = `UPDATE eshkere.advertiser
	SET balance = balance + $1, updated_at = $3
	WHERE id = $2
	RETURNING balance`
)

func (r *PaymentTransactionRepository) Create(ctx context.Context, tx *models.PaymentTransaction) error {
	if tx == nil {
		return fmt.Errorf("payment transaction cannot be nil")
	}

	_, err := r.db.ExecContext(ctx, insertPaymentTransaction, tx.ID, tx.AdvertiserID, tx.Amount, tx.Status, tx.PaymentType)
	if err != nil {
		return fmt.Errorf("insert payment transaction: %w", err)
	}

	return nil
}

func (r *PaymentTransactionRepository) GetByID(ctx context.Context, id string) (*models.PaymentTransaction, error) {
	logger.GetLoggerFromCtx(ctx).Debugf("db: get payment transaction by id: %s", id)

	var tx models.PaymentTransaction
	err := r.db.QueryRowContext(ctx, selectPaymentTransactionByID, id).Scan(
		&tx.ID,
		&tx.AdvertiserID,
		&tx.Amount,
		&tx.Status,
		&tx.PaymentType,
		&tx.CreatedAt,
		&tx.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: payment transaction not found", errs.NotFoundError)
		}
		return nil, fmt.Errorf("get payment transaction by id: %w", err)
	}

	return &tx, nil
}

func (r *PaymentTransactionRepository) Complete(
	ctx context.Context,
	paymentID string,
	status models.PaymentTransactionStatus,
	paymentMethodID string,
	paymentMethodTitle string,
) (*models.PaymentCompletionResult, error) {
	dbTx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin payment completion tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = dbTx.Rollback()
		}
	}()

	var (
		advertiserID int
		amount       int64
		current      models.PaymentTransactionStatus
		paymentType  models.PaymentType
	)
	err = dbTx.QueryRowContext(ctx, selectPaymentTransactionForUpdate, paymentID).Scan(&advertiserID, &amount, &current, &paymentType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: payment transaction not found", errs.NotFoundError)
		}
		return nil, fmt.Errorf("select payment transaction for update: %w", err)
	}

	result := &models.PaymentCompletionResult{
		AdvertiserID: advertiserID,
		PaymentType:  paymentType,
		AlreadyFinal: current == models.PaymentTransactionStatusSucceeded || current == models.PaymentTransactionStatusCanceled,
	}
	if result.AlreadyFinal {
		if current == models.PaymentTransactionStatusSucceeded {
			err = dbTx.QueryRowContext(ctx, `SELECT balance FROM eshkere.advertiser WHERE id = $1`, advertiserID).Scan(&result.Balance)
			if err != nil {
				return nil, fmt.Errorf("load advertiser balance after already-final payment: %w", err)
			}
		}
		if commitErr := dbTx.Commit(); commitErr != nil {
			return nil, fmt.Errorf("commit already-final payment tx: %w", commitErr)
		}
		return result, nil
	}

	now := time.Now()
	if _, err = dbTx.ExecContext(ctx, updatePaymentTransactionStatus, paymentID, status, now); err != nil {
		return nil, fmt.Errorf("update payment transaction status: %w", err)
	}

	if status == models.PaymentTransactionStatusSucceeded {
		if _, err = dbTx.ExecContext(ctx, updateAdvertiserPaymentMethod, nullableString(paymentMethodID), nullableString(paymentMethodTitle), now, advertiserID); err != nil {
			return nil, fmt.Errorf("update advertiser payment method: %w", err)
		}

		switch paymentType {
		case models.PaymentTypeSubscription:
			if err = dbTx.QueryRowContext(ctx, `SELECT balance FROM eshkere.advertiser WHERE id = $1`, advertiserID).Scan(&result.Balance); err != nil {
				return nil, fmt.Errorf("load advertiser balance after subscription payment: %w", err)
			}
		default:
			if err = dbTx.QueryRowContext(ctx, incrementAdvertiserBalance, amount, advertiserID, now).Scan(&result.Balance); err != nil {
				return nil, fmt.Errorf("increment advertiser balance: %w", err)
			}
		}
	}

	if err = dbTx.Commit(); err != nil {
		return nil, fmt.Errorf("commit payment completion tx: %w", err)
	}

	return result, nil
}

func nullableString(v string) any {
	if v == "" {
		return nil
	}
	return v
}
