package postgres

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"eshkere/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestAdvertiserRepository_GetByID_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewAdvertiserRepository(db)

	now := time.Now()
	cols := []string{"id", "name", "surname", "avatar_url", "balance", "company", "city", "tariff", "tariff_expires_at", "role", "saved_payment_method_id", "saved_payment_method_title", "created_at", "updated_at"}

	row := sqlmock.NewRows(cols).
		AddRow(1, "n", "", sql.NullString{}, int64(10), "", "", "basic", sql.NullTime{}, "user", sql.NullString{}, sql.NullString{}, now, sql.NullTime{})

	mock.ExpectQuery(regexp.QuoteMeta(selectAdvertiserByID)).
		WithArgs(1).
		WillReturnRows(row)

	adv, err := repo.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if adv.ID != 1 {
		t.Fatalf("expected id 1")
	}

	adv.Name = "x"
	mock.ExpectExec(regexp.QuoteMeta(updateAdvertiser)).
		WithArgs(adv.Name, adv.Surname, adv.AvatarURL, adv.Balance, adv.Company, adv.City, adv.Tariff, adv.TariffExpiresAt, adv.Role, adv.SavedPaymentMethodID, adv.SavedPaymentMethodTitle, adv.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Update(context.Background(), adv); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestAdvertiserRepository_ListExpiredProAdvertiserIDs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewAdvertiserRepository(db)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(10).AddRow(11)
	mock.ExpectQuery(regexp.QuoteMeta(listExpiredProAdvertiserIDs)).WillReturnRows(rows)

	ids, err := repo.ListExpiredProAdvertiserIDs(context.Background())
	if err != nil {
		t.Fatalf("ListExpiredProAdvertiserIDs: %v", err)
	}
	if len(ids) != 2 || ids[0] != 10 || ids[1] != 11 {
		t.Fatalf("unexpected ids: %v", ids)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestAdvertiserRepository_Update_Nil(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewAdvertiserRepository(db)

	if err := repo.Update(context.Background(), (*models.Advertiser)(nil)); err == nil {
		t.Fatalf("expected error")
	}
}
