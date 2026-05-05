package profilev1

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type stubProfileConn struct {
	lastMethod string
	invokeErr  error
}

func (c *stubProfileConn) Invoke(_ context.Context, method string, _ interface{}, reply interface{}, _ ...grpc.CallOption) error {
	c.lastMethod = method
	if c.invokeErr != nil {
		return c.invokeErr
	}

	switch out := reply.(type) {
	case *GetProfileResponse:
		*out = GetProfileResponse{Found: true, Topics: []*TopicScore{{TopicId: 1, Score: 0.75}}}
	case *TrackEventResponse:
		*out = TrackEventResponse{}
	}

	return nil
}

func (c *stubProfileConn) NewStream(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, errors.New("unexpected stream")
}

type profileUnaryServer struct {
	UnimplementedProfileServiceServer
	lastMethod string
	lastReq    interface{}
}

func (s *profileUnaryServer) GetProfile(_ context.Context, req *GetProfileRequest) (*GetProfileResponse, error) {
	s.lastMethod = "GetProfile"
	s.lastReq = req
	return &GetProfileResponse{Found: true, Topics: []*TopicScore{{TopicId: 2, Score: 0.5}}}, nil
}

func (s *profileUnaryServer) TrackEvent(_ context.Context, req *TrackEventRequest) (*TrackEventResponse, error) {
	s.lastMethod = "TrackEvent"
	s.lastReq = req
	return &TrackEventResponse{}, nil
}

type profileRegistrar struct {
	desc *grpc.ServiceDesc
	srv  interface{}
}

func (r *profileRegistrar) RegisterService(desc *grpc.ServiceDesc, srv interface{}) {
	r.desc = desc
	r.srv = srv
}

func TestProfileMessagesGeneratedMethods(t *testing.T) {
	req := &GetProfileRequest{VisitorId: "visitor"}
	if req.GetVisitorId() != "visitor" {
		t.Fatalf("unexpected visitor id: %q", req.GetVisitorId())
	}
	_, _ = req.Descriptor()
	_ = req.String()
	_ = req.ProtoReflect()
	req.Reset()
	var nilReq *GetProfileRequest
	if nilReq.GetVisitorId() != "" || nilReq.ProtoReflect() == nil {
		t.Fatal("nil GetProfileRequest accessors failed")
	}

	topic := &TopicScore{TopicId: 7, Score: 0.9}
	if topic.GetTopicId() != 7 || topic.GetScore() != 0.9 {
		t.Fatalf("unexpected topic getters: %+v", topic)
	}
	_, _ = topic.Descriptor()
	_ = topic.String()
	_ = topic.ProtoReflect()
	topic.Reset()
	var nilTopic *TopicScore
	if nilTopic.GetTopicId() != 0 || nilTopic.GetScore() != 0 || nilTopic.ProtoReflect() == nil {
		t.Fatal("nil TopicScore accessors failed")
	}

	resp := &GetProfileResponse{Found: true, Topics: []*TopicScore{{TopicId: 1, Score: 0.2}}}
	if !resp.GetFound() || len(resp.GetTopics()) != 1 {
		t.Fatalf("unexpected profile response: %+v", resp)
	}
	_, _ = resp.Descriptor()
	_ = resp.String()
	_ = resp.ProtoReflect()
	resp.Reset()
	var nilResp *GetProfileResponse
	if nilResp.GetFound() || nilResp.GetTopics() != nil || nilResp.ProtoReflect() == nil {
		t.Fatal("nil GetProfileResponse accessors failed")
	}

	eventReq := &TrackEventRequest{VisitorId: "v", TopicId: 3, EventType: "click"}
	if eventReq.GetVisitorId() != "v" || eventReq.GetTopicId() != 3 || eventReq.GetEventType() != "click" {
		t.Fatalf("unexpected track event request: %+v", eventReq)
	}
	_, _ = eventReq.Descriptor()
	_ = eventReq.String()
	_ = eventReq.ProtoReflect()
	eventReq.Reset()
	var nilEventReq *TrackEventRequest
	if nilEventReq.GetVisitorId() != "" || nilEventReq.GetTopicId() != 0 || nilEventReq.GetEventType() != "" || nilEventReq.ProtoReflect() == nil {
		t.Fatal("nil TrackEventRequest accessors failed")
	}

	eventResp := &TrackEventResponse{}
	_, _ = eventResp.Descriptor()
	_ = eventResp.String()
	_ = eventResp.ProtoReflect()
	eventResp.Reset()
	var nilEventResp *TrackEventResponse
	if nilEventResp.ProtoReflect() == nil {
		t.Fatal("nil TrackEventResponse accessors failed")
	}
}

