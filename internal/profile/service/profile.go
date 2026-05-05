package service

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const EventTypeClick = "click"

type TopicScore struct {
	TopicID int
	Score   float64
}

type Repository interface {
	GetTopics(ctx context.Context, visitorID string, limit int) ([]TopicScore, error)
	TrackTopic(ctx context.Context, visitorID string, topicID int, scoreDelta float64, now time.Time) error
}

type Service struct {
	repo Repository
	now  func() time.Time
}

func New(repo Repository) (*Service, error) {
	if repo == nil {
		return nil, errors.New("profile repository is nil")
	}
	return &Service{repo: repo, now: time.Now}, nil
}

func (s *Service) GetProfile(ctx context.Context, visitorID string) ([]TopicScore, bool, error) {
	if visitorID == "" {
		return nil, false, nil
	}

	topics, err := s.repo.GetTopics(ctx, visitorID, 10)
	if err != nil {
		return nil, false, err
	}
	return topics, len(topics) > 0, nil
}

func (s *Service) TrackEvent(ctx context.Context, visitorID string, topicID int, eventType string) error {
	if visitorID == "" || topicID <= 0 {
		return nil
	}

	scoreDelta, ok := eventScoreDelta(eventType)
	if !ok {
		return fmt.Errorf("unsupported profile event type: %s", eventType)
	}

	return s.repo.TrackTopic(ctx, visitorID, topicID, scoreDelta, s.now().UTC())
}

func eventScoreDelta(eventType string) (float64, bool) {
	switch eventType {
	case EventTypeClick:
		return 5, true
	default:
		return 0, false
	}
}
