package authv1

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type stubAuthConn struct {
	lastMethod string
	lastReq    interface{}
	invokeErr  error
}

func (c *stubAuthConn) Invoke(_ context.Context, method string, req, reply interface{}, _ ...grpc.CallOption) error {
	c.lastMethod = method
	c.lastReq = req
	if c.invokeErr != nil {
		return c.invokeErr
	}

	switch out := reply.(type) {
	case *RegisterResponse:
		*out = RegisterResponse{AdvertiserId: 1, SessionId: "reg", ExpiresAt: 10}
	case *LoginResponse:
		*out = LoginResponse{AdvertiserId: 2, SessionId: "login", ExpiresAt: 20}
	case *ValidateSessionResponse:
		*out = ValidateSessionResponse{AdvertiserId: 3, ExpiresAt: 30}
	case *LogoutResponse:
		*out = LogoutResponse{}
	case *GetCredentialsResponse:
		*out = GetCredentialsResponse{AdvertiserId: 4, Email: "a@example.com", Phone: "900"}
	case *UpdateCredentialsResponse:
		*out = UpdateCredentialsResponse{AdvertiserId: 5, Email: "b@example.com", Phone: "901"}
	}

	return nil
}

func (c *stubAuthConn) NewStream(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, errors.New("unexpected stream")
}

type authUnaryServer struct {
	UnimplementedAuthServiceServer
	lastMethod string
	lastReq    interface{}
}

func (s *authUnaryServer) Register(_ context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	s.lastMethod = "Register"
	s.lastReq = req
	return &RegisterResponse{AdvertiserId: 10}, nil
}

func (s *authUnaryServer) Login(_ context.Context, req *LoginRequest) (*LoginResponse, error) {
	s.lastMethod = "Login"
	s.lastReq = req
	return &LoginResponse{AdvertiserId: 11}, nil
}

func (s *authUnaryServer) LoginVKID(_ context.Context, req *LoginVKIDRequest) (*LoginResponse, error) {
	s.lastMethod = "LoginVKID"
	s.lastReq = req
	return &LoginResponse{AdvertiserId: 12}, nil
}

func (s *authUnaryServer) ValidateSession(_ context.Context, req *ValidateSessionRequest) (*ValidateSessionResponse, error) {
	s.lastMethod = "ValidateSession"
	s.lastReq = req
	return &ValidateSessionResponse{AdvertiserId: 13}, nil
}

func (s *authUnaryServer) Logout(_ context.Context, req *LogoutRequest) (*LogoutResponse, error) {
	s.lastMethod = "Logout"
	s.lastReq = req
	return &LogoutResponse{}, nil
}

func (s *authUnaryServer) GetCredentials(_ context.Context, req *GetCredentialsRequest) (*GetCredentialsResponse, error) {
	s.lastMethod = "GetCredentials"
	s.lastReq = req
	return &GetCredentialsResponse{AdvertiserId: req.GetAdvertiserId(), Email: "mail", Phone: "phone"}, nil
}

func (s *authUnaryServer) UpdateCredentials(_ context.Context, req *UpdateCredentialsRequest) (*UpdateCredentialsResponse, error) {
	s.lastMethod = "UpdateCredentials"
	s.lastReq = req
	return &UpdateCredentialsResponse{AdvertiserId: req.GetAdvertiserId(), Email: req.GetEmail(), Phone: req.GetPhone()}, nil
}

type stubRegistrar struct {
	desc *grpc.ServiceDesc
	srv  interface{}
}

func (r *stubRegistrar) RegisterService(desc *grpc.ServiceDesc, srv interface{}) {
	r.desc = desc
	r.srv = srv
}

