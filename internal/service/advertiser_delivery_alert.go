package service

import "eshkere/internal/models"

type DeliveryAlertLevel string

const (
	DeliveryAlertLowBalance       DeliveryAlertLevel = "low_balance"
	DeliveryAlertAtRisk           DeliveryAlertLevel = "at_risk"
	DeliveryAlertPartiallyStopped DeliveryAlertLevel = "partially_stopped"
	DeliveryAlertFullyStopped     DeliveryAlertLevel = "fully_stopped"
)

const (
	deliveryAlertWarningBalance  int64 = 500
	deliveryAlertCriticalBalance int64 = 100
)

type AdvertiserDeliveryAlert struct {
	Level             DeliveryAlertLevel
	Balance           int64
	ActiveCampaigns   int
	AffectedCampaigns int
}

func BuildAdvertiserDeliveryAlert(balance int64, campaigns []*models.AdCampaign) *AdvertiserDeliveryAlert {
	workingCount := 0
	notEnoughMoneyCount := 0

	for _, campaign := range campaigns {
		if campaign == nil {
			continue
		}

		switch campaign.Status {
		case models.AdStatusWorking:
			workingCount++
		case models.AdStatusNotEnoughMoney:
			notEnoughMoneyCount++
		}
	}

	activeCount := workingCount + notEnoughMoneyCount
	if activeCount == 0 {
		return nil
	}

	if notEnoughMoneyCount > 0 {
		level := DeliveryAlertPartiallyStopped
		if notEnoughMoneyCount == activeCount {
			level = DeliveryAlertFullyStopped
		}

		return &AdvertiserDeliveryAlert{
			Level:             level,
			Balance:           balance,
			ActiveCampaigns:   activeCount,
			AffectedCampaigns: notEnoughMoneyCount,
		}
	}

	switch {
	case balance <= deliveryAlertCriticalBalance:
		return &AdvertiserDeliveryAlert{
			Level:             DeliveryAlertAtRisk,
			Balance:           balance,
			ActiveCampaigns:   activeCount,
			AffectedCampaigns: 0,
		}
	case balance <= deliveryAlertWarningBalance:
		return &AdvertiserDeliveryAlert{
			Level:             DeliveryAlertLowBalance,
			Balance:           balance,
			ActiveCampaigns:   activeCount,
			AffectedCampaigns: 0,
		}
	default:
		return nil
	}
}
