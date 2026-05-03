package models

import (
	"database/sql"
	"time"
)

type CooperationForm string
type PayoutCurrency string

const (
	CooperationFormSelfEmployed           CooperationForm = "self_employed"
	CooperationFormIndividualEntrepreneur CooperationForm = "individual_entrepreneur"
	CooperationFormLegalEntity            CooperationForm = "legal_entity"

	PayoutCurrencyRUB PayoutCurrency = "RUB"
	PayoutCurrencyUSD PayoutCurrency = "USD"
	PayoutCurrencyEUR PayoutCurrency = "EUR"
)

type Partner struct {
	ID                     int             `db:"id"`
	LastName               string          `db:"last_name"`
	FirstName              string          `db:"first_name"`
	MiddleName             string          `db:"middle_name"`
	BirthDate              time.Time       `db:"birth_date"`
	CountryCode            string          `db:"country_code"`
	RegistrationRegionCode string          `db:"registration_region_code"`
	CooperationForm        CooperationForm `db:"cooperation_form"`
	PayoutCurrency         PayoutCurrency  `db:"payout_currency"`
	Balance                int64           `db:"balance"`
	CreatedAt              time.Time       `db:"created_at"`
	UpdatedAt              sql.NullTime    `db:"updated_at"`
}
