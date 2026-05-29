package service

import (
	"context"
	"errors"
	"testing"
	"time"

	errs "eshkere/internal/errors"
	"eshkere/internal/models"

	"go.uber.org/mock/gomock"
)

func TestExportCampaignStatsCSV_RequiresPro(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	advRepo := NewMockAdvertiserRepository(ctrl)
	svc, err := NewService(&Config{AdvertiserRepo: advRepo})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	advRepo.EXPECT().GetByID(gomock.Any(), 1).Return(&models.Advertiser{
		ID:     1,
		Tariff: models.TariffTypeBasic,
	}, nil)

	_, err = svc.ExportCampaignStatsCSV(context.Background(), 1, 5, time.Now(), time.Now())
	if !errors.Is(err, errs.ErrProRequired) {
		t.Fatalf("expected ErrProRequired, got %v", err)
	}
}
