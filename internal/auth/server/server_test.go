package server

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	authrepo "eshkere/internal/auth/repository/postgres"
	authsvc "eshkere/internal/auth/service"
	authsession "eshkere/internal/auth/session"
	authv1 "eshkere/pkg/pb/auth/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverStore struct {
	getFunc    func(context.Context, string) (authsession.Session, error)
	saveFunc   func(context.Context, string, authsession.Session, time.Duration) error
	deleteFunc func(context.Context, string) error
}

func (s *serverStore) Save(ctx context.Context, sessionID string, sess authsession.Session, ttl time.Duration) error {
	return s.saveFunc(ctx, sessionID, sess, ttl)
}
func (s *serverStore) Get(ctx context.Context, sessionID string) (authsession.Session, error) {
	return s.getFunc(ctx, sessionID)
}
func (s *serverStore) Delete(ctx context.Context, sessionID string) error {
	return s.deleteFunc(ctx, sessionID)
}

type serverRepo struct {
	createFunc     func(context.Context, string, string, string) (int64, error)
	getByEmailFunc func(context.Context, string) (*authrepo.Credential, error)
	getByPhoneFunc func(context.Context, string) (*authrepo.Credential, error)
	getByIDFunc    func(context.Context, int64) (*authrepo.Credential, error)
	updateFunc     func(context.Context, int64, string, string) error
	getByVKFunc    func(context.Context, int64) (*authrepo.Credential, error)
	createVKFunc   func(context.Context, string, string, int64) (int64, error)
	linkVKFunc     func(context.Context, int64, int64) error
	deleteFunc     func(context.Context, int64) error
}

func (s *serverRepo) Create(ctx context.Context, email, phone, hash string) (int64, error) {
	return s.createFunc(ctx, email, phone, hash)
}
func (s *serverRepo) CreateVK(ctx context.Context, email, phone string, vkUserID int64) (int64, error) {
	return s.createVKFunc(ctx, email, phone, vkUserID)
}
func (s *serverRepo) GetByEmail(ctx context.Context, email string) (*authrepo.Credential, error) {
	return s.getByEmailFunc(ctx, email)
}
func (s *serverRepo) GetByPhone(ctx context.Context, phone string) (*authrepo.Credential, error) {
	return s.getByPhoneFunc(ctx, phone)
}
func (s *serverRepo) GetByVKUserID(ctx context.Context, vkUserID int64) (*authrepo.Credential, error) {
	return s.getByVKFunc(ctx, vkUserID)
}
func (s *serverRepo) GetByID(ctx context.Context, id int64) (*authrepo.Credential, error) {
	return s.getByIDFunc(ctx, id)
}
func (s *serverRepo) LinkVKUserID(ctx context.Context, id, vkUserID int64) error {
	return s.linkVKFunc(ctx, id, vkUserID)
}
func (s *serverRepo) Update(ctx context.Context, id int64, email, phone string) error {
	return s.updateFunc(ctx, id, email, phone)
}
func (s *serverRepo) Delete(ctx context.Context, id int64) error {
	return s.deleteFunc(ctx, id)
}

