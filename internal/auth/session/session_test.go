package session

import (
	"context"
	"errors"
	"testing"
	"time"
)

type stubStore struct {
	saveFunc   func(context.Context, string, Session, time.Duration) error
	getFunc    func(context.Context, string) (Session, error)
	deleteFunc func(context.Context, string) error
}

func (s *stubStore) Save(ctx context.Context, sessionID string, sess Session, ttl time.Duration) error {
	if s.saveFunc != nil {
		return s.saveFunc(ctx, sessionID, sess, ttl)
	}
	return nil
}

func (s *stubStore) Get(ctx context.Context, sessionID string) (Session, error) {
	if s.getFunc != nil {
		return s.getFunc(ctx, sessionID)
	}
	return Session{}, nil
}

func (s *stubStore) Delete(ctx context.Context, sessionID string) error {
	if s.deleteFunc != nil {
		return s.deleteFunc(ctx, sessionID)
	}
	return nil
}

func TestManagerCreateGetDelete(t *testing.T) {
	now := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)
	var savedID string
	var savedSession Session
	var savedTTL time.Duration

	store := &stubStore{
		saveFunc: func(_ context.Context, sessionID string, sess Session, ttl time.Duration) error {
			savedID = sessionID
			savedSession = sess
			savedTTL = ttl
			return nil
		},
		getFunc: func(_ context.Context, sessionID string) (Session, error) {
			if sessionID != savedID {
				t.Fatalf("unexpected session id: %s", sessionID)
			}
			return Session{AdvertiserID: 42}, nil
		},
		deleteFunc: func(_ context.Context, sessionID string) error {
			if sessionID != savedID {
				t.Fatalf("unexpected delete session id: %s", sessionID)
			}
			return nil
		},
	}

	manager := NewManager(store, 2*time.Hour)
	manager.now = func() time.Time { return now }

	sessionID, expiresAt, err := manager.Create(context.Background(), 42)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if sessionID == "" {
		t.Fatal("expected non-empty session id")
	}
	if len(sessionID) != 64 {
		t.Fatalf("unexpected session id length: %d", len(sessionID))
	}
	if !expiresAt.Equal(now.Add(2 * time.Hour)) {
		t.Fatalf("unexpected expiry: %v", expiresAt)
	}
	if savedSession.AdvertiserID != 42 || !savedSession.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("unexpected saved session: %#v", savedSession)
	}
	if savedTTL != 2*time.Hour {
		t.Fatalf("unexpected ttl: %v", savedTTL)
	}

	got, err := manager.Get(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.AdvertiserID != 42 {
		t.Fatalf("unexpected advertiser id: %d", got.AdvertiserID)
	}
	if !got.ExpiresAt.Equal(now.Add(2 * time.Hour)) {
		t.Fatalf("unexpected refreshed expiry: %v", got.ExpiresAt)
	}

	if err := manager.Delete(context.Background(), sessionID); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestManagerGetMapsStoreNotFound(t *testing.T) {
	manager := NewManager(&stubStore{
		getFunc: func(context.Context, string) (Session, error) {
			return Session{}, ErrStoreSessionNotFound
		},
	}, time.Hour)

	_, err := manager.Get(context.Background(), "sid")
	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestManagerGetSaveErrorReturned(t *testing.T) {
	saveErr := errors.New("save failed")
	manager := NewManager(&stubStore{
		getFunc: func(context.Context, string) (Session, error) {
			return Session{AdvertiserID: 1}, nil
		},
		saveFunc: func(context.Context, string, Session, time.Duration) error {
			return saveErr
		},
	}, time.Hour)

	_, err := manager.Get(context.Background(), "sid")
	if !errors.Is(err, saveErr) {
		t.Fatalf("expected save error, got %v", err)
	}
}
