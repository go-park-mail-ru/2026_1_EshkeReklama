package internal

type AdResponse struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	TargetAction string `json:"target_action"`
	Price        int    `json:"price"`
}

type ListAdsResponse struct {
	AdvertiserID int          `json:"advertiser_id"`
	Ads          []AdResponse `json:"ads"`
}
