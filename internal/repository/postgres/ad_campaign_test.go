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

func TestAdCampaignRepository_ListAndUpdateAndDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewAdCampaignRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "advertiser_id", "status", "name", "daily_budget", "main_action", "created_at", "updated_at"}).
		AddRow(1, 7, models.AdStatusWorking, "c", int64(10), "", now, sql.NullTime{})

	mock.ExpectQuery(regexp.QuoteMeta(selectAdCampaignsByAdvertiserID)).
		WithArgs(7).
		WillReturnRows(rows)

	list, err := repo.ListByAdvertiserID(context.Background(), 7)
	if err != nil {
		t.Fatalf("ListByAdvertiserID: %v", err)
	}
	if len(list) != 1 || list[0].ID != 1 {
		t.Fatalf("unexpected list: %+v", list)
	}

	c := &models.AdCampaign{ID: 1, AdvertiserID: 7, Status: models.AdStatusWorking, Name: "n", DailyBudget: 1, UpdatedAt: sql.NullTime{Time: now, Valid: true}}
	mock.ExpectExec(regexp.QuoteMeta(updateAdCampaign)).
		WithArgs(c.AdvertiserID, c.Status, c.Name, c.DailyBudget, c.MainAction, c.UpdatedAt, c.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Update(context.Background(), c); err != nil {
		t.Fatalf("Update: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(deleteAdCampaign)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Delete(context.Background(), 1); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
