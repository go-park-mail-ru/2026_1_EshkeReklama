package postgres

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestAdvertiserRepository_CreateProfile_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewAdvertiserRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(insertProfile)).
		WithArgs(int64(123), "name").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.CreateProfile(context.Background(), 123, "name"); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestAdvertiserRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewAdvertiserRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(selectAdvertiserByID)).
		WithArgs(777).
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetByID(context.Background(), 777)
	if err == nil {
		t.Fatalf("expected error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