func TestServerHappyPathAndMapErr(t *testing.T) {
	hash, err := authsvcHash("secret123")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	created := false
	repo := &serverRepo{
		createFunc: func(context.Context, string, string, string) (int64, error) {
			created = true
			return 7, nil
		},
		createVKFunc: func(context.Context, string, string, int64) (int64, error) { return 0, nil },
		getByEmailFunc: func(_ context.Context, email string) (*authrepo.Credential, error) {
			switch email {
			case "user@example.com":
				if !created {
					return nil, sql.ErrNoRows
				}
				return &authrepo.Credential{ID: 7, Email: email, Phone: "9001234567", PasswordHash: hash}, nil
			case "new@example.com":
				return nil, sql.ErrNoRows
			default:
				return nil, sql.ErrNoRows
			}
		},
		getByPhoneFunc: func(context.Context, string) (*authrepo.Credential, error) { return nil, sql.ErrNoRows },
		getByIDFunc: func(context.Context, int64) (*authrepo.Credential, error) {
			return &authrepo.Credential{ID: 7, Email: "user@example.com", Phone: "9001234567"}, nil
		},
		updateFunc:  func(context.Context, int64, string, string) error { return nil },
		getByVKFunc: func(context.Context, int64) (*authrepo.Credential, error) { return nil, sql.ErrNoRows },
		linkVKFunc:  func(context.Context, int64, int64) error { return nil },
		deleteFunc:  func(context.Context, int64) error { return nil },
	}
	var deletedSession string
	store := &serverStore{
		saveFunc: func(context.Context, string, authsession.Session, time.Duration) error { return nil },
		getFunc: func(_ context.Context, sessionID string) (authsession.Session, error) {
			return authsession.Session{AdvertiserID: 7, ExpiresAt: time.Now().Add(time.Hour)}, nil
		},
		deleteFunc: func(_ context.Context, sessionID string) error {
			deletedSession = sessionID
			return nil
		},
	}

	manager := authsession.NewManager(store, time.Hour)
	svc := authsvc.NewCredentialsService(repo, nil)
	srv := New(svc, manager)

	registerResp, err := srv.Register(context.Background(), &authv1.RegisterRequest{
		Email:    "user@example.com",
		Phone:    "9001234567",
		Password: "secret123",
	})
	if err != nil || registerResp.AdvertiserId != 7 || registerResp.SessionId == "" {
		t.Fatalf("register: resp=%#v err=%v", registerResp, err)
	}

	loginResp, err := srv.Login(context.Background(), &authv1.LoginRequest{
		Identifier: "user@example.com",
		Password:   "secret123",
	})
	if err != nil || loginResp.AdvertiserId != 7 {
		t.Fatalf("login: resp=%#v err=%v", loginResp, err)
	}

	validateResp, err := srv.ValidateSession(context.Background(), &authv1.ValidateSessionRequest{SessionId: "sid"})
	if err != nil || validateResp.AdvertiserId != 7 {
		t.Fatalf("validate: resp=%#v err=%v", validateResp, err)
	}

	credResp, err := srv.GetCredentials(context.Background(), &authv1.GetCredentialsRequest{AdvertiserId: 7})
	if err != nil || credResp.Email != "user@example.com" {
		t.Fatalf("get credentials: resp=%#v err=%v", credResp, err)
	}

	updateResp, err := srv.UpdateCredentials(context.Background(), &authv1.UpdateCredentialsRequest{
		AdvertiserId: 7,
		Email:        "new@example.com",
		Phone:        "9001234567",
	})
	if err != nil || updateResp.Email != "new@example.com" {
		t.Fatalf("update credentials: resp=%#v err=%v", updateResp, err)
	}

	if _, err := srv.Logout(context.Background(), &authv1.LogoutRequest{SessionId: "sid"}); err != nil || deletedSession != "sid" {
		t.Fatalf("logout err=%v deleted=%q", err, deletedSession)
	}

	cases := []struct {
		err  error
		code codes.Code
	}{
		{authsvc.ErrEmailTaken, codes.AlreadyExists},
		{authsvc.ErrPhoneTaken, codes.AlreadyExists},
		{authsvc.ErrVKIDConflict, codes.AlreadyExists},
		{authsvc.ErrInvalidCredentials, codes.Unauthenticated},
		{authsvc.ErrInvalidArg, codes.InvalidArgument},
		{authsvc.ErrVKIDUnavailable, codes.FailedPrecondition},
		{sql.ErrNoRows, codes.NotFound},
		{errors.New("boom"), codes.Internal},
	}
	for _, tc := range cases {
		if status.Code(mapErr(tc.err)) != tc.code {
			t.Fatalf("unexpected code for %v: %v", tc.err, status.Code(mapErr(tc.err)))
		}
	}
}

func TestServerErrorPaths(t *testing.T) {
	repo := &serverRepo{
		createFunc: func(context.Context, string, string, string) (int64, error) { return 0, authsvc.ErrEmailTaken },
		createVKFunc: func(context.Context, string, string, int64) (int64, error) {
			return 0, nil
		},
		getByEmailFunc: func(context.Context, string) (*authrepo.Credential, error) { return nil, sql.ErrNoRows },
		getByPhoneFunc: func(context.Context, string) (*authrepo.Credential, error) { return nil, sql.ErrNoRows },
		getByIDFunc:    func(context.Context, int64) (*authrepo.Credential, error) { return nil, sql.ErrNoRows },
		updateFunc:     func(context.Context, int64, string, string) error { return nil },
		getByVKFunc:    func(context.Context, int64) (*authrepo.Credential, error) { return nil, sql.ErrNoRows },
		linkVKFunc:     func(context.Context, int64, int64) error { return nil },
		deleteFunc:     func(context.Context, int64) error { return nil },
	}
	store := &serverStore{
		saveFunc: func(context.Context, string, authsession.Session, time.Duration) error {
			return errors.New("save failed")
		},
		getFunc: func(context.Context, string) (authsession.Session, error) {
			return authsession.Session{}, authsession.ErrStoreSessionNotFound
		},
		deleteFunc: func(context.Context, string) error {
			return errors.New("delete failed")
		},
	}
	manager := authsession.NewManager(store, time.Hour)
	svc := authsvc.NewCredentialsService(repo, nil)
	srv := New(svc, manager)

	if _, err := srv.Register(context.Background(), &authv1.RegisterRequest{
		Email:    "user@example.com",
		Phone:    "9001234567",
		Password: "secret123",
	}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("expected already exists, got %v", status.Code(err))
	}
	if _, err := srv.ValidateSession(context.Background(), &authv1.ValidateSessionRequest{SessionId: "sid"}); status.Code(err) != codes.NotFound {
		t.Fatalf("expected not found, got %v", status.Code(err))
	}
	if _, err := srv.Logout(context.Background(), &authv1.LogoutRequest{SessionId: "sid"}); status.Code(err) != codes.Internal {
		t.Fatalf("expected internal, got %v", status.Code(err))
	}
}

func authsvcHash(password string) (string, error) {
	return authsvcTestHash(password)
}
