package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"eshkere/internal/models"
	"eshkere/internal/service"
)

func TestAdRequest_AcceptsEmbedTokenWithoutCSRF(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	svc.requestAdFn = func(_ context.Context, embedToken, visitorID string) (*service.AdRequestResult, error) {
		if embedToken != "pb_live123" {
			t.Fatalf("unexpected embed token: %s", embedToken)
		}
		if visitorID != "v1" {
			t.Fatalf("unexpected visitor id: %s", visitorID)
		}
		return &service.AdRequestResult{
			RequestID: "11111111-1111-4111-8111-111111111111",
			Ad: &models.Ad{
				ID:        77,
				Title:     "Title",
				ShortDesc: "Desc",
				ImageURL:  "https://cdn.example/ad.png",
				TargetURL: "https://target.example",
			},
		}, nil
	}

	req := httptest.NewRequest(http.MethodPost, APIPrefix+"/ad/request", bytes.NewBufferString(`{"embed_token":"pb_live123","visitor_id":"v1"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}

	var envelope struct {
		Data struct {
			RequestID string `json:"request_id"`
			Ad        struct {
				ID int `json:"id"`
			} `json:"ad"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if envelope.Data.RequestID == "" || envelope.Data.Ad.ID != 77 {
		t.Fatalf("unexpected ad request response: %+v", envelope.Data)
	}
}

func TestPartnerBlockEmbed_ResponseContainsIframe(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	ac.addSession("partner-embed-sess", 44)
	sess := &http.Cookie{Name: testCookieName, Value: "partner-embed-sess"}
	csrf := getCSRF(t, r)

	svc.getPartnerBlockEmbedCodeFn = func(_ context.Context, partnerID, siteID, blockID int, baseURL string) (string, string, error) {
		if partnerID != 44 || siteID != 101 || blockID != 9001 {
			t.Fatalf("unexpected args: partnerID=%d siteID=%d blockID=%d", partnerID, siteID, blockID)
		}
		return "pb_embed123",
			`<iframe src="https://ads.example/public/partner/blocks/pb_embed123/frame" width="300" height="250" style="border:0;overflow:hidden" loading="lazy" referrerpolicy="strict-origin-when-cross-origin"></iframe>`,
			nil
	}

	req := httptest.NewRequest(http.MethodGet, APIPrefix+"/partners/sites/101/blocks/9001/embed", nil)
	req.AddCookie(sess)
	req.AddCookie(csrf)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}

	var envelope struct {
		Data struct {
			EmbedToken  string `json:"embed_token"`
			HTMLSnippet string `json:"html_snippet"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if envelope.Data.EmbedToken != "pb_embed123" {
		t.Fatalf("unexpected embed token: %s", envelope.Data.EmbedToken)
	}
	if !strings.Contains(envelope.Data.HTMLSnippet, "<iframe") ||
		!strings.Contains(envelope.Data.HTMLSnippet, `src="https://ads.example/public/partner/blocks/pb_embed123/frame"`) ||
		!strings.Contains(envelope.Data.HTMLSnippet, `referrerpolicy="strict-origin-when-cross-origin"`) {
		t.Fatalf("expected iframe snippet, got: %s", envelope.Data.HTMLSnippet)
	}
}

func TestPartnerBlockFrame_ReturnsRenderableHTML(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	req := httptest.NewRequest(http.MethodGet, APIPrefix+"/public/partner/blocks/pb_frame123/frame", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}
	if contentType := rr.Result().Header.Get("Content-Type"); !strings.Contains(contentType, "text/html") {
		t.Fatalf("expected text/html content type, got: %s", contentType)
	}
	body := rr.Body.String()
	for _, want := range []string{
		`const embedToken = "pb_frame123";`,
		`localStorage.getItem`,
		`visitor_id: visitorId()`,
		`fetch("/api/ad/request"`,
		`renderFallback`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected frame HTML to contain %q, got: %s", want, body)
		}
	}
}

func TestClickAd_RedirectsToTarget(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	svc.clickAdFn = func(_ context.Context, requestID string) (string, error) {
		if requestID != "req123" {
			t.Fatalf("unexpected request id: %s", requestID)
		}
		return "https://target.example/path", nil
	}

	req := httptest.NewRequest(http.MethodGet, APIPrefix+"/click/req123", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("expected 302 got %d body=%s", rr.Code, rr.Body.String())
	}
	if location := rr.Result().Header.Get("Location"); location != "https://target.example/path" {
		t.Fatalf("unexpected redirect: %s", location)
	}
}
