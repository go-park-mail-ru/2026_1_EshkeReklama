package models

import (
	"database/sql"
	"time"
)

type Advertiser struct {
	ID           int          `db:"id"`
	Name         string       `db:"name"`
	Email        string       `db:"email"`
	Phone        string       `db:"phone_number"`
	PasswordHash string       `db:"password_hash"`
	PasswordSalt string       `db:"password_salt"`
	Balance      int64        `db:"balance"`
	CreatedAt    time.Time    `db:"created_at"`
	UpdatedAt    sql.NullTime `db:"updated_at"`
}
