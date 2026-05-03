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

	svc.requestAdFn = func(_ context.Context, embedToken string) (*service.AdRequestResult, error) {
		if embedToken != "pb_live123" {
			t.Fatalf("unexpected embed token: %s", embedToken)
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

	req := httptest.NewRequest(http.MethodPost, "/ad/request", bytes.NewBufferString(`{"embed_token":"pb_live123"}`))
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

func TestPartnerBlockEmbed_ResponseContainsDivAndScript(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	ac.addSession("partner-embed-sess", 44)
	sess := &http.Cookie{Name: testCookieName, Value: "partner-embed-sess"}
	csrf := getCSRF(t, r)

	svc.getPartnerBlockEmbedCodeFn = func(_ context.Context, partnerID, siteID, blockID int, baseURL, adSDKURL string) (string, string, string, string, error) {
		if partnerID != 44 || siteID != 101 || blockID != 9001 {
			t.Fatalf("unexpected args: partnerID=%d siteID=%d blockID=%d", partnerID, siteID, blockID)
		}
		if adSDKURL != "" {
			t.Fatalf("unexpected ad sdk url: %s", adSDKURL)
		}
		return "pb_embed123",
			"https://ads.example/public/ad-sdk.js",
			"https://ads.example/public/partner/blocks/pb_embed123/frame",
			"<div data-eshkere-ad=\"pb_embed123\"></div>\n<script async src=\"https://ads.example/public/ad-sdk.js\"></script>",
			nil
	}

	req := httptest.NewRequest(http.MethodGet, "/partners/sites/101/blocks/9001/embed", nil)
	req.AddCookie(sess)
	req.AddCookie(csrf)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}

	var envelope struct {
		Data struct {
			HTMLSnippet string `json:"html_snippet"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !strings.Contains(envelope.Data.HTMLSnippet, "data-eshkere-ad") || !strings.Contains(envelope.Data.HTMLSnippet, "<script async src=") {
		t.Fatalf("expected div+script snippet, got: %s", envelope.Data.HTMLSnippet)
	}
}
