package postgres

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"eshkere/internal/models"
	"eshkere/pkg/logger"

	"go.uber.org/zap"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestAppealRepositoryCRUD(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewAppealRepository(db)
	now := time.Now()
	updated := now.Add(time.Minute)
	ctx := logger.CtxWithLogger(context.Background(), zap.NewNop().Sugar())
	appeal := &models.Appeal{
		AdvertiserID: sql.NullInt64{Int64: 7, Valid: true},
		Status:       models.AppealStatusOpen,
		Category:     models.AppealCategoryBug,
		Title:        "Issue",
		Description:  "Something broke",
		ImageURL:     "image.png",
		Name:         "Ivan",
		Email:        "user@example.com",
	}

	mock.ExpectQuery(regexp.QuoteMeta(insertAppeal)).
		WithArgs(appeal.AdvertiserID, appeal.Status, appeal.Category, appeal.Title, appeal.Description, appeal.ImageURL, appeal.Name, appeal.Email).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
	if err := repo.Create(ctx, appeal); err != nil || appeal.ID != 3 {
		t.Fatalf("create: %#v %v", appeal, err)
	}

	mock.ExpectExec(regexp.QuoteMeta(updateAppealImage)).
		WithArgs("new.png", 3).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.UpdateImage(ctx, 3, "new.png"); err != nil {
		t.Fatalf("update image: %v", err)
	}

	rowCols := []string{"id", "advertiser_id", "status", "category", "title", "description", "image_url", "name", "email", "created_at", "updated_at"}
	mock.ExpectQuery(regexp.QuoteMeta(selectAppealByID)).
		WithArgs(3).
		WillReturnRows(sqlmock.NewRows(rowCols).
			AddRow(3, sql.NullInt64{Int64: 7, Valid: true}, models.AppealStatusOpen, models.AppealCategoryBug, "Issue", "Something broke", "new.png", "Ivan", "user@example.com", now, updated))
	got, err := repo.GetByID(ctx, 3)
	if err != nil || got.ID != 3 || got.Title != "Issue" {
		t.Fatalf("get by id: %#v %v", got, err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(selectAppealsByAdvertiserID)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows(rowCols).
			AddRow(3, sql.NullInt64{Int64: 7, Valid: true}, models.AppealStatusOpen, models.AppealCategoryBug, "Issue", "Something broke", "new.png", "Ivan", "user@example.com", now, updated))
	list, err := repo.ListByAdvertiserID(ctx, 7)
	if err != nil || len(list) != 1 {
		t.Fatalf("list by advertiser id: %#v %v", list, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
