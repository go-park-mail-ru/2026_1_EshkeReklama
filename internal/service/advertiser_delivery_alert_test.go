package service

import (
	"testing"

	"eshkere/internal/models"
)

func TestBuildAdvertiserDeliveryAlert_NoActiveCampaigns(t *testing.T) {
	alert := BuildAdvertiserDeliveryAlert(100, []*models.AdCampaign{
		{Status: models.AdStatusModeration},
		{Status: models.AdStatusTurnedOff},
	})
	if alert != nil {
		t.Fatalf("expected nil alert, got %+v", alert)
	}
}

func TestBuildAdvertiserDeliveryAlert_LowBalance(t *testing.T) {
	alert := BuildAdvertiserDeliveryAlert(500, []*models.AdCampaign{
		{Status: models.AdStatusWorking},
		{Status: models.AdStatusRejected},
	})
	if alert == nil {
		t.Fatal("expected alert")
	}
	if alert.Level != DeliveryAlertLowBalance {
		t.Fatalf("expected low_balance, got %s", alert.Level)
	}
	if alert.ActiveCampaigns != 1 || alert.AffectedCampaigns != 0 {
		t.Fatalf("unexpected counters: %+v", alert)
	}
}

func TestBuildAdvertiserDeliveryAlert_AtRisk(t *testing.T) {
	alert := BuildAdvertiserDeliveryAlert(100, []*models.AdCampaign{
		{Status: models.AdStatusWorking},
		{Status: models.AdStatusWorking},
	})
	if alert == nil {
		t.Fatal("expected alert")
	}
	if alert.Level != DeliveryAlertAtRisk {
		t.Fatalf("expected at_risk, got %s", alert.Level)
	}
	if alert.ActiveCampaigns != 2 || alert.AffectedCampaigns != 0 {
		t.Fatalf("unexpected counters: %+v", alert)
	}
}

func TestBuildAdvertiserDeliveryAlert_PartiallyStopped(t *testing.T) {
	alert := BuildAdvertiserDeliveryAlert(50, []*models.AdCampaign{
		{Status: models.AdStatusWorking},
		{Status: models.AdStatusNotEnoughMoney},
		{Status: models.AdStatusRejected},
	})
	if alert == nil {
		t.Fatal("expected alert")
	}
	if alert.Level != DeliveryAlertPartiallyStopped {
		t.Fatalf("expected partially_stopped, got %s", alert.Level)
	}
	if alert.ActiveCampaigns != 2 || alert.AffectedCampaigns != 1 {
		t.Fatalf("unexpected counters: %+v", alert)
	}
}

func TestBuildAdvertiserDeliveryAlert_FullyStopped(t *testing.T) {
	alert := BuildAdvertiserDeliveryAlert(0, []*models.AdCampaign{
		{Status: models.AdStatusNotEnoughMoney},
		{Status: models.AdStatusNotEnoughMoney},
	})
	if alert == nil {
		t.Fatal("expected alert")
	}
	if alert.Level != DeliveryAlertFullyStopped {
		t.Fatalf("expected fully_stopped, got %s", alert.Level)
	}
	if alert.ActiveCampaigns != 2 || alert.AffectedCampaigns != 2 {
		t.Fatalf("unexpected counters: %+v", alert)
	}
}
