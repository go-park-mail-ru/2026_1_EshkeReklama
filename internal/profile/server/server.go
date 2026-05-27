package server

import (
	"context"

	profileservice "eshkere/internal/profile/service"
	profilev1 "eshkere/pkg/pb/profile/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
	GetProfile(ctx context.Context, visitorID string) (*profileservice.Profile, bool, error)
	TrackEvent(ctx context.Context, visitorID string, topicID, regionID int, eventType string) error
}

type Server struct {
	profilev1.UnimplementedProfileServiceServer
	service Service
}

func New(service Service) *Server {
	return &Server{service: service}
}

func (s *Server) GetProfile(ctx context.Context, req *profilev1.GetProfileRequest) (*profilev1.GetProfileResponse, error) {
	profile, found, err := s.service.GetProfile(ctx, req.GetVisitorId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &profilev1.GetProfileResponse{
		Found: found,
	}
	if profile != nil {
		resp.Topics = make([]*profilev1.TopicScore, 0, len(profile.Topics))
		for _, topic := range profile.Topics {
			resp.Topics = append(resp.Topics, &profilev1.TopicScore{
				TopicId: int32(topic.TopicID),
				Score:   topic.Score,
			})
		}
		resp.Regions = make([]*profilev1.RegionScore, 0, len(profile.Regions))
		for _, region := range profile.Regions {
			resp.Regions = append(resp.Regions, &profilev1.RegionScore{
				RegionId: int32(region.RegionID),
				Score:    region.Score,
			})
		}
	}
	return resp, nil
}

func (s *Server) TrackEvent(ctx context.Context, req *profilev1.TrackEventRequest) (*profilev1.TrackEventResponse, error) {
	if err := s.service.TrackEvent(ctx, req.GetVisitorId(), int(req.GetTopicId()), int(req.GetRegionId()), req.GetEventType()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &profilev1.TrackEventResponse{}, nil
}
