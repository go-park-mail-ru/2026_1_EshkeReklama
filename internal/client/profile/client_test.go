package profileclient

import (
	"context"
	"net"
	"testing"
	"time"

	profilev1 "eshkere/pkg/pb/profile/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type profileTestServer struct {
	profilev1.UnimplementedProfileServiceServer
}

func (s *profileTestServer) GetProfile(context.Context, *profilev1.GetProfileRequest) (*profilev1.GetProfileResponse, error) {
	return &profilev1.GetProfileResponse{
		Found: true,
		Topics: []*profilev1.TopicScore{
			{TopicId: 3, Score: 7.5},
		},
	}, nil
}

func (s *profileTestServer) TrackEvent(context.Context, *profilev1.TrackEventRequest) (*profilev1.TrackEventResponse, error) {
	return &profilev1.TrackEventResponse{}, nil
}

func TestProfileClient(t *testing.T) {
	conn := newProfileBufConn(t, &profileTestServer{})
	client := &Client{conn: conn, rpc: profilev1.NewProfileServiceClient(conn)}
	defer client.Close()

	topics, found, err := client.GetProfile(context.Background(), "visitor-1")
	if err != nil || !found || len(topics) != 1 || topics[0].TopicID != 3 {
		t.Fatalf("get profile: topics=%v found=%v err=%v", topics, found, err)
	}

	if err := client.TrackEvent(context.Background(), "visitor-1", 3, EventTypeClick); err != nil {
		t.Fatalf("track event: %v", err)
	}
}

func newProfileBufConn(t *testing.T, srv profilev1.ProfileServiceServer) *grpc.ClientConn {
	t.Helper()
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	profilev1.RegisterProfileServiceServer(server, srv)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(cancel)
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial bufconn: %v", err)
	}
	return conn
}