func TestAuthMessagesGeneratedMethods(t *testing.T) {
	t.Run("register request", func(t *testing.T) {
		msg := &RegisterRequest{Email: "a@example.com", Phone: "900", Password: "secret"}
		if msg.GetEmail() != "a@example.com" || msg.GetPhone() != "900" || msg.GetPassword() != "secret" {
			t.Fatalf("unexpected getters: %+v", msg)
		}
		if got := msg.String(); got == "" {
			t.Fatal("String returned empty")
		}
		if got := msg.ProtoReflect().Descriptor().FullName(); got != "eshkere.auth.v1.RegisterRequest" {
			t.Fatalf("unexpected full name: %s", got)
		}
		if raw, idx := msg.Descriptor(); len(raw) == 0 || len(idx) != 1 || idx[0] != 0 {
			t.Fatalf("unexpected descriptor: len=%d idx=%v", len(raw), idx)
		}
		msg.Reset()
		if msg.GetEmail() != "" || msg.GetPhone() != "" || msg.GetPassword() != "" {
			t.Fatalf("Reset did not clear message: %+v", msg)
		}
		var nilMsg *RegisterRequest
		if nilMsg.GetEmail() != "" || nilMsg.GetPhone() != "" || nilMsg.GetPassword() != "" || nilMsg.ProtoReflect() == nil {
			t.Fatal("nil RegisterRequest accessors failed")
		}
	})

	t.Run("register response", func(t *testing.T) {
		msg := &RegisterResponse{AdvertiserId: 1, SessionId: "sid", ExpiresAt: 2}
		if msg.GetAdvertiserId() != 1 || msg.GetSessionId() != "sid" || msg.GetExpiresAt() != 2 {
			t.Fatalf("unexpected getters: %+v", msg)
		}
		_, _ = msg.Descriptor()
		_ = msg.String()
		_ = msg.ProtoReflect()
		msg.Reset()
		var nilMsg *RegisterResponse
		if nilMsg.GetAdvertiserId() != 0 || nilMsg.GetSessionId() != "" || nilMsg.GetExpiresAt() != 0 || nilMsg.ProtoReflect() == nil {
			t.Fatal("nil RegisterResponse accessors failed")
		}
	})

	t.Run("login request", func(t *testing.T) {
		msg := &LoginRequest{Identifier: "mail", Password: "secret"}
		if msg.GetIdentifier() != "mail" || msg.GetPassword() != "secret" {
			t.Fatalf("unexpected getters: %+v", msg)
		}
		_, _ = msg.Descriptor()
		_ = msg.String()
		_ = msg.ProtoReflect()
		msg.Reset()
		var nilMsg *LoginRequest
		if nilMsg.GetIdentifier() != "" || nilMsg.GetPassword() != "" || nilMsg.ProtoReflect() == nil {
			t.Fatal("nil LoginRequest accessors failed")
		}
	})

	t.Run("login vkid request", func(t *testing.T) {
		msg := &LoginVKIDRequest{Code: "code", DeviceId: "device", CodeVerifier: "verifier"}
		if msg.GetCode() != "code" || msg.GetDeviceId() != "device" || msg.GetCodeVerifier() != "verifier" {
			t.Fatalf("unexpected getters: %+v", msg)
		}
		_, _ = msg.Descriptor()
		_ = msg.String()
		_ = msg.ProtoReflect()
		msg.Reset()
		var nilMsg *LoginVKIDRequest
		if nilMsg.GetCode() != "" || nilMsg.GetDeviceId() != "" || nilMsg.GetCodeVerifier() != "" || nilMsg.ProtoReflect() == nil {
			t.Fatal("nil LoginVKIDRequest accessors failed")
		}
	})

	t.Run("login response", func(t *testing.T) {
		msg := &LoginResponse{AdvertiserId: 2, SessionId: "sid", ExpiresAt: 3}
		if msg.GetAdvertiserId() != 2 || msg.GetSessionId() != "sid" || msg.GetExpiresAt() != 3 {
			t.Fatalf("unexpected getters: %+v", msg)
		}
		_, _ = msg.Descriptor()
		_ = msg.String()
		_ = msg.ProtoReflect()
		msg.Reset()
		var nilMsg *LoginResponse
		if nilMsg.GetAdvertiserId() != 0 || nilMsg.GetSessionId() != "" || nilMsg.GetExpiresAt() != 0 || nilMsg.ProtoReflect() == nil {
			t.Fatal("nil LoginResponse accessors failed")
		}
	})

	t.Run("validate session", func(t *testing.T) {
		req := &ValidateSessionRequest{SessionId: "sid"}
		if req.GetSessionId() != "sid" {
			t.Fatalf("unexpected session id: %q", req.GetSessionId())
		}
		_, _ = req.Descriptor()
		_ = req.String()
		_ = req.ProtoReflect()
		req.Reset()
		resp := &ValidateSessionResponse{AdvertiserId: 4, ExpiresAt: 5}
		if resp.GetAdvertiserId() != 4 || resp.GetExpiresAt() != 5 {
			t.Fatalf("unexpected response getters: %+v", resp)
		}
		_, _ = resp.Descriptor()
		_ = resp.String()
		_ = resp.ProtoReflect()
		resp.Reset()
		var nilReq *ValidateSessionRequest
		var nilResp *ValidateSessionResponse
		if nilReq.GetSessionId() != "" || nilReq.ProtoReflect() == nil || nilResp.GetAdvertiserId() != 0 || nilResp.GetExpiresAt() != 0 || nilResp.ProtoReflect() == nil {
			t.Fatal("nil ValidateSession accessors failed")
		}
	})

	t.Run("logout request and response", func(t *testing.T) {
		req := &LogoutRequest{SessionId: "sid"}
		if req.GetSessionId() != "sid" {
			t.Fatalf("unexpected session id: %q", req.GetSessionId())
		}
		_, _ = req.Descriptor()
		_ = req.String()
		_ = req.ProtoReflect()
		req.Reset()
		resp := &LogoutResponse{}
		_, _ = resp.Descriptor()
		_ = resp.String()
		_ = resp.ProtoReflect()
		resp.Reset()
		var nilReq *LogoutRequest
		var nilResp *LogoutResponse
		if nilReq.GetSessionId() != "" || nilReq.ProtoReflect() == nil || nilResp.ProtoReflect() == nil {
			t.Fatal("nil Logout accessors failed")
		}
	})

	t.Run("credentials messages", func(t *testing.T) {
		req := &GetCredentialsRequest{AdvertiserId: 6}
		if req.GetAdvertiserId() != 6 {
			t.Fatalf("unexpected advertiser id: %d", req.GetAdvertiserId())
		}
		_, _ = req.Descriptor()
		_ = req.String()
		_ = req.ProtoReflect()
		req.Reset()

		resp := &GetCredentialsResponse{AdvertiserId: 7, Email: "e", Phone: "p"}
		if resp.GetAdvertiserId() != 7 || resp.GetEmail() != "e" || resp.GetPhone() != "p" {
			t.Fatalf("unexpected credentials response: %+v", resp)
		}
		_, _ = resp.Descriptor()
		_ = resp.String()
		_ = resp.ProtoReflect()
		resp.Reset()

		updateReq := &UpdateCredentialsRequest{AdvertiserId: 8, Email: "e2", Phone: "p2"}
		if updateReq.GetAdvertiserId() != 8 || updateReq.GetEmail() != "e2" || updateReq.GetPhone() != "p2" {
			t.Fatalf("unexpected update request: %+v", updateReq)
		}
		_, _ = updateReq.Descriptor()
		_ = updateReq.String()
		_ = updateReq.ProtoReflect()
		updateReq.Reset()

		updateResp := &UpdateCredentialsResponse{AdvertiserId: 9, Email: "e3", Phone: "p3"}
		if updateResp.GetAdvertiserId() != 9 || updateResp.GetEmail() != "e3" || updateResp.GetPhone() != "p3" {
			t.Fatalf("unexpected update response: %+v", updateResp)
		}
		_, _ = updateResp.Descriptor()
		_ = updateResp.String()
		_ = updateResp.ProtoReflect()
		updateResp.Reset()

		var nilReq *GetCredentialsRequest
		var nilResp *GetCredentialsResponse
		var nilUpdateReq *UpdateCredentialsRequest
		var nilUpdateResp *UpdateCredentialsResponse
		if nilReq.GetAdvertiserId() != 0 || nilReq.ProtoReflect() == nil {
			t.Fatal("nil GetCredentialsRequest accessors failed")
		}
		if nilResp.GetAdvertiserId() != 0 || nilResp.GetEmail() != "" || nilResp.GetPhone() != "" || nilResp.ProtoReflect() == nil {
			t.Fatal("nil GetCredentialsResponse accessors failed")
		}
		if nilUpdateReq.GetAdvertiserId() != 0 || nilUpdateReq.GetEmail() != "" || nilUpdateReq.GetPhone() != "" || nilUpdateReq.ProtoReflect() == nil {
			t.Fatal("nil UpdateCredentialsRequest accessors failed")
		}
		if nilUpdateResp.GetAdvertiserId() != 0 || nilUpdateResp.GetEmail() != "" || nilUpdateResp.GetPhone() != "" || nilUpdateResp.ProtoReflect() == nil {
			t.Fatal("nil UpdateCredentialsResponse accessors failed")
		}
	})
}

