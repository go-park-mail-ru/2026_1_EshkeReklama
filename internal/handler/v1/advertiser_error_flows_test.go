package v1

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	errs "eshkere/internal/errors"
	"eshkere/internal/models"
)

func TestRegister_ProfileCreationFailureRollsBackSession(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	csrf := getCSRF(t, r)

	logoutCalled := false
	ac.registerFn = func(_ context.Context, email, phone, password string) (int64, string, int64, error) {
		return 22, "sess-register", 9999999999, nil
	}
	ac.logoutFn = func(_ context.Context, sessionID string) error {
		logoutCalled = true
		if sessionID != "sess-register" {
			t.Fatalf("unexpected session id: %s", sessionID)
		}
		return nil
	}
	svc.createAdvertiserProfileFn = func(_ context.Context, id int64, name, email string) error {
		if id != 22 || name != "Tester" || email != "tester@mail.test" {
			t.Fatalf("unexpected profile args: id=%d name=%q email=%q", id, name, email)
		}
		return errors.New("profile insert failed")
	}

	req := httptest.NewRequest(http.MethodPost, "/advertisers/register", bytes.NewBufferString(`{"name":"Tester","email":"tester@mail.test","phone":"+70000000000","password":"secret"}`))
	req.AddCookie(csrf)
	req.Header.Set("X-CSRF-Token", csrf.Value)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 got %d body=%s", rr.Code, rr.Body.String())
	}
	if !logoutCalled {
		t.Fatal("expected auth logout rollback to be called")
	}
}

func TestLogin_GetCredentialsFailure(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	csrf := getCSRF(t, r)

	ac.loginFn = func(_ context.Context, identifier, password string) (int64, string, int64, error) {
		if identifier != "test@mail.com" || password != "secret" {
			t.Fatalf("unexpected login args: identifier=%q password=%q", identifier, password)
		}
		return 5, "sess-login", 9999999999, nil
	}
	ac.getCredentialsFn = func(_ context.Context, advertiserID int64) (string, string, bool, error) {
		if advertiserID != 5 {
			t.Fatalf("unexpected advertiser id: %d", advertiserID)
		}
		return "", "", false, errors.New("credentials unavailable")
	}

	req := httptest.NewRequest(http.MethodPost, "/advertisers/login", bytes.NewBufferString(`{"identifier":"test@mail.com","password":"secret"}`))
	req.AddCookie(csrf)
	req.Header.Set("X-CSRF-Token", csrf.Value)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestLogout_ErrorFromAuthClient(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, ac, 1)

	ac.logoutFn = func(_ context.Context, sessionID string) error {
		if sessionID != sess.Value {
			t.Fatalf("unexpected session id: %s", sessionID)
		}
		return errors.New("logout failed")
	}

	req := httptest.NewRequest(http.MethodPost, "/advertisers/logout", nil)
	req.AddCookie(sess)
	req.AddCookie(csrf)
	req.Header.Set("X-CSRF-Token", csrf.Value)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestLoginVKID_EnsureProfileFailureLogsOut(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	csrf := getCSRF(t, r)

	logoutCalled := false
	ac.loginVKIDFn = func(_ context.Context, accessToken string, userID int64) (int64, string, int64, string, string, error) {
		if accessToken != "vk-token" || userID != 77 {
			t.Fatalf("unexpected vk payload: token=%q userID=%d", accessToken, userID)
		}
		return 7, "vk-failed-session", 9999999999, "Vasya", "Petrov", nil
	}
	ac.getCredentialsFn = func(_ context.Context, advertiserID int64) (string, string, bool, error) {
		if advertiserID != 7 {
			t.Fatalf("unexpected advertiser id: %d", advertiserID)
		}
		return "vk@example.com", "+79991112233", false, nil
	}
	ac.logoutFn = func(_ context.Context, sessionID string) error {
		logoutCalled = true
		if sessionID != "vk-failed-session" {
			t.Fatalf("unexpected logout session id: %s", sessionID)
		}
		return nil
	}
	svc.getAdvertiserByIDFn = func(_ context.Context, id int) (*models.Advertiser, error) {
		if id != 7 {
			t.Fatalf("unexpected advertiser id: %d", id)
		}
		return nil, errs.NotFoundError
	}
	svc.createAdvertiserProfileFn = func(_ context.Context, id int64, name, email string) error {
		if id != 7 || name != "Vasya" || email != "vk@example.com" {
			t.Fatalf("unexpected profile args: id=%d name=%q email=%q", id, name, email)
		}
		return errors.New("profile create failed")
	}

	req := httptest.NewRequest(http.MethodPost, "/advertisers/login/vk", bytes.NewBufferString(`{"access_token":"vk-token","user_id":77}`))
	req.AddCookie(csrf)
	req.Header.Set("X-CSRF-Token", csrf.Value)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 got %d body=%s", rr.Code, rr.Body.String())
	}
	if !logoutCalled {
		t.Fatal("expected logout rollback after ensure profile failure")
	}
}
