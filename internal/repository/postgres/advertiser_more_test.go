package postgres

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestAdvertiserRepository_GetByEmailPhone_UpdateDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewAdvertiserRepository(db)

	now := time.Now()
	row := sqlmock.NewRows([]string{"id", "name", "email", "phone_number", "avatar_url", "password_hash", "password_salt", "balance", "created_at", "updated_at"}).
		AddRow(1, "n", "e@test", "900", sql.NullString{}, "h", "s", int64(10), now, sql.NullTime{})

	mock.ExpectQuery(regexp.QuoteMeta(selectAdvertiserByEmail)).
		WithArgs("e@test").
		WillReturnRows(row)

	adv, err := repo.GetByEmail(context.Background(), "e@test")
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}
	if adv.ID != 1 {
		t.Fatalf("expected id 1")
	}

	row2 := sqlmock.NewRows([]string{"id", "name", "email", "phone_number", "avatar_url", "password_hash", "password_salt", "balance", "created_at", "updated_at"}).
		AddRow(2, "n", "e2@test", "900", sql.NullString{}, "h", "s", int64(0), now, sql.NullTime{})
	mock.ExpectQuery(regexp.QuoteMeta(selectAdvertiserByPhone)).
		WithArgs("900").
		WillReturnRows(row2)

	adv2, err := repo.GetByPhone(context.Background(), "900")
	if err != nil {
		t.Fatalf("GetByPhone: %v", err)
	}
	if adv2.ID != 2 {
		t.Fatalf("expected id 2")
	}

	adv2.Name = "x"
	mock.ExpectExec(regexp.QuoteMeta(updateAdvertiser)).
		WithArgs(adv2.Name, adv2.Email, adv2.Phone, adv2.AvatarURL, adv2.PasswordHash, adv2.PasswordSalt, adv2.Balance, adv2.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Update(context.Background(), adv2); err != nil {
		t.Fatalf("Update: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(deleteAdvertiser)).
		WithArgs(2).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Delete(context.Background(), 2); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