func TestAuthServiceClientGeneratedMethods(t *testing.T) {
	conn := &stubAuthConn{}
	client := NewAuthServiceClient(conn)
	ctx := context.Background()

	if resp, err := client.Register(ctx, &RegisterRequest{Email: "a"}); err != nil || resp.GetAdvertiserId() != 1 || conn.lastMethod != AuthService_Register_FullMethodName {
		t.Fatalf("Register failed: resp=%+v err=%v method=%s", resp, err, conn.lastMethod)
	}
	if resp, err := client.Login(ctx, &LoginRequest{Identifier: "id"}); err != nil || resp.GetAdvertiserId() != 2 || conn.lastMethod != AuthService_Login_FullMethodName {
		t.Fatalf("Login failed: resp=%+v err=%v method=%s", resp, err, conn.lastMethod)
	}
	if resp, err := client.LoginVKID(ctx, &LoginVKIDRequest{Code: "code"}); err != nil || resp.GetAdvertiserId() != 2 || conn.lastMethod != AuthService_LoginVKID_FullMethodName {
		t.Fatalf("LoginVKID failed: resp=%+v err=%v method=%s", resp, err, conn.lastMethod)
	}
	if resp, err := client.ValidateSession(ctx, &ValidateSessionRequest{SessionId: "sid"}); err != nil || resp.GetAdvertiserId() != 3 || conn.lastMethod != AuthService_ValidateSession_FullMethodName {
		t.Fatalf("ValidateSession failed: resp=%+v err=%v method=%s", resp, err, conn.lastMethod)
	}
	if _, err := client.Logout(ctx, &LogoutRequest{SessionId: "sid"}); err != nil || conn.lastMethod != AuthService_Logout_FullMethodName {
		t.Fatalf("Logout failed: err=%v method=%s", err, conn.lastMethod)
	}
	if resp, err := client.GetCredentials(ctx, &GetCredentialsRequest{AdvertiserId: 4}); err != nil || resp.GetAdvertiserId() != 4 || conn.lastMethod != AuthService_GetCredentials_FullMethodName {
		t.Fatalf("GetCredentials failed: resp=%+v err=%v method=%s", resp, err, conn.lastMethod)
	}
	if resp, err := client.UpdateCredentials(ctx, &UpdateCredentialsRequest{AdvertiserId: 5}); err != nil || resp.GetAdvertiserId() != 5 || conn.lastMethod != AuthService_UpdateCredentials_FullMethodName {
		t.Fatalf("UpdateCredentials failed: resp=%+v err=%v method=%s", resp, err, conn.lastMethod)
	}

	conn.invokeErr = errors.New("boom")
	if _, err := client.Login(ctx, &LoginRequest{}); !errors.Is(err, conn.invokeErr) {
		t.Fatalf("expected invoke error, got %v", err)
	}
}

