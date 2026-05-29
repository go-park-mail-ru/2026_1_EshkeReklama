package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	errs "eshkere/internal/errors"
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"

	"go.uber.org/mock/gomock"
)

func TestRequireProActive(t *testing.T) {
	t.Parallel()

	future := time.Now().Add(24 * time.Hour)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	advRepo := NewMockAdvertiserRepository(ctrl)
	svc, err := NewService(&Config{AdvertiserRepo: advRepo})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	advRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&models.Advertiser{
		ID:              1,
		Tariff:          models.TariffTypePro,
		TariffExpiresAt: sql.NullTime{Time: future, Valid: true},
	}, nil)

	if _, err := svc.RequireProActive(context.Background(), 1); err != nil {
		t.Fatalf("RequireProActive: %v", err)
	}

	advRepo.EXPECT().GetByID(gomock.Any(), 2).Return(&models.Advertiser{
		ID:     2,
		Tariff: models.TariffTypeBasic,
	}, nil)

	if _, err := svc.RequireProActive(context.Background(), 2); !errors.Is(err, errs.ErrProRequired) {
		t.Fatalf("expected ErrProRequired, got %v", err)
	}
}

func TestRunSubscriptionExpiryCycle(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	advRepo := NewMockAdvertiserRepository(ctrl)
	svc, err := NewService(&Config{AdvertiserRepo: advRepo})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	advRepo.EXPECT().ListExpiredProAdvertiserIDs(gomock.Any()).Return([]int{10, 11}, nil)
	advRepo.EXPECT().GetByID(gomock.Any(), 10).Return(&models.Advertiser{
		ID:     10,
		Tariff: models.TariffTypePro,
	}, nil)
	advRepo.EXPECT().Update(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, adv *models.Advertiser) error {
			if adv.ID != 10 || adv.Tariff != models.TariffTypeBasic || adv.TariffExpiresAt.Valid {
				t.Fatalf("unexpected deactivate update for advertiser 10: %+v", adv)
			}
			return nil
		})
	advRepo.EXPECT().GetByID(gomock.Any(), 11).Return(&models.Advertiser{
		ID:     11,
		Tariff: models.TariffTypePro,
	}, nil)
	advRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	deactivated, err := svc.RunSubscriptionExpiryCycle(context.Background())
	if err != nil {
		t.Fatalf("RunSubscriptionExpiryCycle: %v", err)
	}
	if deactivated != 2 {
		t.Fatalf("deactivated = %d, want 2", deactivated)
	}
}

func TestCheckActiveCampaignLimit_ExpiredPro(t *testing.T) {
	t.Parallel()

	past := time.Now().Add(-24 * time.Hour)
	adv := &models.Advertiser{
		Tariff:          models.TariffTypePro,
		TariffExpiresAt: sql.NullTime{Time: past, Valid: true},
	}

	err := checkActiveCampaignLimit(adv, adv.MaxCampaigns())
	if !errors.Is(err, errs.ErrProRequired) {
		t.Fatalf("expected ErrProRequired, got %v", err)
	}
}

func TestPurchaseProSubscription_Success(t *testing.T) {
	t.Parallel()

	price := int64(models.SubscriptionPriceRub)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	advRepo := NewMockAdvertiserRepository(ctrl)
	cRepo := NewMockAdCampaignRepository(ctrl)
	svc, err := NewService(&Config{AdvertiserRepo: advRepo, AdCampaignRepo: cRepo})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	advRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&models.Advertiser{
		ID:      1,
		Balance: price + 500,
		Tariff:  models.TariffTypeBasic,
	}, nil)
	advRepo.EXPECT().Update(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, adv *models.Advertiser) error {
			if adv.Balance != 500 {
				t.Fatalf("balance after debit = %d, want 500", adv.Balance)
			}
			return nil
		})
	advRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&models.Advertiser{
		ID:      1,
		Balance: 500,
		Tariff:  models.TariffTypeBasic,
	}, nil)
	advRepo.EXPECT().Update(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, adv *models.Advertiser) error {
			if adv.Tariff != models.TariffTypePro || !adv.TariffExpiresAt.Valid {
				t.Fatalf("unexpected pro activation: %+v", adv)
			}
			return nil
		})
	cRepo.EXPECT().CountActiveByAdvertiserID(gomock.Any(), 1).Return(0, nil)

	info, err := svc.PurchaseProSubscription(context.Background(), 1)
	if err != nil {
		t.Fatalf("PurchaseProSubscription: %v", err)
	}
	if !info.IsProActive || info.Tariff != string(models.TariffTypePro) {
		t.Fatalf("unexpected tariff info: %+v", info)
	}
}

func TestPurchaseProSubscription_InsufficientBalance(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	advRepo := NewMockAdvertiserRepository(ctrl)
	svc, err := NewService(&Config{AdvertiserRepo: advRepo})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	advRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&models.Advertiser{
		ID:      1,
		Balance: 100,
		Tariff:  models.TariffTypeBasic,
	}, nil)

	_, err = svc.PurchaseProSubscription(context.Background(), 1)
	if !errors.Is(err, errs.ErrInsufficientBalance) {
		t.Fatalf("expected ErrInsufficientBalance, got %v", err)
	}
}

func TestCreateAdCampaign_ExpiredProLimit(t *testing.T) {
	t.Parallel()

	past := time.Now().Add(-24 * time.Hour)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	advRepo := NewMockAdvertiserRepository(ctrl)
	cRepo := NewMockAdCampaignRepository(ctrl)
	svc, err := NewService(&Config{AdvertiserRepo: advRepo, AdCampaignRepo: cRepo})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	advRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&models.Advertiser{
		ID:              1,
		Tariff:          models.TariffTypePro,
		TariffExpiresAt: sql.NullTime{Time: past, Valid: true},
	}, nil)
	cRepo.EXPECT().CountActiveByAdvertiserID(gomock.Any(), 1).Return(models.MaxCampaignsBasic, nil)

	_, err = svc.CreateAdCampaign(context.Background(), &serviceinput.CreateAdCampaign{
		AdvertiserID: 1,
		Name:         "camp",
	})
	if !errors.Is(err, errs.ErrProRequired) {
		t.Fatalf("expected ErrProRequired, got %v", err)
	}
}