func TestProfileServiceClientGeneratedMethods(t *testing.T) {
	conn := &stubProfileConn{}
	client := NewProfileServiceClient(conn)
	ctx := context.Background()

	resp, err := client.GetProfile(ctx, &GetProfileRequest{VisitorId: "visitor"})
	if err != nil || !resp.GetFound() || conn.lastMethod != "/eshkere.profile.v1.ProfileService/GetProfile" {
		t.Fatalf("GetProfile failed: resp=%+v err=%v method=%s", resp, err, conn.lastMethod)
	}

	if _, err := client.TrackEvent(ctx, &TrackEventRequest{VisitorId: "visitor"}); err != nil || conn.lastMethod != "/eshkere.profile.v1.ProfileService/TrackEvent" {
		t.Fatalf("TrackEvent failed: err=%v method=%s", err, conn.lastMethod)
	}

	conn.invokeErr = errors.New("boom")
	if _, err := client.GetProfile(ctx, &GetProfileRequest{}); !errors.Is(err, conn.invokeErr) {
		t.Fatalf("expected invoke error, got %v", err)
	}
}

func TestProfileServiceServerGeneratedMethods(t *testing.T) {
	srv := &profileUnaryServer{}
	registrar := &profileRegistrar{}
	RegisterProfileServiceServer(registrar, srv)
	if registrar.desc == nil || registrar.desc.ServiceName != "eshkere.profile.v1.ProfileService" || registrar.srv != srv {
		t.Fatalf("unexpected registration: %+v", registrar.desc)
	}
	if len(ProfileService_ServiceDesc.Methods) != 2 || ProfileService_ServiceDesc.Metadata != "proto/profile/v1/profile.proto" {
		t.Fatalf("unexpected service desc: %+v", ProfileService_ServiceDesc)
	}

	ctx := context.Background()
	resp, err := _ProfileService_GetProfile_Handler(srv, ctx, func(v interface{}) error {
		*v.(*GetProfileRequest) = GetProfileRequest{VisitorId: "visitor"}
		return nil
	}, func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if info.FullMethod != "/eshkere.profile.v1.ProfileService/GetProfile" {
			t.Fatalf("unexpected method: %s", info.FullMethod)
		}
		return handler(ctx, req)
	})
	if err != nil || !resp.(*GetProfileResponse).GetFound() || srv.lastMethod != "GetProfile" {
		t.Fatalf("get profile handler failed: resp=%+v err=%v method=%s", resp, err, srv.lastMethod)
	}

	if _, err := _ProfileService_GetProfile_Handler(srv, ctx, func(interface{}) error { return errors.New("decode") }, nil); err == nil {
		t.Fatal("expected decode error")
	}

	if _, err := _ProfileService_TrackEvent_Handler(srv, ctx, func(v interface{}) error {
		*v.(*TrackEventRequest) = TrackEventRequest{VisitorId: "visitor", TopicId: 1, EventType: "click"}
		return nil
	}, nil); err != nil || srv.lastMethod != "TrackEvent" {
		t.Fatalf("track event handler failed: err=%v method=%s", err, srv.lastMethod)
	}
}

func TestProfileUnimplementedServer(t *testing.T) {
	srv := UnimplementedProfileServiceServer{}
	ctx := context.Background()

	if _, err := srv.GetProfile(ctx, &GetProfileRequest{}); status.Code(err) != codes.Unimplemented {
		t.Fatalf("expected unimplemented GetProfile, got %v", err)
	}
	if _, err := srv.TrackEvent(ctx, &TrackEventRequest{}); status.Code(err) != codes.Unimplemented {
		t.Fatalf("expected unimplemented TrackEvent, got %v", err)
	}
	srv.mustEmbedUnimplementedProfileServiceServer()
}
