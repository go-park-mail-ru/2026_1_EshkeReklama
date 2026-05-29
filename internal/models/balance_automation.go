package models

import (
	"database/sql"
	"time"
)

type AdvertiserNotificationSettings struct {
	AdvertiserID      int            `db:"advertiser_id"`
	EmailEnabled      bool           `db:"email_enabled"`
	TelegramEnabled   bool           `db:"telegram_enabled"`
	TelegramChatID    sql.NullString `db:"telegram_chat_id"`
	WarningThreshold  int64          `db:"warning_threshold"`
	CriticalThreshold int64          `db:"critical_threshold"`
	UpdatedAt         time.Time      `db:"updated_at"`
}

type AdvertiserAutopaySettings struct {
	AdvertiserID    int       `db:"advertiser_id"`
	Enabled         bool      `db:"enabled"`
	ThresholdAmount int64     `db:"threshold_amount"`
	TopUpAmount     int64     `db:"top_up_amount"`
	UpdatedAt       time.Time `db:"updated_at"`
}

type AdvertiserAutopayCandidate struct {
	AdvertiserID         int
	Balance              int64
	SavedPaymentMethodID string
	ThresholdAmount      int64
	TopUpAmount          int64
}
