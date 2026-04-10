package session

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"time"
)

var ErrSessionNotFound = errors.New("session not found")

type Session struct {
	AdvertiserID int       `json:"advertiser_id"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type CookieConfig struct {
	Name     string
	Path     string
	HTTPOnly bool
	Secure   bool
	SameSite http.SameSite
}

type Manager struct {
	store  Store
	ttl    time.Duration
	now    func() time.Time
	cookie CookieConfig
}

func NewManager(store Store, ttl time.Duration, cookie CookieConfig) *Manager {
	return &Manager{
		store:  store,
		ttl:    ttl,
		now:    time.Now,
		cookie: cookie,
	}
}

func (m *Manager) Create(w http.ResponseWriter, r *http.Request, advertiserID int) error {
	if cookie, err := r.Cookie(m.cookie.Name); err == nil {
		if err := m.store.Delete(r.Context(), cookie.Value); err != nil {
			return err
		}
	}

	sessionID, err := generateSessionID()
	if err != nil {
		return err
	}

	expiresAt := m.now().Add(m.ttl)
	sess := Session{
		AdvertiserID: advertiserID,
		ExpiresAt:    expiresAt,
	}

	if err := m.store.Save(r.Context(), sessionID, sess, m.ttl); err != nil {
		return err
	}

	m.setCookie(w, sessionID, expiresAt)

	return nil
}

func (m *Manager) Get(w http.ResponseWriter, r *http.Request) (Session, error) {
	cookie, err := r.Cookie(m.cookie.Name)
	if err != nil {
		return Session{}, ErrSessionNotFound
	}

	sess, err := m.store.Get(r.Context(), cookie.Value)
	if err != nil {
		if errors.Is(err, ErrStoreSessionNotFound) {
			return Session{}, ErrSessionNotFound
		}
		return Session{}, err
	}

	expiresAt := m.now().Add(m.ttl)
	sess.ExpiresAt = expiresAt

	if err := m.store.Save(r.Context(), cookie.Value, sess, m.ttl); err != nil {
		return Session{}, err
	}

	m.setCookie(w, cookie.Value, expiresAt)

	return sess, nil
}

func (m *Manager) Destroy(w http.ResponseWriter, r *http.Request) error {
	if cookie, err := r.Cookie(m.cookie.Name); err == nil {
		if err := m.store.Delete(r.Context(), cookie.Value); err != nil {
			return err
		}
	}

	m.expireCookie(w)

	return nil
}

func (m *Manager) setCookie(w http.ResponseWriter, sessionID string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     m.cookie.Name,
		Value:    sessionID,
		Path:     m.cookie.Path,
		Expires:  expiresAt,
		MaxAge:   int(m.ttl.Seconds()),
		HttpOnly: m.cookie.HTTPOnly,
		Secure:   m.cookie.Secure,
		SameSite: m.cookie.SameSite,
	})
}

func (m *Manager) expireCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     m.cookie.Name,
		Value:    "",
		Path:     m.cookie.Path,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: m.cookie.HTTPOnly,
		Secure:   m.cookie.Secure,
		SameSite: m.cookie.SameSite,
	})
}

func generateSessionID() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
