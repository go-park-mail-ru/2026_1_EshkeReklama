package postgres

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPartnerRepository_SettleDailyEarnings(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewPartnerRepository(db)
	date := time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(selectUnsettledPartnerEarnings)).
		WithArgs(date).
		WillReturnRows(sqlmock.NewRows([]string{"partner_id", "sum"}).AddRow(42, int64(1400)))
	mock.ExpectExec(regexp.QuoteMeta(updatePartnerBalanceByEarning)).
		WithArgs(int64(1400), 42).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(markPartnerBlockEarningsSettled)).
		WithArgs(date).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	total, err := repo.SettleDailyEarnings(context.Background(), date)
	if err != nil {
		t.Fatalf("SettleDailyEarnings: %v", err)
	}
	if total != 1400 {
		t.Fatalf("expected settled total 1400 got %d", total)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
