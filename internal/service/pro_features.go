package service

import "eshkere/internal/models"

// TariffFeatures описывает доступные возможности текущего тарифа.
type TariffFeatures struct {
	PriorityModeration bool `json:"priority_moderation"`
	StatsExport        bool `json:"stats_export"`
	TelegramAlerts     bool `json:"telegram_alerts"`
}

func tariffFeaturesFor(adv *models.Advertiser) TariffFeatures {
	active := adv != nil && adv.IsProActive()
	return TariffFeatures{
		PriorityModeration: active,
		StatsExport:        active,
		TelegramAlerts:     active,
	}
}
