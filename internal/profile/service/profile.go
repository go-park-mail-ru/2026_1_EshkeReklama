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

type RegionScore struct {
	RegionID int
	Score    float64
}

type Profile struct {
	Topics  []TopicScore
	Regions []RegionScore
}

type Repository interface {
	GetTopics(ctx context.Context, visitorID string, limit int) ([]TopicScore, error)
	GetRegions(ctx context.Context, visitorID string, limit int) ([]RegionScore, error)
	TrackTopic(ctx context.Context, visitorID string, topicID int, scoreDelta float64, now time.Time) error
	TrackRegion(ctx context.Context, visitorID string, regionID int, scoreDelta float64, now time.Time) error
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

func (s *Service) GetProfile(ctx context.Context, visitorID string) (*Profile, bool, error) {
	if visitorID == "" {
		return nil, false, nil
	}

	topics, err := s.repo.GetTopics(ctx, visitorID, 10)
	if err != nil {
		return nil, false, err
	}
	regions, err := s.repo.GetRegions(ctx, visitorID, 10)
	if err != nil {
		return nil, false, err
	}

	profile := &Profile{Topics: topics, Regions: regions}
	return profile, len(topics) > 0 || len(regions) > 0, nil
}

func (s *Service) TrackEvent(ctx context.Context, visitorID string, topicID, regionID int, eventType string) error {
	if visitorID == "" || (topicID <= 0 && regionID <= 0) {
		return nil
	}

	scoreDelta, ok := eventScoreDelta(eventType)
	if !ok {
		return fmt.Errorf("unsupported profile event type: %s", eventType)
	}

	now := s.now().UTC()
	if topicID > 0 {
		if err := s.repo.TrackTopic(ctx, visitorID, topicID, scoreDelta, now); err != nil {
			return err
		}
	}
	if regionID > 0 {
		if err := s.repo.TrackRegion(ctx, visitorID, regionID, scoreDelta, now); err != nil {
			return err
		}
	}
	return nil
}

func eventScoreDelta(eventType string) (float64, bool) {
	switch eventType {
	case EventTypeClick:
		return 5, true
	default:
		return 0, false
	}
}
