package repository

import "sync"

type Ads struct {
	ID           int
	Title        string
	TargetAction string
	Price        int
}
type AdsRepository struct {
	mu      sync.RWMutex
	AdsByID map[int][]Ads
}

func InitAdsRepository() *AdsRepository {
	mockAds := []Ads{
		{ID: 1, Title: "iPhone 14", TargetAction: "Покупка", Price: 10000},
		{ID: 2, Title: "Магазин техники ЭщкеTech", TargetAction: "Переход по ссылке", Price: 15000},
	}
	adsRepo := &AdsRepository{
		mu: sync.RWMutex{},
		AdsByID: map[int][]Ads{
			1: mockAds,
		},
	}
	return adsRepo
}

func (r *AdsRepository) ListByID(id int) []Ads {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.AdsByID[id]
}
