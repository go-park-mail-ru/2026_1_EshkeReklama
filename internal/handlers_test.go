package internal

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"eshkere/internal/session"

	"github.com/gorilla/mux"
)

func resetUsersForTests() {
	userMu.Lock()
	defer userMu.Unlock()

	usersByEmail = make(map[string]*User)
	usersByPhone = make(map[string]*User)
	lastID = 0

	// same as init()
	lastID++
	u := &User{
		ID:       lastID,
		Email:    "test@mail.com",
		Phone:    "+79991234567",
		Password: "123123",
	}
	usersByEmail[u.Email] = u
	usersByPhone[u.Phone] = u
}

func newTestRouter(sm *session.Manager) *mux.Router {
	r := mux.NewRouter().StrictSlash(true)
	Register(r, NewAPI(APIConfig{SessionManager: sm}))
	return r
}

func TestRegister_SuccessAndDuplicate(t *testing.T) {
	resetUsersForTests()
	sm := session.NewManager()
	r := newTestRouter(sm)

	body := `{"email":"a@a.test","phone":"+70000000000","password":"p"}`
	req := httptest.NewRequest(http.MethodPost, AdvertiserGroupURI+RegisterURI, bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}
	if len(rr.Result().Cookies()) == 0 {
		t.Fatalf("expected session cookie to be set")
	}

	// duplicate register
	req2 := httptest.NewRequest(http.MethodPost, AdvertiserGroupURI+RegisterURI, bytes.NewBufferString(body))
	rr2 := httptest.NewRecorder()
	r.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d body=%s", rr2.Code, rr2.Body.String())
	}
}

func TestLogin_InvalidAndSuccess(t *testing.T) {
	resetUsersForTests()
	sm := session.NewManager()
	r := newTestRouter(sm)

	// invalid password
	req := httptest.NewRequest(http.MethodPost, AdvertiserGroupURI+LoginURI, bytes.NewBufferString(`{"identifier":"test@mail.com","password":"bad"}`))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d body=%s", rr.Code, rr.Body.String())
	}

	// success by email
	req2 := httptest.NewRequest(http.MethodPost, AdvertiserGroupURI+LoginURI, bytes.NewBufferString(`{"identifier":"test@mail.com","password":"123123"}`))
	rr2 := httptest.NewRecorder()
	r.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr2.Code, rr2.Body.String())
	}
	if len(rr2.Result().Cookies()) == 0 {
		t.Fatalf("expected session cookie to be set")
	}

	var envelope struct {
		Status string        `json:"status"`
		Data   LoginResponse `json:"data"`
	}
	if err := json.Unmarshal(rr2.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if envelope.Status != "ok" || envelope.Data.Email != "test@mail.com" || envelope.Data.ID == 0 {
		t.Fatalf("unexpected response: %#v", envelope)
	}
}

func TestLogout_AlwaysOK(t *testing.T) {
	resetUsersForTests()
	sm := session.NewManager()
	r := newTestRouter(sm)

	req := httptest.NewRequest(http.MethodPost, AdvertiserGroupURI+LogoutURI, nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestListAds_UnauthorizedAndAuthorized(t *testing.T) {
	resetUsersForTests()
	sm := session.NewManager()
	r := newTestRouter(sm)

	// unauthorized
	req := httptest.NewRequest(http.MethodGet, AdsGroupURI, nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 got %d body=%s", rr.Code, rr.Body.String())
	}

	// create a session for advertiserID=1
	createReq := httptest.NewRequest(http.MethodPost, "/", nil)
	createRR := httptest.NewRecorder()
	if err := sm.Create(createRR, createReq, 1); err != nil {
		t.Fatalf("Create: %v", err)
	}
	cookies := createRR.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected cookie")
	}

	req2 := httptest.NewRequest(http.MethodGet, AdsGroupURI, nil)
	req2.AddCookie(cookies[0])
	rr2 := httptest.NewRecorder()
	r.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr2.Code, rr2.Body.String())
	}

	var envelope struct {
		Status string          `json:"status"`
		Data   ListAdsResponse `json:"data"`
	}
	if err := json.Unmarshal(rr2.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if envelope.Status != "ok" {
		t.Fatalf("expected ok status, got %q", envelope.Status)
	}
	if envelope.Data.AdvertiserID != 1 {
		t.Fatalf("expected advertiser_id 1 got %d", envelope.Data.AdvertiserID)
	}
	if len(envelope.Data.Ads) != 2 {
		t.Fatalf("expected 2 ads for advertiser 1 got %d", len(envelope.Data.Ads))
	}
}
