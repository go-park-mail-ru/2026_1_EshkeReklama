package profileclient

import (
	"context"
	"time"

	"eshkere/internal/service"

	profilev1 "eshkere/pkg/pb/profile/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const EventTypeClick = "click"

type Client struct {
	conn *grpc.ClientConn
	rpc  profilev1.ProfileServiceClient
}

func New(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn, rpc: profilev1.NewProfileServiceClient(conn)}, nil
}

func (c *Client) Close() error { return c.conn.Close() }

func (c *Client) GetProfile(ctx context.Context, visitorID string) (*service.Profile, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()

	resp, err := c.rpc.GetProfile(ctx, &profilev1.GetProfileRequest{VisitorId: visitorID})
	if err != nil {
		return nil, false, err
	}

	profile := &service.Profile{
		Topics:  make([]service.TopicScore, 0, len(resp.GetTopics())),
		Regions: make([]service.RegionScore, 0, len(resp.GetRegions())),
	}
	for _, topic := range resp.GetTopics() {
		profile.Topics = append(profile.Topics, service.TopicScore{
			TopicID: int(topic.GetTopicId()),
			Score:   topic.GetScore(),
		})
	}
	for _, region := range resp.GetRegions() {
		profile.Regions = append(profile.Regions, service.RegionScore{
			RegionID: int(region.GetRegionId()),
			Score:    region.GetScore(),
		})
	}

	return profile, resp.GetFound(), nil
}

func (c *Client) TrackEvent(ctx context.Context, visitorID string, topicID, regionID int, eventType string) error {
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	_, err := c.rpc.TrackEvent(ctx, &profilev1.TrackEventRequest{
		VisitorId: visitorID,
		TopicId:   int32(topicID),
		RegionId:  int32(regionID),
		EventType: eventType,
	})
	return err
}
