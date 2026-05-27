package service

import (
	"context"
	"errors"
	"testing"
	"time"
)

type stubRepo struct {
	getTopicsFunc   func(context.Context, string, int) ([]TopicScore, error)
	getRegionsFunc  func(context.Context, string, int) ([]RegionScore, error)
	trackTopicFunc  func(context.Context, string, int, float64, time.Time) error
	trackRegionFunc func(context.Context, string, int, float64, time.Time) error
}

func (s *stubRepo) GetTopics(ctx context.Context, visitorID string, limit int) ([]TopicScore, error) {
	return s.getTopicsFunc(ctx, visitorID, limit)
}

func (s *stubRepo) GetRegions(ctx context.Context, visitorID string, limit int) ([]RegionScore, error) {
	if s.getRegionsFunc == nil {
		return nil, nil
	}
	return s.getRegionsFunc(ctx, visitorID, limit)
}

func (s *stubRepo) TrackTopic(ctx context.Context, visitorID string, topicID int, scoreDelta float64, now time.Time) error {
	return s.trackTopicFunc(ctx, visitorID, topicID, scoreDelta, now)
}

func (s *stubRepo) TrackRegion(ctx context.Context, visitorID string, regionID int, scoreDelta float64, now time.Time) error {
	if s.trackRegionFunc == nil {
		return nil
	}
	return s.trackRegionFunc(ctx, visitorID, regionID, scoreDelta, now)
}

func TestNewRejectsNilRepository(t *testing.T) {
	_, err := New(nil)
	if err == nil {
		t.Fatal("expected error for nil repository")
	}
}

func TestGetProfile(t *testing.T) {
	repo := &stubRepo{
		getTopicsFunc: func(_ context.Context, visitorID string, limit int) ([]TopicScore, error) {
			if visitorID != "visitor-1" || limit != 10 {
				t.Fatalf("unexpected args: %q %d", visitorID, limit)
			}
			return []TopicScore{{TopicID: 7, Score: 3.5}}, nil
		},
		getRegionsFunc: func(_ context.Context, visitorID string, limit int) ([]RegionScore, error) {
			if visitorID != "visitor-1" || limit != 10 {
				t.Fatalf("unexpected args: %q %d", visitorID, limit)
			}
			return []RegionScore{{RegionID: 2, Score: 5}}, nil
		},
	}
	svc, _ := New(repo)

	profile, found, err := svc.GetProfile(context.Background(), "visitor-1")
	if err != nil {
		t.Fatalf("get profile: %v", err)
	}
	if !found || len(profile.Topics) != 1 || profile.Topics[0].TopicID != 7 ||
		len(profile.Regions) != 1 || profile.Regions[0].RegionID != 2 {
		t.Fatalf("unexpected result: found=%v profile=%v", found, profile)
	}
}

func TestGetProfileEmptyVisitorID(t *testing.T) {
	svc, _ := New(&stubRepo{
		getTopicsFunc: func(context.Context, string, int) ([]TopicScore, error) {
			t.Fatal("repo should not be called")
			return nil, nil
		},
	})

	profile, found, err := svc.GetProfile(context.Background(), "")
	if err != nil || found || profile != nil {
		t.Fatalf("unexpected result: profile=%v found=%v err=%v", profile, found, err)
	}
}

func TestTrackEvent(t *testing.T) {
	now := time.Date(2026, 5, 5, 13, 0, 0, 0, time.UTC)
	called := false
	regionCalled := false
	repo := &stubRepo{
		getTopicsFunc: func(context.Context, string, int) ([]TopicScore, error) { return nil, nil },
		trackTopicFunc: func(_ context.Context, visitorID string, topicID int, scoreDelta float64, gotNow time.Time) error {
			called = true
			if visitorID != "visitor-1" || topicID != 11 || scoreDelta != 5 || !gotNow.Equal(now.UTC()) {
				t.Fatalf("unexpected args: %q %d %v %v", visitorID, topicID, scoreDelta, gotNow)
			}
			return nil
		},
		trackRegionFunc: func(_ context.Context, visitorID string, regionID int, scoreDelta float64, gotNow time.Time) error {
			regionCalled = true
			if visitorID != "visitor-1" || regionID != 22 || scoreDelta != 5 || !gotNow.Equal(now.UTC()) {
				t.Fatalf("unexpected args: %q %d %v %v", visitorID, regionID, scoreDelta, gotNow)
			}
			return nil
		},
	}
	svc, _ := New(repo)
	svc.now = func() time.Time { return now }

	if err := svc.TrackEvent(context.Background(), "visitor-1", 11, 22, EventTypeClick); err != nil {
		t.Fatalf("track event: %v", err)
	}
	if !called || !regionCalled {
		t.Fatal("expected repository call")
	}
}

func TestTrackEventUnsupportedType(t *testing.T) {
	svc, _ := New(&stubRepo{
		getTopicsFunc:  func(context.Context, string, int) ([]TopicScore, error) { return nil, nil },
		trackTopicFunc: func(context.Context, string, int, float64, time.Time) error { return nil },
	})

	err := svc.TrackEvent(context.Background(), "visitor-1", 11, 0, "view")
	if err == nil || err.Error() == "" {
		t.Fatalf("expected unsupported event error, got %v", err)
	}
}

func TestTrackEventSkipsInvalidInputAndPropagatesRepoErrors(t *testing.T) {
	repoErr := errors.New("repo down")
	svc, _ := New(&stubRepo{
		getTopicsFunc: func(context.Context, string, int) ([]TopicScore, error) { return nil, nil },
		trackTopicFunc: func(context.Context, string, int, float64, time.Time) error {
			return repoErr
		},
	})

	if err := svc.TrackEvent(context.Background(), "", 11, 22, EventTypeClick); err != nil {
		t.Fatalf("expected nil for empty visitor, got %v", err)
	}
	if err := svc.TrackEvent(context.Background(), "visitor-1", 0, 0, EventTypeClick); err != nil {
		t.Fatalf("expected nil for zero topic and region, got %v", err)
	}
	if err := svc.TrackEvent(context.Background(), "visitor-1", 11, 0, EventTypeClick); !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}

func TestEventScoreDelta(t *testing.T) {
	if score, ok := eventScoreDelta(EventTypeClick); !ok || score != 5 {
		t.Fatalf("unexpected click score: %v %v", score, ok)
	}
	if score, ok := eventScoreDelta("other"); ok || score != 0 {
		t.Fatalf("unexpected unknown score: %v %v", score, ok)
	}
}
