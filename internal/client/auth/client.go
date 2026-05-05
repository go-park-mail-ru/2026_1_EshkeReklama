package authclient

import (
	"context"
	"time"

	errs "eshkere/internal/errors"
	authv1 "eshkere/pkg/pb/auth/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// Compile-time check that Client satisfies the handler's AuthClient interface.
// (avoids import cycle — the interface is redeclared in the handler package)
var _ interface {
	Register(ctx context.Context, email, phone, password string) (int64, string, int64, error)
	Login(ctx context.Context, identifier, password string) (int64, string, int64, error)
	LoginVKID(ctx context.Context, code, deviceID, codeVerifier string) (int64, string, int64, error)
	ValidateSession(ctx context.Context, sessionID string) (int64, error)
	Logout(ctx context.Context, sessionID string) error
	GetCredentials(ctx context.Context, advertiserID int64) (string, string, error)
	UpdateCredentials(ctx context.Context, advertiserID int64, email, phone string) (string, string, error)
} = (*Client)(nil)

type Client struct {
	conn *grpc.ClientConn
	rpc  authv1.AuthServiceClient
}

func New(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn, rpc: authv1.NewAuthServiceClient(conn)}, nil
}

func (c *Client) Close() error { return c.conn.Close() }

func (c *Client) Register(ctx context.Context, email, phone, password string) (int64, string, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	resp, err := c.rpc.Register(ctx, &authv1.RegisterRequest{
		Email: email, Phone: phone, Password: password,
	})
	if err != nil {
		return 0, "", 0, mapErr(err)
	}
	return resp.AdvertiserId, resp.SessionId, resp.ExpiresAt, nil
}

func (c *Client) Login(ctx context.Context, identifier, password string) (int64, string, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	resp, err := c.rpc.Login(ctx, &authv1.LoginRequest{
		Identifier: identifier, Password: password,
	})
	if err != nil {
		return 0, "", 0, mapErr(err)
	}
	return resp.AdvertiserId, resp.SessionId, resp.ExpiresAt, nil
}

func (c *Client) LoginVKID(ctx context.Context, code, deviceID, codeVerifier string) (int64, string, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.rpc.LoginVKID(ctx, &authv1.LoginVKIDRequest{
		Code:         code,
		DeviceId:     deviceID,
		CodeVerifier: codeVerifier,
	})
	if err != nil {
		return 0, "", 0, mapErr(err)
	}
	return resp.AdvertiserId, resp.SessionId, resp.ExpiresAt, nil
}

func (c *Client) ValidateSession(ctx context.Context, sessionID string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	resp, err := c.rpc.ValidateSession(ctx, &authv1.ValidateSessionRequest{
		SessionId: sessionID,
	})
	if err != nil {
		return 0, mapErr(err)
	}
	return resp.AdvertiserId, nil
}

func (c *Client) Logout(ctx context.Context, sessionID string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := c.rpc.Logout(ctx, &authv1.LogoutRequest{SessionId: sessionID})
	if err != nil {
		return mapErr(err)
	}
	return nil
}

func (c *Client) GetCredentials(ctx context.Context, advertiserID int64) (string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	resp, err := c.rpc.GetCredentials(ctx, &authv1.GetCredentialsRequest{AdvertiserId: advertiserID})
	if err != nil {
		return "", "", mapErr(err)
	}

	return resp.Email, resp.Phone, nil
}

func (c *Client) UpdateCredentials(ctx context.Context, advertiserID int64, email, phone string) (string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	resp, err := c.rpc.UpdateCredentials(ctx, &authv1.UpdateCredentialsRequest{
		AdvertiserId: advertiserID,
		Email:        email,
		Phone:        phone,
	})
	if err != nil {
		return "", "", mapErr(err)
	}

	return resp.Email, resp.Phone, nil
}

func mapErr(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return err
	}
	switch st.Code() {
	case codes.AlreadyExists:
		if st.Message() == "vk credentials conflict" {
			return errs.ErrVKIDConflict
		}
		if st.Message() == "phone taken" {
			return errs.ErrPhoneTaken
		}
		return errs.ErrEmailTaken
	case codes.Unauthenticated:
		return errs.ErrInvalidCredentials
	case codes.FailedPrecondition:
		return errs.NotImplementedError
	case codes.NotFound:
		if st.Message() == "credentials not found" {
			return errs.NotFoundError
		}
		return errs.ErrSessionNotFound
	case codes.InvalidArgument:
		return errs.ErrInvalidAdvertiserArg
	default:
		return err
	}
}