func TestAuthServiceServerGeneratedMethods(t *testing.T) {
	srv := &authUnaryServer{}
	registrar := &stubRegistrar{}

	RegisterAuthServiceServer(registrar, srv)
	if registrar.desc == nil || registrar.desc.ServiceName != "eshkere.auth.v1.AuthService" || registrar.srv != srv {
		t.Fatalf("unexpected registration: %+v", registrar.desc)
	}
	if len(AuthService_ServiceDesc.Methods) != 7 || AuthService_ServiceDesc.Metadata != "proto/auth/v1/auth.proto" {
		t.Fatalf("unexpected service desc: %+v", AuthService_ServiceDesc)
	}

	ctx := context.Background()
	decodeRegister := func(v interface{}) error {
		*v.(*RegisterRequest) = RegisterRequest{Email: "a@example.com"}
		return nil
	}
	if resp, err := _AuthService_Register_Handler(srv, ctx, decodeRegister, nil); err != nil || resp.(*RegisterResponse).GetAdvertiserId() != 10 || srv.lastMethod != "Register" {
		t.Fatalf("register handler failed: resp=%+v err=%v method=%s", resp, err, srv.lastMethod)
	}
	if _, err := _AuthService_Register_Handler(srv, ctx, func(interface{}) error { return errors.New("decode") }, nil); err == nil {
		t.Fatal("expected register decode error")
	}

	decodeLogin := func(v interface{}) error {
		*v.(*LoginRequest) = LoginRequest{Identifier: "mail"}
		return nil
	}
	if resp, err := _AuthService_Login_Handler(srv, ctx, decodeLogin, func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if info.FullMethod != AuthService_Login_FullMethodName {
			t.Fatalf("unexpected method: %s", info.FullMethod)
		}
		return handler(ctx, req)
	}); err != nil || resp.(*LoginResponse).GetAdvertiserId() != 11 || srv.lastMethod != "Login" {
		t.Fatalf("login handler failed: resp=%+v err=%v method=%s", resp, err, srv.lastMethod)
	}

	if resp, err := _AuthService_LoginVKID_Handler(srv, ctx, func(v interface{}) error {
		*v.(*LoginVKIDRequest) = LoginVKIDRequest{Code: "code"}
		return nil
	}, nil); err != nil || resp.(*LoginResponse).GetAdvertiserId() != 12 || srv.lastMethod != "LoginVKID" {
		t.Fatalf("login vkid handler failed: resp=%+v err=%v method=%s", resp, err, srv.lastMethod)
	}

	if resp, err := _AuthService_ValidateSession_Handler(srv, ctx, func(v interface{}) error {
		*v.(*ValidateSessionRequest) = ValidateSessionRequest{SessionId: "sid"}
		return nil
	}, func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if info.FullMethod != AuthService_ValidateSession_FullMethodName {
			t.Fatalf("unexpected method: %s", info.FullMethod)
		}
		return handler(ctx, req)
	}); err != nil || resp.(*ValidateSessionResponse).GetAdvertiserId() != 13 || srv.lastMethod != "ValidateSession" {
		t.Fatalf("validate handler failed: resp=%+v err=%v method=%s", resp, err, srv.lastMethod)
	}

	if _, err := _AuthService_Logout_Handler(srv, ctx, func(v interface{}) error {
		*v.(*LogoutRequest) = LogoutRequest{SessionId: "sid"}
		return nil
	}, nil); err != nil || srv.lastMethod != "Logout" {
		t.Fatalf("logout handler failed: err=%v method=%s", err, srv.lastMethod)
	}

	if resp, err := _AuthService_GetCredentials_Handler(srv, ctx, func(v interface{}) error {
		*v.(*GetCredentialsRequest) = GetCredentialsRequest{AdvertiserId: 99}
		return nil
	}, nil); err != nil || resp.(*GetCredentialsResponse).GetAdvertiserId() != 99 || srv.lastMethod != "GetCredentials" {
		t.Fatalf("get credentials handler failed: resp=%+v err=%v method=%s", resp, err, srv.lastMethod)
	}

	if resp, err := _AuthService_UpdateCredentials_Handler(srv, ctx, func(v interface{}) error {
		*v.(*UpdateCredentialsRequest) = UpdateCredentialsRequest{AdvertiserId: 77, Email: "x", Phone: "y"}
		return nil
	}, nil); err != nil || resp.(*UpdateCredentialsResponse).GetAdvertiserId() != 77 || srv.lastMethod != "UpdateCredentials" {
		t.Fatalf("update credentials handler failed: resp=%+v err=%v method=%s", resp, err, srv.lastMethod)
	}
}

