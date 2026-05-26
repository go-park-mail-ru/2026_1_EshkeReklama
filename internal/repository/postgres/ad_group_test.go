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

func TestAdGroupRepository_CRUD(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewAdGroupRepository(db)

	g := &models.AdGroup{AdCampaignID: 1, TopicID: 2, RegionID: 3, Name: "g", AgeFrom: 18, AgeTo: 25, Gender: "any"}
	mock.ExpectQuery(regexp.QuoteMeta(insertAdGroup)).
		WithArgs(g.AdCampaignID, g.TopicID, g.RegionID, g.Name, g.AgeFrom, g.AgeTo, g.Gender).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))

	if err := repo.Create(context.Background(), g); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if g.ID != 5 {
		t.Fatalf("expected id 5")
	}

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(selectAdGroupByID)).
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows([]string{"id", "ad_campaign_id", "topic_id", "region_id", "name", "age_from", "age_to", "gender", "created_at", "updated_at"}).
			AddRow(5, 1, 2, 3, "g", 18, 25, "any", now, sql.NullTime{}),
		)

	got, err := repo.GetByID(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.ID != 5 {
		t.Fatalf("expected id 5")
	}

	mock.ExpectQuery(regexp.QuoteMeta(selectAdGroupsByCampaignID)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "ad_campaign_id", "topic_id", "region_id", "name", "age_from", "age_to", "gender", "created_at", "updated_at"}).
			AddRow(5, 1, 2, 3, "g", 18, 25, "any", now, sql.NullTime{}),
		)

	list, err := repo.ListByCampaignID(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListByCampaignID: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1")
	}

	g.UpdatedAt = sql.NullTime{Time: now, Valid: true}
	mock.ExpectExec(regexp.QuoteMeta(updateAdGroup)).
		WithArgs(g.AdCampaignID, g.TopicID, g.RegionID, g.Name, g.AgeFrom, g.AgeTo, g.Gender, g.UpdatedAt, g.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Update(context.Background(), g); err != nil {
		t.Fatalf("Update: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(deleteAdGroup)).
		WithArgs(5).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Delete(context.Background(), 5); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
