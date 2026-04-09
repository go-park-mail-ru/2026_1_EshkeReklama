package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"eshkere/internal/handler/dto"
	"eshkere/internal/session"

	"github.com/gorilla/mux"
)

func newTestRouter(sm *session.Manager) *mux.Router {
	r := mux.NewRouter().StrictSlash(true)
	Register(r, NewAPI(APIConfig{
		SessionManager: sm,
	}))
	return r
}

func TestRegister_NotImplemented(t *testing.T) {
	sm := session.NewManager()
	r := newTestRouter(sm)

	body := `{"email":"a@a.test","phone":"+70000000000","password":"p"}`
	req := httptest.NewRequest(http.MethodPost, AdvertiserGroupURI+RegisterURI, bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501 got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestLogin_NotImplemented(t *testing.T) {
	sm := session.NewManager()
	r := newTestRouter(sm)

	req := httptest.NewRequest(http.MethodPost, AdvertiserGroupURI+LoginURI, bytes.NewBufferString(`{"identifier":"test@mail.com","password":"bad"}`))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501 got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestLogout_AlwaysOK(t *testing.T) {
	sm := session.NewManager()
	r := newTestRouter(sm)

	req := httptest.NewRequest(http.MethodPost, AdvertiserGroupURI+LogoutURI, nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestListAds_UnauthorizedAndEmptyList(t *testing.T) {
	sm := session.NewManager()
	r := newTestRouter(sm)

	req := httptest.NewRequest(http.MethodGet, AdsGroupURI, nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 got %d body=%s", rr.Code, rr.Body.String())
	}

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
		Status string              `json:"status"`
		Data   dto.ListAdsResponse `json:"data"`
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
	if len(envelope.Data.Ads) != 0 {
		t.Fatalf("expected empty ads list, got %d", len(envelope.Data.Ads))
	}
}
