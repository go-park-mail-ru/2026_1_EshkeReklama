package models

import "time"

type AdCandidate struct {
	CampaignID        int
	AdvertiserID      int
	DailyBudget       int64
	CPMPrice          int64
	SpentToday        int64
	AdvertiserBalance int64
	Ad                *Ad
}

type ImpressionReservation struct {
	CampaignID      int
	AdvertiserID    int
	PartnerBlockID  int
	SpendDate       time.Time
	Price           int64
	DailyBudget     int64
	PartnerReward   int64
	PlatformRevenue int64
}
