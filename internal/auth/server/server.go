package server

import (
	"context"
	"database/sql"
	"errors"

	authsvc "eshkere/internal/auth/service"
	authsession "eshkere/internal/auth/session"
	authv1 "eshkere/pkg/pb/auth/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	authv1.UnimplementedAuthServiceServer

	creds   *authsvc.CredentialsService
	session *authsession.Manager
}

func New(creds *authsvc.CredentialsService, session *authsession.Manager) *Server {
	return &Server{creds: creds, session: session}
}

func (s *Server) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	id, err := s.creds.Register(ctx, req.Email, req.Phone, req.Password)
	if err != nil {
		return nil, mapErr(err)
	}

	sessionID, expiresAt, err := s.session.Create(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, "create session")
	}

	return &authv1.RegisterResponse{
		AdvertiserId: id,
		SessionId:    sessionID,
		ExpiresAt:    expiresAt.Unix(),
	}, nil
}

func (s *Server) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	id, err := s.creds.Authenticate(ctx, req.Identifier, req.Password)
	if err != nil {
		return nil, mapErr(err)
	}

	sessionID, expiresAt, err := s.session.Create(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, "create session")
	}

	return &authv1.LoginResponse{
		AdvertiserId: id,
		SessionId:    sessionID,
		ExpiresAt:    expiresAt.Unix(),
	}, nil
}

func (s *Server) ValidateSession(ctx context.Context, req *authv1.ValidateSessionRequest) (*authv1.ValidateSessionResponse, error) {
	sess, err := s.session.Get(ctx, req.SessionId)
	if err != nil {
		if errors.Is(err, authsession.ErrSessionNotFound) {
			return nil, status.Error(codes.NotFound, "session not found")
		}
		return nil, status.Error(codes.Internal, "get session")
	}

	return &authv1.ValidateSessionResponse{
		AdvertiserId: sess.AdvertiserID,
		ExpiresAt:    sess.ExpiresAt.Unix(),
	}, nil
}

func (s *Server) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	if err := s.session.Delete(ctx, req.SessionId); err != nil {
		return nil, status.Error(codes.Internal, "delete session")
	}
	return &authv1.LogoutResponse{}, nil
}

func (s *Server) GetCredentials(ctx context.Context, req *authv1.GetCredentialsRequest) (*authv1.GetCredentialsResponse, error) {
	cred, err := s.creds.GetByID(ctx, req.AdvertiserId)
	if err != nil {
		return nil, mapErr(err)
	}

	return &authv1.GetCredentialsResponse{
		AdvertiserId: cred.ID,
		Email:        cred.Email,
		Phone:        cred.Phone,
	}, nil
}

func (s *Server) UpdateCredentials(ctx context.Context, req *authv1.UpdateCredentialsRequest) (*authv1.UpdateCredentialsResponse, error) {
	cred, err := s.creds.Update(ctx, req.AdvertiserId, req.Email, req.Phone)
	if err != nil {
		return nil, mapErr(err)
	}

	return &authv1.UpdateCredentialsResponse{
		AdvertiserId: cred.ID,
		Email:        cred.Email,
		Phone:        cred.Phone,
	}, nil
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, authsvc.ErrEmailTaken):
		return status.Error(codes.AlreadyExists, "email taken")
	case errors.Is(err, authsvc.ErrPhoneTaken):
		return status.Error(codes.AlreadyExists, "phone taken")
	case errors.Is(err, authsvc.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, "invalid credentials")
	case errors.Is(err, authsvc.ErrInvalidArg):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, sql.ErrNoRows):
		return status.Error(codes.NotFound, "credentials not found")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
