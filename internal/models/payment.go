package models

import (
	"database/sql"
	"time"
)

type PaymentTransactionStatus string

const (
	PaymentTransactionStatusPending   PaymentTransactionStatus = "pending"
	PaymentTransactionStatusSucceeded PaymentTransactionStatus = "succeeded"
	PaymentTransactionStatusCanceled  PaymentTransactionStatus = "canceled"
)

type PaymentTransaction struct {
	ID           string                   `db:"id"`
	AdvertiserID int                      `db:"advertiser_id"`
	Amount       int64                    `db:"amount"`
	Status       PaymentTransactionStatus `db:"status"`
	CreatedAt    time.Time                `db:"created_at"`
	UpdatedAt    sql.NullTime             `db:"updated_at"`
}

type PaymentCompletionResult struct {
	AdvertiserID int
	Balance      int64
	AlreadyFinal bool
}
