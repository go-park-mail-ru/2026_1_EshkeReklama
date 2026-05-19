package analytics

import "time"

const (
	EventTypeImpression = "impression"
	EventTypeClick      = "click"
)

type AdEvent struct {
	EventID         string    `json:"event_id"`
	EventType       string    `json:"event_type"`
	RequestID       string    `json:"request_id"`
	OccurredAt      time.Time `json:"occurred_at"`
	VisitorID       string    `json:"visitor_id"`
	AdvertiserID    int       `json:"advertiser_id"`
	CampaignID      int       `json:"campaign_id"`
	AdGroupID       int       `json:"ad_group_id"`
	AdID            int       `json:"ad_id"`
	PartnerBlockID  int       `json:"partner_block_id"`
	PartnerSiteID   int       `json:"partner_site_id"`
	TopicID         int       `json:"topic_id"`
	Price           int64     `json:"price"`
	PartnerReward   int64     `json:"partner_reward"`
	PlatformRevenue int64     `json:"platform_revenue"`
}
