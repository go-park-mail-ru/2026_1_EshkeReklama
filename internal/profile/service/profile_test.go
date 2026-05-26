package service

import (
	"context"
	"errors"
	"testing"
	"time"
)

type stubRepo struct {
	getTopicsFunc func(context.Context, string, int) ([]TopicScore, error)
	trackFunc     func(context.Context, string, int, float64, time.Time) error
}

func (s *stubRepo) GetTopics(ctx context.Context, visitorID string, limit int) ([]TopicScore, error) {
	return s.getTopicsFunc(ctx, visitorID, limit)
}

func (s *stubRepo) TrackTopic(ctx context.Context, visitorID string, topicID int, scoreDelta float64, now time.Time) error {
	return s.trackFunc(ctx, visitorID, topicID, scoreDelta, now)
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
	}
	svc, _ := New(repo)

	topics, found, err := svc.GetProfile(context.Background(), "visitor-1")
	if err != nil {
		t.Fatalf("get profile: %v", err)
	}
	if !found || len(topics) != 1 || topics[0].TopicID != 7 {
		t.Fatalf("unexpected result: found=%v topics=%v", found, topics)
	}
}

func TestGetProfileEmptyVisitorID(t *testing.T) {
	svc, _ := New(&stubRepo{
		getTopicsFunc: func(context.Context, string, int) ([]TopicScore, error) {
			t.Fatal("repo should not be called")
			return nil, nil
		},
	})

	topics, found, err := svc.GetProfile(context.Background(), "")
	if err != nil || found || topics != nil {
		t.Fatalf("unexpected result: topics=%v found=%v err=%v", topics, found, err)
	}
}

func TestTrackEvent(t *testing.T) {
	now := time.Date(2026, 5, 5, 13, 0, 0, 0, time.UTC)
	called := false
	repo := &stubRepo{
		getTopicsFunc: func(context.Context, string, int) ([]TopicScore, error) { return nil, nil },
		trackFunc: func(_ context.Context, visitorID string, topicID int, scoreDelta float64, gotNow time.Time) error {
			called = true
			if visitorID != "visitor-1" || topicID != 11 || scoreDelta != 5 || !gotNow.Equal(now.UTC()) {
				t.Fatalf("unexpected args: %q %d %v %v", visitorID, topicID, scoreDelta, gotNow)
			}
			return nil
		},
	}
	svc, _ := New(repo)
	svc.now = func() time.Time { return now }

	if err := svc.TrackEvent(context.Background(), "visitor-1", 11, EventTypeClick); err != nil {
		t.Fatalf("track event: %v", err)
	}
	if !called {
		t.Fatal("expected repository call")
	}
}

func TestTrackEventUnsupportedType(t *testing.T) {
	svc, _ := New(&stubRepo{
		getTopicsFunc: func(context.Context, string, int) ([]TopicScore, error) { return nil, nil },
		trackFunc:     func(context.Context, string, int, float64, time.Time) error { return nil },
	})

	err := svc.TrackEvent(context.Background(), "visitor-1", 11, "view")
	if err == nil || err.Error() == "" {
		t.Fatalf("expected unsupported event error, got %v", err)
	}
}

func TestTrackEventSkipsInvalidInputAndPropagatesRepoErrors(t *testing.T) {
	repoErr := errors.New("repo down")
	svc, _ := New(&stubRepo{
		getTopicsFunc: func(context.Context, string, int) ([]TopicScore, error) { return nil, nil },
		trackFunc: func(context.Context, string, int, float64, time.Time) error {
			return repoErr
		},
	})

	if err := svc.TrackEvent(context.Background(), "", 11, EventTypeClick); err != nil {
		t.Fatalf("expected nil for empty visitor, got %v", err)
	}
	if err := svc.TrackEvent(context.Background(), "visitor-1", 0, EventTypeClick); err != nil {
		t.Fatalf("expected nil for zero topic, got %v", err)
	}
	if err := svc.TrackEvent(context.Background(), "visitor-1", 11, EventTypeClick); !errors.Is(err, repoErr) {
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
