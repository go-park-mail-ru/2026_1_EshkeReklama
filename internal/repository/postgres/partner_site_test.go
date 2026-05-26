package postgres

import (
	"context"
	"regexp"
	"testing"
	"time"

	"eshkere/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPartnerSiteRepositoryCRUD(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewPartnerSiteRepository(db)
	now := time.Now()
	updated := now.Add(time.Minute)

	site := &models.PartnerSite{PartnerID: 7, Domain: "example.com", SiteName: "Example", Status: models.PartnerSiteStatusDraft}
	mock.ExpectQuery(regexp.QuoteMeta(insertPartnerSite)).
		WithArgs(7, "example.com", "Example", models.PartnerSiteStatusDraft).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	if err := repo.Create(context.Background(), site); err != nil || site.ID != 1 {
		t.Fatalf("create: site=%#v err=%v", site, err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(selectPartnerSiteByID)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "partner_id", "domain", "site_name", "status", "created_at", "updated_at"}).
			AddRow(1, 7, "example.com", "Example", models.PartnerSiteStatusActive, now, updated))
	got, err := repo.GetByID(context.Background(), 1)
	if err != nil || got.Status != models.PartnerSiteStatusActive {
		t.Fatalf("get by id: %#v %v", got, err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(selectPartnerSitesByPartnerID)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"id", "partner_id", "domain", "site_name", "status", "created_at", "updated_at"}).
			AddRow(1, 7, "example.com", "Example", models.PartnerSiteStatusActive, now, updated).
			AddRow(2, 7, "example.org", "Example Org", models.PartnerSiteStatusDraft, now, nil))
	sites, err := repo.ListByPartnerID(context.Background(), 7)
	if err != nil || len(sites) != 2 {
		t.Fatalf("list by partner id: %#v %v", sites, err)
	}

	mock.ExpectExec(regexp.QuoteMeta(updatePartnerSite)).
		WithArgs("example.net", "Renamed", models.PartnerSiteStatusBlocked, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.Update(context.Background(), &models.PartnerSite{
		ID:       1,
		Domain:   "example.net",
		SiteName: "Renamed",
		Status:   models.PartnerSiteStatusBlocked,
	}); err != nil {
		t.Fatalf("update: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(selectPartnerSiteExistsByDomain)).
		WithArgs("example.net").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	exists, err := repo.ExistsByDomain(context.Background(), "example.net")
	if err != nil || !exists {
		t.Fatalf("exists by domain: %v %v", exists, err)
	}

	mock.ExpectExec(regexp.QuoteMeta(deletePartnerSite)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.Delete(context.Background(), 1); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
