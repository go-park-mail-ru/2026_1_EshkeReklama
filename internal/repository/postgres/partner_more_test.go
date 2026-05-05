package postgres

import (
	"context"
	"regexp"
	"testing"
	"time"

	"eshkere/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPartnerRepositoryCRUD(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewPartnerRepository(db)
	now := time.Now()
	updated := now.Add(time.Minute)
	partner := &models.Partner{
		ID:                     5,
		LastName:               "Ivanov",
		FirstName:              "Ivan",
		MiddleName:             "Ivanovich",
		BirthDate:              time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC),
		CountryCode:            "RU",
		RegistrationRegionCode: "RU-MOW",
		CooperationForm:        models.CooperationFormSelfEmployed,
		PayoutCurrency:         models.PayoutCurrencyRUB,
	}

	mock.ExpectExec(regexp.QuoteMeta(insertPartnerProfile)).
		WithArgs(5, "Ivanov", "Ivan", "Ivanovich", partner.BirthDate, "RU", "RU-MOW", models.CooperationFormSelfEmployed, models.PayoutCurrencyRUB).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.CreateProfile(context.Background(), partner); err != nil {
		t.Fatalf("create profile: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(selectPartnerByID)).
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows([]string{"id", "last_name", "first_name", "middle_name", "birth_date", "country_code", "registration_region_code", "cooperation_form", "payout_currency", "balance", "created_at", "updated_at"}).
			AddRow(5, "Ivanov", "Ivan", "Ivanovich", partner.BirthDate, "RU", "RU-MOW", models.CooperationFormSelfEmployed, models.PayoutCurrencyRUB, int64(100), now, updated))
	got, err := repo.GetByID(context.Background(), 5)
	if err != nil || got.ID != 5 || got.Balance != 100 {
		t.Fatalf("get by id: %#v %v", got, err)
	}

	mock.ExpectExec(regexp.QuoteMeta(updatePartner)).
		WithArgs("Petrov", "Petr", "Petrovich", partner.BirthDate, "RU", "RU-SPE", models.CooperationFormLegalEntity, models.PayoutCurrencyUSD, 5).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.Update(context.Background(), &models.Partner{
		ID:                     5,
		LastName:               "Petrov",
		FirstName:              "Petr",
		MiddleName:             "Petrovich",
		BirthDate:              partner.BirthDate,
		CountryCode:            "RU",
		RegistrationRegionCode: "RU-SPE",
		CooperationForm:        models.CooperationFormLegalEntity,
		PayoutCurrency:         models.PayoutCurrencyUSD,
	}); err != nil {
		t.Fatalf("update: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
