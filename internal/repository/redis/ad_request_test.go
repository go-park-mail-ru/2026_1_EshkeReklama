package redis

import (
	"context"
	"testing"
	"time"

	"eshkere/internal/service"

	"github.com/alicebob/miniredis/v2"
	redigo "github.com/gomodule/redigo/redis"
)

func TestAdRequestStore(t *testing.T) {
	mr := miniredis.RunT(t)
	pool := &redigo.Pool{
		Dial: func() (redigo.Conn, error) {
			return redigo.Dial("tcp", mr.Addr())
		},
	}
	store := NewAdRequestStore(pool)

	if err := store.Save(context.Background(), service.AdRequestRecord{}, time.Hour); err != nil {
		t.Fatalf("expected nil for empty request id, got %v", err)
	}

	record := service.AdRequestRecord{
		RequestID:       "req-1",
		VisitorID:       "visitor-1",
		AdvertiserID:    2,
		CampaignID:      4,
		AdGroupID:       5,
		AdID:            3,
		PartnerBlockID:  6,
		PartnerSiteID:   8,
		TopicID:         7,
		TargetURL:       "https://example.com",
		Price:           10,
		PartnerReward:   7,
		PlatformRevenue: 3,
	}
	if err := store.Save(context.Background(), record, time.Hour); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := store.Get(context.Background(), "req-1")
	if err != nil || got.RequestID != "req-1" || got.AdID != 3 || got.TopicID != 7 ||
		got.AdvertiserID != 2 || got.CampaignID != 4 || got.AdGroupID != 5 ||
		got.PartnerBlockID != 6 || got.PartnerSiteID != 8 || got.Price != 10 ||
		got.PartnerReward != 7 || got.PlatformRevenue != 3 {
		t.Fatalf("get: %#v %v", got, err)
	}
	first, err := store.MarkClickedOnce(context.Background(), "req-1", time.Hour)
	if err != nil || !first {
		t.Fatalf("mark first click: %v %v", first, err)
	}
	again, err := store.MarkClickedOnce(context.Background(), "req-1", time.Hour)
	if err != nil || again {
		t.Fatalf("mark second click: %v %v", again, err)
	}
	if _, err := store.Get(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty request id")
	}
	if adRequestKey("req-1") != "adreq:req-1" {
		t.Fatalf("unexpected redis key")
	}
}
