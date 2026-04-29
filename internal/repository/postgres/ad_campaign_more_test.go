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

func TestAdCampaignRepository_CreateAndGetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewAdCampaignRepository(db)

	c := &models.AdCampaign{AdvertiserID: 7, Status: models.AdStatusModeration, Name: "n", DailyBudget: 10}
	mock.ExpectQuery(regexp.QuoteMeta(insertAdCampaign)).
		WithArgs(c.AdvertiserID, c.Status, c.Name, c.DailyBudget, c.MainAction).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))

	if err := repo.Create(context.Background(), c); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if c.ID != 3 {
		t.Fatalf("expected id 3 got %d", c.ID)
	}

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(selectAdCampaignByID)).
		WithArgs(3).
		WillReturnRows(sqlmock.NewRows([]string{"id", "advertiser_id", "status", "name", "daily_budget", "main_action", "created_at", "updated_at"}).
			AddRow(3, 7, models.AdStatusModeration, "n", int64(10), "", now, sql.NullTime{}),
		)

	got, err := repo.GetByID(context.Background(), 3)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.ID != 3 || got.AdvertiserID != 7 {
		t.Fatalf("unexpected campaign: %+v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
