package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"regexp"
	"testing"
	"time"

	"eshkere/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPartnerBlockRepositoryCRUD(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewPartnerBlockRepository(db)
	now := time.Now()
	updated := now.Add(time.Minute)
	background := sql.NullString{String: "#fff", Valid: true}
	payload := json.RawMessage(`{"reserved":true}`)

	block := &models.PartnerBlock{
		PartnerSiteID:                3,
		Name:                         "Main",
		BlockType:                    models.PartnerBlockTypeBanner,
		Status:                       models.PartnerBlockStatusActive,
		EmbedToken:                   "token",
		CPMStrategy:                  models.CPMStrategyMaxIncome,
		AmpMode:                      models.AmpModeDisabled,
		SizeMode:                     models.SizeModeAdaptive,
		BorderMode:                   models.BorderModeAuto,
		CornerMode:                   models.CornerModeAuto,
		Theme:                        models.ThemeModeLight,
		InterscrollerMode:            models.InterscrollerModeAuto,
		InterscrollerBackgroundColor: background,
		RevenueShareBPS:              7000,
		SelfAdSettings:               payload,
	}
	mock.ExpectQuery(regexp.QuoteMeta(insertPartnerBlock)).
		WithArgs(3, "Main", models.PartnerBlockTypeBanner, models.PartnerBlockStatusActive, "token",
			models.CPMStrategyMaxIncome, models.AmpModeDisabled, models.SizeModeAdaptive, models.BorderModeAuto,
			models.CornerModeAuto, models.ThemeModeLight, models.InterscrollerModeAuto, background, 7000, payload).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))
	if err := repo.Create(context.Background(), block); err != nil || block.ID != 5 {
		t.Fatalf("create: %#v %v", block, err)
	}

	rowCols := []string{"id", "partner_site_id", "name", "block_type", "status", "embed_token", "cpm_strategy", "amp_mode", "size_mode", "border_mode", "corner_mode", "theme", "interscroller_mode", "interscroller_background_color", "revenue_share_bps", "self_ad_settings", "created_at", "updated_at"}
	mock.ExpectQuery(regexp.QuoteMeta(selectPartnerBlockByID)).
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows(rowCols).
			AddRow(5, 3, "Main", models.PartnerBlockTypeBanner, models.PartnerBlockStatusActive, "token",
				models.CPMStrategyMaxIncome, models.AmpModeDisabled, models.SizeModeAdaptive, models.BorderModeAuto,
				models.CornerModeAuto, models.ThemeModeLight, models.InterscrollerModeAuto, background, 7000, payload, now, updated))
	got, err := repo.GetByID(context.Background(), 5)
	if err != nil || got.ID != 5 {
		t.Fatalf("get by id: %#v %v", got, err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(selectPartnerBlocksBySiteID)).
		WithArgs(3).
		WillReturnRows(sqlmock.NewRows(rowCols).
			AddRow(5, 3, "Main", models.PartnerBlockTypeBanner, models.PartnerBlockStatusActive, "token",
				models.CPMStrategyMaxIncome, models.AmpModeDisabled, models.SizeModeAdaptive, models.BorderModeAuto,
				models.CornerModeAuto, models.ThemeModeLight, models.InterscrollerModeAuto, background, 7000, payload, now, updated))
	blocks, err := repo.ListBySiteID(context.Background(), 3)
	if err != nil || len(blocks) != 1 {
		t.Fatalf("list by site id: %#v %v", blocks, err)
	}

	mock.ExpectExec(regexp.QuoteMeta(updatePartnerBlock)).
		WithArgs("Updated", models.PartnerBlockStatusInactive, models.CPMStrategyMaxIncome, models.AmpModeDisabled,
			models.SizeModeAdaptive, models.BorderModeAuto, models.CornerModeAuto, models.ThemeModeLight,
			models.InterscrollerModeAuto, background, 7000, payload, 5).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.Update(context.Background(), &models.PartnerBlock{
		ID:                           5,
		Name:                         "Updated",
		Status:                       models.PartnerBlockStatusInactive,
		CPMStrategy:                  models.CPMStrategyMaxIncome,
		AmpMode:                      models.AmpModeDisabled,
		SizeMode:                     models.SizeModeAdaptive,
		BorderMode:                   models.BorderModeAuto,
		CornerMode:                   models.CornerModeAuto,
		Theme:                        models.ThemeModeLight,
		InterscrollerMode:            models.InterscrollerModeAuto,
		InterscrollerBackgroundColor: background,
		RevenueShareBPS:              7000,
		SelfAdSettings:               payload,
	}); err != nil {
		t.Fatalf("update: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(selectPartnerBlockByEmbedToken)).
		WithArgs("token").
		WillReturnRows(sqlmock.NewRows(rowCols).
			AddRow(5, 3, "Main", models.PartnerBlockTypeBanner, models.PartnerBlockStatusActive, "token",
				models.CPMStrategyMaxIncome, models.AmpModeDisabled, models.SizeModeAdaptive, models.BorderModeAuto,
				models.CornerModeAuto, models.ThemeModeLight, models.InterscrollerModeAuto, background, 7000, payload, now, updated))
	if got, err = repo.GetByEmbedToken(context.Background(), "token"); err != nil || got.EmbedToken != "token" {
		t.Fatalf("get by embed token: %#v %v", got, err)
	}

	mock.ExpectExec(regexp.QuoteMeta(deletePartnerBlock)).
		WithArgs(5).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.Delete(context.Background(), 5); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
