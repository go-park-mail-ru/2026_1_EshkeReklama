package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"
)

var ErrSessionNotFound = errors.New("session not found")

type Session struct {
	AdvertiserID int64     `json:"advertiser_id"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type Manager struct {
	store Store
	ttl   time.Duration
	now   func() time.Time
}

func NewManager(store Store, ttl time.Duration) *Manager {
	return &Manager{store: store, ttl: ttl, now: time.Now}
}

func (m *Manager) Create(ctx context.Context, advertiserID int64) (sessionID string, expiresAt time.Time, err error) {
	sessionID, err = generateSessionID()
	if err != nil {
		return "", time.Time{}, err
	}

	expiresAt = m.now().Add(m.ttl)
	sess := Session{
		AdvertiserID: advertiserID,
		ExpiresAt:    expiresAt,
	}

	if err := m.store.Save(ctx, sessionID, sess, m.ttl); err != nil {
		return "", time.Time{}, err
	}

	return sessionID, expiresAt, nil
}

func (m *Manager) Get(ctx context.Context, sessionID string) (Session, error) {
	sess, err := m.store.Get(ctx, sessionID)
	if err != nil {
		if errors.Is(err, ErrStoreSessionNotFound) {
			return Session{}, ErrSessionNotFound
		}
		return Session{}, err
	}

	expiresAt := m.now().Add(m.ttl)
	sess.ExpiresAt = expiresAt

	if err := m.store.Save(ctx, sessionID, sess, m.ttl); err != nil {
		return Session{}, err
	}

	return sess, nil
}

func (m *Manager) Delete(ctx context.Context, sessionID string) error {
	return m.store.Delete(ctx, sessionID)
}

func generateSessionID() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
