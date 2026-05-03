package v1

import (
	"bytes"
	"context"
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPartner_CreateSite_And_Block(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{}
	r := newTestRouter(ac, svc)

	csrf := getCSRF(t, r)
	ac.addSession("partner-site-sess", 44)
	sess := &http.Cookie{Name: testCookieName, Value: "partner-site-sess"}

	svc.createPartnerSiteFn = func(_ context.Context, in *serviceinput.CreatePartnerSite) (*models.PartnerSite, error) {
		if in.PartnerID != 44 || in.Domain != "example.com" {
			t.Fatalf("unexpected create site input: %+v", in)
		}
		return &models.PartnerSite{ID: 101, PartnerID: 44, Domain: "example.com", SiteName: "Example"}, nil
	}
	svc.createPartnerBlockFn = func(_ context.Context, partnerID int, in *serviceinput.CreatePartnerBlock) (*models.PartnerBlock, error) {
		if partnerID != 44 || in.PartnerSiteID != 101 || in.BlockType != models.PartnerBlockTypeBanner {
			t.Fatalf("unexpected create block input: partnerID=%d in=%+v", partnerID, in)
		}
		return &models.PartnerBlock{ID: 9001, PartnerSiteID: 101, Name: in.Name, BlockType: in.BlockType, Status: models.PartnerBlockStatusDraft}, nil
	}

	siteReq := httptest.NewRequest(http.MethodPost, "/partners/sites", bytes.NewBufferString(`{"domain":"example.com","site_name":"Example"}`))
	siteReq.AddCookie(sess)
	siteReq.AddCookie(csrf)
	siteReq.Header.Set("X-CSRF-Token", csrf.Value)
	siteRR := httptest.NewRecorder()
	r.ServeHTTP(siteRR, siteReq)
	if siteRR.Code != http.StatusCreated {
		t.Fatalf("expected 201 got %d body=%s", siteRR.Code, siteRR.Body.String())
	}

	blockReq := httptest.NewRequest(http.MethodPost, "/partners/sites/101/blocks", bytes.NewBufferString(`{"block_type":"banner","name":"Баннер (03.05.2026)"}`))
	blockReq.AddCookie(sess)
	blockReq.AddCookie(csrf)
	blockReq.Header.Set("X-CSRF-Token", csrf.Value)
	blockRR := httptest.NewRecorder()
	r.ServeHTTP(blockRR, blockReq)
	if blockRR.Code != http.StatusCreated {
		t.Fatalf("expected 201 got %d body=%s", blockRR.Code, blockRR.Body.String())
	}
}
