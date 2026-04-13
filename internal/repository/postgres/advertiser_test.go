package postgres

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"eshkere/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestAdvertiserRepository_Create_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewAdvertiserRepository(db)

	adv := &models.Advertiser{
		Name:         "n",
		Email:        "e@test",
		Phone:        "+7000",
		PasswordHash: "h",
		PasswordSalt: "s",
		Balance:      0,
	}

	mock.ExpectQuery(regexp.QuoteMeta(insertAdvertiser)).
		WithArgs(adv.Name, adv.Email, adv.Phone, adv.PasswordHash, adv.PasswordSalt, adv.Balance).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(123))

	id, err := repo.Create(context.Background(), adv)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 123 || adv.ID != 123 {
		t.Fatalf("expected id 123 got %d adv.ID=%d", id, adv.ID)
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
