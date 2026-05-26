package postgres

import (
	"context"
	"regexp"
	"testing"
	"time"

	"eshkere/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestAdRepository_ReserveImpression(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewAdRepository(db)
	date := time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(updateAdvertiserBalanceForImpression)).
		WithArgs(int64(20), 4).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(upsertCampaignDailySpend)).
		WithArgs(3, date, int64(20), int64(14), int64(6), int64(1000)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(upsertPartnerBlockDailyEarning)).
		WithArgs(7, date, int64(14)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	ok, err := repo.ReserveImpression(context.Background(), models.ImpressionReservation{
		CampaignID:      3,
		AdvertiserID:    4,
		PartnerBlockID:  7,
		SpendDate:       date,
		Price:           20,
		DailyBudget:     1000,
		PartnerReward:   14,
		PlatformRevenue: 6,
	})
	if err != nil {
		t.Fatalf("ReserveImpression: %v", err)
	}
	if !ok {
		t.Fatalf("expected reservation to succeed")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestAdRepository_ReserveImpressionInsufficientBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewAdRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(updateAdvertiserBalanceForImpression)).
		WithArgs(int64(20), 4).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	ok, err := repo.ReserveImpression(context.Background(), models.ImpressionReservation{
		AdvertiserID: 4,
		Price:        20,
	})
	if err != nil {
		t.Fatalf("ReserveImpression: %v", err)
	}
	if ok {
		t.Fatalf("expected reservation to fail")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
