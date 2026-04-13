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

func TestFeedLinkRepository_UpsertAndGet(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewFeedLinkRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(upsertFeedLink)).
		WithArgs(1, "tok").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.UpsertByAdvertiserID(context.Background(), 1, "tok"); err != nil {
		t.Fatalf("UpsertByAdvertiserID: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(selectAdvertiserByFeedToken)).
		WithArgs("tok").
		WillReturnRows(sqlmock.NewRows([]string{"advertiser_id"}).AddRow(1))

	id, err := repo.GetAdvertiserIDByToken(context.Background(), "tok")
	if err != nil {
		t.Fatalf("GetAdvertiserIDByToken: %v", err)
	}
	if id != 1 {
		t.Fatalf("expected 1 got %d", id)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestAdRepository_CreateListDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewAdRepository(db)

	ad := &models.Ad{AdGroupID: 2, Status: models.AdStatusModeration, Title: "t", ShortDesc: "s", ImageURL: "i", TargetURL: "u"}
	mock.ExpectQuery(regexp.QuoteMeta(insertAd)).
		WithArgs(ad.AdGroupID, ad.Status, ad.Title, ad.ShortDesc, ad.ImageURL, ad.TargetURL).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(9))

	if err := repo.Create(context.Background(), ad); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if ad.ID != 9 {
		t.Fatalf("expected id 9 got %d", ad.ID)
	}

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(selectAdsByAdGroupID)).
		WithArgs(2).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "ad_group_id", "status", "title", "short_desc", "image_url", "target_url", "created_at", "updated_at"}).
				AddRow(9, 2, models.AdStatusWorking, "t", "s", "i", "u", now, sql.NullTime{}),
		)

	list, err := repo.ListByAdGroupID(context.Background(), 2)
	if err != nil {
		t.Fatalf("ListByAdGroupID: %v", err)
	}
	if len(list) != 1 || list[0].ID != 9 {
		t.Fatalf("unexpected list: %+v", list)
	}

	mock.ExpectExec(regexp.QuoteMeta(deleteAd)).
		WithArgs(9).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Delete(context.Background(), 9); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
