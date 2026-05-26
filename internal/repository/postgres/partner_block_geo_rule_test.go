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

func TestPartnerBlockGeoRuleRepository(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewPartnerBlockGeoRuleRepository(db)
	now := time.Now()
	updated := now.Add(time.Minute)

	mock.ExpectQuery(regexp.QuoteMeta(selectPartnerBlockGeoRulesByBlockID)).
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows([]string{"id", "partner_block_id", "geo_code", "is_enabled", "cpmv", "created_at", "updated_at"}).
			AddRow(1, 5, "all", true, sql.NullInt64{Int64: 15, Valid: true}, now, updated))
	rules, err := repo.ListByBlockID(context.Background(), 5)
	if err != nil || len(rules) != 1 || rules[0].GeoCode != "all" {
		t.Fatalf("list by block id: %#v %v", rules, err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(deletePartnerBlockGeoRulesByBlockID)).
		WithArgs(5).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertPartnerBlockGeoRule)).
		WithArgs(5, "RU-MOW", true, sql.NullInt64{Int64: 20, Valid: true}).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	err = repo.ReplaceByBlockID(context.Background(), 5, []*models.PartnerBlockGeoRule{
		nil,
		{GeoCode: "RU-MOW", IsEnabled: true, CPMV: sql.NullInt64{Int64: 20, Valid: true}},
	})
	if err != nil {
		t.Fatalf("replace by block id: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
