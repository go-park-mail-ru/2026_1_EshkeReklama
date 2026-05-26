package server

import (
	"context"
	"errors"
	"testing"

	profileservice "eshkere/internal/profile/service"
	profilev1 "eshkere/pkg/pb/profile/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type stubService struct {
	getProfileFunc func(context.Context, string) ([]profileservice.TopicScore, bool, error)
	trackEventFunc func(context.Context, string, int, string) error
}

func (s *stubService) GetProfile(ctx context.Context, visitorID string) ([]profileservice.TopicScore, bool, error) {
	return s.getProfileFunc(ctx, visitorID)
}

func (s *stubService) TrackEvent(ctx context.Context, visitorID string, topicID int, eventType string) error {
	return s.trackEventFunc(ctx, visitorID, topicID, eventType)
}

func TestGetProfile(t *testing.T) {
	srv := New(&stubService{
		getProfileFunc: func(_ context.Context, visitorID string) ([]profileservice.TopicScore, bool, error) {
			if visitorID != "visitor-1" {
				t.Fatalf("unexpected visitor id: %s", visitorID)
			}
			return []profileservice.TopicScore{{TopicID: 9, Score: 4.2}}, true, nil
		},
		trackEventFunc: func(context.Context, string, int, string) error { return nil },
	})

	resp, err := srv.GetProfile(context.Background(), &profilev1.GetProfileRequest{VisitorId: "visitor-1"})
	if err != nil {
		t.Fatalf("get profile: %v", err)
	}
	if !resp.Found || len(resp.Topics) != 1 || resp.Topics[0].TopicId != 9 {
		t.Fatalf("unexpected response: %#v", resp)
	}
}

func TestGetProfileErrorMappedToInternal(t *testing.T) {
	srv := New(&stubService{
		getProfileFunc: func(context.Context, string) ([]profileservice.TopicScore, bool, error) {
			return nil, false, errors.New("boom")
		},
		trackEventFunc: func(context.Context, string, int, string) error { return nil },
	})

	_, err := srv.GetProfile(context.Background(), &profilev1.GetProfileRequest{VisitorId: "visitor-1"})
	if status.Code(err) != codes.Internal {
		t.Fatalf("unexpected code: %v", status.Code(err))
	}
}

func TestTrackEvent(t *testing.T) {
	srv := New(&stubService{
		getProfileFunc: func(context.Context, string) ([]profileservice.TopicScore, bool, error) { return nil, false, nil },
		trackEventFunc: func(_ context.Context, visitorID string, topicID int, eventType string) error {
			if visitorID != "visitor-1" || topicID != 3 || eventType != "click" {
				t.Fatalf("unexpected args: %q %d %q", visitorID, topicID, eventType)
			}
			return nil
		},
	})

	resp, err := srv.TrackEvent(context.Background(), &profilev1.TrackEventRequest{
		VisitorId: "visitor-1",
		TopicId:   3,
		EventType: "click",
	})
	if err != nil {
		t.Fatalf("track event: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response")
	}
}

func TestTrackEventErrorMappedToInvalidArgument(t *testing.T) {
	srv := New(&stubService{
		getProfileFunc: func(context.Context, string) ([]profileservice.TopicScore, bool, error) { return nil, false, nil },
		trackEventFunc: func(context.Context, string, int, string) error {
			return errors.New("bad event")
		},
	})

	_, err := srv.TrackEvent(context.Background(), &profilev1.TrackEventRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("unexpected code: %v", status.Code(err))
	}
}
