package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	errs "eshkere/internal/errors"
	"eshkere/internal/models"

	"go.uber.org/mock/gomock"
)

type stubAutopayRepo struct {
	getByAdvertiserIDFn func(ctx context.Context, advertiserID int) (*models.AdvertiserAutopaySettings, error)
	upsertFn            func(ctx context.Context, settings *models.AdvertiserAutopaySettings) error
	listEligibleFn      func(ctx context.Context) ([]models.AdvertiserAutopayCandidate, error)
}

func (s stubAutopayRepo) GetByAdvertiserID(ctx context.Context, advertiserID int) (*models.AdvertiserAutopaySettings, error) {
	return s.getByAdvertiserIDFn(ctx, advertiserID)
}
func (s stubAutopayRepo) Upsert(ctx context.Context, settings *models.AdvertiserAutopaySettings) error {
	return s.upsertFn(ctx, settings)
}
func (s stubAutopayRepo) ListEligible(ctx context.Context) ([]models.AdvertiserAutopayCandidate, error) {
	if s.listEligibleFn != nil {
		return s.listEligibleFn(ctx)
	}
	return nil, nil
}

type stubNotificationRepo struct {
	getByAdvertiserIDFn func(ctx context.Context, advertiserID int) (*models.AdvertiserNotificationSettings, error)
	upsertFn            func(ctx context.Context, settings *models.AdvertiserNotificationSettings) error
	listEnabledFn       func(ctx context.Context) ([]models.AdvertiserNotificationSettings, error)
}

func (s stubNotificationRepo) GetByAdvertiserID(ctx context.Context, advertiserID int) (*models.AdvertiserNotificationSettings, error) {
	return s.getByAdvertiserIDFn(ctx, advertiserID)
}
func (s stubNotificationRepo) Upsert(ctx context.Context, settings *models.AdvertiserNotificationSettings) error {
	return s.upsertFn(ctx, settings)
}
func (s stubNotificationRepo) ListEnabled(ctx context.Context) ([]models.AdvertiserNotificationSettings, error) {
	return s.listEnabledFn(ctx)
}

type stubPartnerIncomeRepo struct {
	listFn func(ctx context.Context, partnerID int, from, to time.Time) ([]PartnerIncomeRow, error)
}

func (s stubPartnerIncomeRepo) ListPartnerIncome(ctx context.Context, partnerID int, from, to time.Time) ([]PartnerIncomeRow, error) {
	return s.listFn(ctx, partnerID, from, to)
}

func TestBalanceSettingsAndNotificationSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var gotAutoSettings *models.AdvertiserAutopaySettings
	autoSettings := &models.AdvertiserAutopaySettings{AdvertiserID: 2, Enabled: true, ThresholdAmount: 1500, TopUpAmount: 3000}
	autoRepo := stubAutopayRepo{
		getByAdvertiserIDFn: func(_ context.Context, advertiserID int) (*models.AdvertiserAutopaySettings, error) {
			if advertiserID != 2 {
				t.Fatalf("unexpected advertiser id for autopay repo: %d", advertiserID)
			}
			return autoSettings, nil
		},
		upsertFn: func(_ context.Context, settings *models.AdvertiserAutopaySettings) error {
			gotAutoSettings = settings
			return nil
		},
	}
	var gotNotifySettings *models.AdvertiserNotificationSettings
	notifySettings := &models.AdvertiserNotificationSettings{AdvertiserID: 3, EmailEnabled: true, WarningThreshold: 500, CriticalThreshold: 100}
	notifyRepo := stubNotificationRepo{
		getByAdvertiserIDFn: func(_ context.Context, advertiserID int) (*models.AdvertiserNotificationSettings, error) {
			if advertiserID != 3 {
				t.Fatalf("unexpected advertiser id for notification repo: %d", advertiserID)
			}
			return notifySettings, nil
		},
		upsertFn: func(_ context.Context, settings *models.AdvertiserNotificationSettings) error {
			gotNotifySettings = settings
			return nil
		},
		listEnabledFn: func(_ context.Context) ([]models.AdvertiserNotificationSettings, error) {
			return []models.AdvertiserNotificationSettings{*notifySettings}, nil
		},
	}
	svc, _ := NewService(&Config{
		AutopaySettingsRepo:      autoRepo,
		NotificationSettingsRepo: notifyRepo,
	})

	if _, err := svc.GetAdvertiserAutopaySettings(context.Background(), 0); !errors.Is(err, errs.ErrInvalidAdvertiserArg) {
		t.Fatalf("expected invalid advertiser arg, got %v", err)
	}
	if err := svc.UpdateAdvertiserAutopaySettings(context.Background(), &models.AdvertiserAutopaySettings{AdvertiserID: 1, Enabled: true, ThresholdAmount: 999, TopUpAmount: 2000}); !errors.Is(err, errs.ErrInvalidAdvertiserArg) {
		t.Fatalf("expected threshold validation error, got %v", err)
	}
	if err := svc.UpdateAdvertiserAutopaySettings(context.Background(), &models.AdvertiserAutopaySettings{AdvertiserID: 1, Enabled: true, ThresholdAmount: 1500, TopUpAmount: 1000}); !errors.Is(err, errs.ErrInvalidAdvertiserArg) {
		t.Fatalf("expected limit validation error, got %v", err)
	}

	if got, err := svc.GetAdvertiserAutopaySettings(context.Background(), 2); err != nil || got != autoSettings {
		t.Fatalf("GetAdvertiserAutopaySettings mismatch got=%+v err=%v", got, err)
	}
	if err := svc.UpdateAdvertiserAutopaySettings(context.Background(), autoSettings); err != nil {
		t.Fatalf("UpdateAdvertiserAutopaySettings: %v", err)
	}
	if gotAutoSettings != autoSettings {
		t.Fatalf("expected autopay upsert to receive original settings, got %+v", gotAutoSettings)
	}

	if _, err := svc.GetAdvertiserNotificationSettings(context.Background(), -1); !errors.Is(err, errs.ErrInvalidAdvertiserArg) {
		t.Fatalf("expected invalid advertiser error, got %v", err)
	}
	if err := svc.UpdateAdvertiserNotificationSettings(context.Background(), &models.AdvertiserNotificationSettings{AdvertiserID: 1, WarningThreshold: 100, CriticalThreshold: 100}); !errors.Is(err, errs.ErrInvalidAdvertiserArg) {
		t.Fatalf("expected invalid notification thresholds, got %v", err)
	}

	if got, err := svc.GetAdvertiserNotificationSettings(context.Background(), 3); err != nil || got != notifySettings {
		t.Fatalf("GetAdvertiserNotificationSettings mismatch got=%+v err=%v", got, err)
	}
	if err := svc.UpdateAdvertiserNotificationSettings(context.Background(), notifySettings); err != nil {
		t.Fatalf("UpdateAdvertiserNotificationSettings: %v", err)
	}
	if gotNotifySettings != notifySettings {
		t.Fatalf("expected notification upsert to receive original settings, got %+v", gotNotifySettings)
	}
	if enabled, err := svc.GetEnabledNotificationSettings(context.Background()); err != nil || len(enabled) != 1 {
		t.Fatalf("GetEnabledNotificationSettings mismatch enabled=%+v err=%v", enabled, err)
	}

	svc.notificationSettingsRepo = nil
	if enabled, err := svc.GetEnabledNotificationSettings(context.Background()); err != nil || enabled != nil {
		t.Fatalf("expected nil settings with nil repo, got %+v err=%v", enabled, err)
	}
}

func TestGetPartnerIncomeStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	incomeRepo := stubPartnerIncomeRepo{
		listFn: func(_ context.Context, partnerID int, from, to time.Time) ([]PartnerIncomeRow, error) {
			if partnerID != 5 || from.Format("2006-01-02") != "2026-05-01" || to.Format("2006-01-02") != "2026-05-02" {
				t.Fatalf("unexpected partner income query: partner=%d from=%s to=%s", partnerID, from, to)
			}
			return []PartnerIncomeRow{
				{Impressions: 1000, Reward: 250},
				{Impressions: 500, Reward: 125},
			}, nil
		},
	}
	svc, _ := NewService(&Config{PartnerIncomeRepo: incomeRepo})

	if _, err := svc.GetPartnerIncomeStats(context.Background(), 0, time.Now(), time.Now()); !errors.Is(err, errs.BadRequestError) {
		t.Fatalf("expected bad request for partner id, got %v", err)
	}
	if _, err := svc.GetPartnerIncomeStats(context.Background(), 1, time.Date(2026, 5, 3, 0, 0, 0, 0, time.UTC), time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)); !errors.Is(err, errs.BadRequestError) {
		t.Fatalf("expected bad request for date range, got %v", err)
	}

	svc.partnerIncomeRepo = nil
	if _, err := svc.GetPartnerIncomeStats(context.Background(), 1, time.Now(), time.Now()); !errors.Is(err, errs.NotImplementedError) {
		t.Fatalf("expected not implemented without repo, got %v", err)
	}
	svc.partnerIncomeRepo = incomeRepo

	from := time.Date(2026, 5, 1, 12, 0, 0, 0, time.FixedZone("x", 3*3600))
	to := time.Date(2026, 5, 2, 22, 0, 0, 0, time.FixedZone("x", 3*3600))
	stats, err := svc.GetPartnerIncomeStats(context.Background(), 5, from, to)
	if err != nil {
		t.Fatalf("GetPartnerIncomeStats: %v", err)
	}
	if stats.Impressions != 1500 || stats.Reward != 375 || stats.ECPM != 250 {
		t.Fatalf("unexpected partner income stats: %+v", stats)
	}
}

func TestGetAdvertiserByIDAndAppealByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	advRepo := NewMockAdvertiserRepository(ctrl)
	appealRepo := NewMockAppealRepository(ctrl)
	appealStorage := NewMockAppealStorage(ctrl)
	svc, _ := NewService(&Config{
		AdvertiserRepo: advRepo,
		AppealRepo:     appealRepo,
		AppealStorage:  appealStorage,
	})

	if _, err := svc.GetAdvertiserByID(context.Background(), 0); !errors.Is(err, errs.ErrInvalidAdvertiserArg) {
		t.Fatalf("expected invalid advertiser id error, got %v", err)
	}

	adv := &models.Advertiser{ID: 1, AvatarURL: sql.NullString{String: "avatars/1.png", Valid: true}}
	advRepo.EXPECT().GetByID(gomock.Any(), 1).Return(adv, nil)
	if got, err := svc.GetAdvertiserByID(context.Background(), 1); err != nil || got.ID != 1 {
		t.Fatalf("GetAdvertiserByID mismatch got=%+v err=%v", got, err)
	}

	appeal := &models.Appeal{ID: 2, ImageURL: "appeals/2.png"}
	appealRepo.EXPECT().GetByID(gomock.Any(), 2).Return(appeal, nil)
	appealStorage.EXPECT().GetAppealImageURL("appeals/2.png").Return("https://cdn.example/appeals/2.png")
	if got, err := svc.GetAppealByID(context.Background(), 2); err != nil || got.ImageURL != "https://cdn.example/appeals/2.png" {
		t.Fatalf("GetAppealByID mismatch got=%+v err=%v", got, err)
	}
}
