package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	redis "github.com/gomodule/redigo/redis"
)

func TestRepositoryGetTopicsAndTrackTopic(t *testing.T) {
	mr := miniredis.RunT(t)
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", mr.Addr())
		},
	}

	repo := New(pool)
	now := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)
	if err := repo.TrackTopic(context.Background(), "visitor-1", 3, 5, now); err != nil {
		t.Fatalf("track topic: %v", err)
	}
	if err := repo.TrackTopic(context.Background(), "", 3, 5, now); err != nil {
		t.Fatalf("expected nil for empty visitor, got %v", err)
	}
	if err := repo.TrackTopic(context.Background(), "visitor-1", 0, 5, now); err != nil {
		t.Fatalf("expected nil for invalid topic, got %v", err)
	}

	topics, err := repo.GetTopics(context.Background(), "visitor-1", 10)
	if err != nil || len(topics) != 1 || topics[0].TopicID != 3 || topics[0].Score != 5 {
		t.Fatalf("get topics: %v %v", topics, err)
	}
	if got, err := repo.GetTopics(context.Background(), "", 10); err != nil || got != nil {
		t.Fatalf("expected nil for empty visitor, got %v %v", got, err)
	}
	if err := repo.TrackRegion(context.Background(), "visitor-1", 9, 5, now); err != nil {
		t.Fatalf("track region: %v", err)
	}
	regions, err := repo.GetRegions(context.Background(), "visitor-1", 10)
	if err != nil || len(regions) != 1 || regions[0].RegionID != 9 || regions[0].Score != 5 {
		t.Fatalf("get regions: %v %v", regions, err)
	}
	if profileKey("abc") != "profile:abc" || topicsKey("abc") != "profile:abc:topics" ||
		regionsKey("abc") != "profile:abc:regions" {
		t.Fatal("unexpected redis keys")
	}
}
