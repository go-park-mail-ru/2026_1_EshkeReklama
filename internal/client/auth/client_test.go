package authclient

import (
	"context"
	"net"
	"testing"
	"time"

	errs "eshkere/internal/errors"
	authv1 "eshkere/pkg/pb/auth/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type authTestServer struct {
	authv1.UnimplementedAuthServiceServer
}

func (s *authTestServer) Register(context.Context, *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	return &authv1.RegisterResponse{AdvertiserId: 11, SessionId: "sid-1", ExpiresAt: 123}, nil
}
func (s *authTestServer) Login(context.Context, *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	return nil, status.Error(codes.Unauthenticated, "invalid credentials")
}
func (s *authTestServer) LoginVKID(context.Context, *authv1.LoginVKIDRequest) (*authv1.LoginResponse, error) {
	return nil, status.Error(codes.AlreadyExists, "vk credentials conflict")
}
func (s *authTestServer) ValidateSession(context.Context, *authv1.ValidateSessionRequest) (*authv1.ValidateSessionResponse, error) {
	return nil, status.Error(codes.NotFound, "session not found")
}
func (s *authTestServer) Logout(context.Context, *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	return &authv1.LogoutResponse{}, nil
}
func (s *authTestServer) GetCredentials(context.Context, *authv1.GetCredentialsRequest) (*authv1.GetCredentialsResponse, error) {
	return &authv1.GetCredentialsResponse{Email: "user@example.com", Phone: "9001234567"}, nil
}
func (s *authTestServer) UpdateCredentials(context.Context, *authv1.UpdateCredentialsRequest) (*authv1.UpdateCredentialsResponse, error) {
	return nil, status.Error(codes.InvalidArgument, "bad arg")
}

func TestClientCallsAndMapsErrors(t *testing.T) {
	conn := newAuthBufConn(t, &authTestServer{})
	client := &Client{conn: conn, rpc: authv1.NewAuthServiceClient(conn)}
	defer client.Close()

	id, sessionID, expiresAt, err := client.Register(context.Background(), "user@example.com", "9001234567", "secret123")
	if err != nil || id != 11 || sessionID != "sid-1" || expiresAt != 123 {
		t.Fatalf("register: %d %q %d %v", id, sessionID, expiresAt, err)
	}

	if _, _, _, err := client.Login(context.Background(), "user@example.com", "bad"); err != errs.ErrInvalidCredentials {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
	if _, _, _, _, _, err := client.LoginVKID(context.Background(), "vk-token", 123); err != errs.ErrVKIDConflict {
		t.Fatalf("expected vkid conflict, got %v", err)
	}
	if _, err := client.ValidateSession(context.Background(), "sid"); err != errs.ErrSessionNotFound {
		t.Fatalf("expected session not found, got %v", err)
	}
	if err := client.Logout(context.Background(), "sid"); err != nil {
		t.Fatalf("logout: %v", err)
	}

	email, phone, err := client.GetCredentials(context.Background(), 11)
	if err != nil || email != "user@example.com" || phone != "9001234567" {
		t.Fatalf("get credentials: %q %q %v", email, phone, err)
	}

	if _, _, err := client.UpdateCredentials(context.Background(), 11, "new@example.com", "9001234567"); err != errs.ErrInvalidAdvertiserArg {
		t.Fatalf("expected invalid advertiser arg, got %v", err)
	}
}

func TestMapErr(t *testing.T) {
	cases := []struct {
		err  error
		want error
	}{
		{status.Error(codes.AlreadyExists, "phone taken"), errs.ErrPhoneTaken},
		{status.Error(codes.AlreadyExists, "email taken"), errs.ErrEmailTaken},
		{status.Error(codes.FailedPrecondition, "nyi"), errs.NotImplementedError},
		{status.Error(codes.NotFound, "credentials not found"), errs.NotFoundError},
		{status.Error(codes.NotFound, "session not found"), errs.ErrSessionNotFound},
		{status.Error(codes.InvalidArgument, "bad"), errs.ErrInvalidAdvertiserArg},
	}
	for _, tc := range cases {
		if got := mapErr(tc.err); got != tc.want {
			t.Fatalf("expected %v, got %v", tc.want, got)
		}
	}
}

func newAuthBufConn(t *testing.T, srv authv1.AuthServiceServer) *grpc.ClientConn {
	t.Helper()
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	authv1.RegisterAuthServiceServer(server, srv)
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