func TestAuthUnimplementedServer(t *testing.T) {
	srv := UnimplementedAuthServiceServer{}
	ctx := context.Background()

	cases := []struct {
		name string
		call func() error
	}{
		{"Register", func() error { _, err := srv.Register(ctx, &RegisterRequest{}); return err }},
		{"Login", func() error { _, err := srv.Login(ctx, &LoginRequest{}); return err }},
		{"LoginVKID", func() error { _, err := srv.LoginVKID(ctx, &LoginVKIDRequest{}); return err }},
		{"ValidateSession", func() error { _, err := srv.ValidateSession(ctx, &ValidateSessionRequest{}); return err }},
		{"Logout", func() error { _, err := srv.Logout(ctx, &LogoutRequest{}); return err }},
		{"GetCredentials", func() error { _, err := srv.GetCredentials(ctx, &GetCredentialsRequest{}); return err }},
		{"UpdateCredentials", func() error { _, err := srv.UpdateCredentials(ctx, &UpdateCredentialsRequest{}); return err }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.call()
			if status.Code(err) != codes.Unimplemented {
				t.Fatalf("expected Unimplemented, got %v", err)
			}
		})
	}

	srv.mustEmbedUnimplementedAuthServiceServer()
	srv.testEmbeddedByValue()
}

func TestAuthGeneratedConstants(t *testing.T) {
	if AuthService_Register_FullMethodName == "" ||
		AuthService_Login_FullMethodName == "" ||
		AuthService_LoginVKID_FullMethodName == "" ||
		AuthService_ValidateSession_FullMethodName == "" ||
		AuthService_Logout_FullMethodName == "" ||
		AuthService_GetCredentials_FullMethodName == "" ||
		AuthService_UpdateCredentials_FullMethodName == "" {
		t.Fatal("full method names must be set")
	}

	var _ AuthServiceClient = NewAuthServiceClient(&stubAuthConn{})
	var _ UnsafeAuthServiceServer = (*authUnaryServer)(nil)

	_ = metadata.New(nil)
}
