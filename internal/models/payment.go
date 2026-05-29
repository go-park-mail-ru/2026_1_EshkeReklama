package models

import (
	"database/sql"
	"time"
)

type PaymentTransactionStatus string
type PaymentType string

const (
	PaymentTransactionStatusPending   PaymentTransactionStatus = "pending"
	PaymentTransactionStatusSucceeded PaymentTransactionStatus = "succeeded"
	PaymentTransactionStatusCanceled  PaymentTransactionStatus = "canceled"
)

const (
	PaymentTypeBalance      PaymentType = "balance"
	PaymentTypeSubscription PaymentType = "subscription"
)

type PaymentTransaction struct {
	ID           string                   `db:"id"`
	AdvertiserID int                      `db:"advertiser_id"`
	Amount       int64                    `db:"amount"`
	Status       PaymentTransactionStatus `db:"status"`
	PaymentType  PaymentType              `db:"payment_type"`
	CreatedAt    time.Time                `db:"created_at"`
	UpdatedAt    sql.NullTime             `db:"updated_at"`
}

type PaymentCompletionResult struct {
	AdvertiserID int
	PaymentType  PaymentType
	Balance      int64
	AlreadyFinal bool
}
