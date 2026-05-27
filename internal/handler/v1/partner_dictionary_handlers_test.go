package v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"eshkere/internal/service"
)

func TestPartnerDictionaryHandlers(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{
		listPartnerCountriesFn: func(_ context.Context) []service.DictionaryItem {
			return []service.DictionaryItem{{Code: "RU", Name: "Russia"}}
		},
		listPartnerRegionsFn: func(_ context.Context, countryCode string) []service.DictionaryItem {
			if countryCode != "RU" {
				t.Fatalf("unexpected country code: %s", countryCode)
			}
			return []service.DictionaryItem{{Code: "MSK", Name: "Moscow"}}
		},
		listPartnerFormsFn: func(_ context.Context) []service.DictionaryItem {
			return []service.DictionaryItem{{Code: "self", Name: "Self-employed"}}
		},
		listPartnerCurrenciesFn: func(_ context.Context) []service.DictionaryItem {
			return []service.DictionaryItem{{Code: "RUB", Name: "Ruble"}}
		},
		listPartnerBlockTypesFn: func(_ context.Context) []service.BlockTypeDictionaryItem {
			return []service.BlockTypeDictionaryItem{{Code: "banner", Name: "Banner", Description: "Banner ad", Platforms: []string{"web", "mobile"}}}
		},
		getPartnerGeoTreeFn: func(_ context.Context) []*service.GeoTreeNode {
			return []*service.GeoTreeNode{{Code: "RU", Name: "Russia", Children: []*service.GeoTreeNode{{Code: "MSK", Name: "Moscow"}}}}
		},
		getPartnerIncomeStatsFn: func(_ context.Context, partnerID int, from, to time.Time) (*service.PartnerIncomeStats, error) {
			if partnerID != 7 {
				t.Fatalf("unexpected partner id: %d", partnerID)
			}
			return &service.PartnerIncomeStats{
				From:        from,
				To:          to,
				Impressions: 100,
				Reward:      50,
				ECPM:        500,
				Rows: []service.PartnerIncomeRow{
					{Date: from, SiteID: 1, SiteName: "Main", Domain: "site.test", BlockID: 2, BlockName: "Sidebar", Impressions: 100, Reward: 50},
				},
			}, nil
		},
	}
	r := newTestRouter(ac, svc)
	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, ac, 7)

	tests := []struct {
		path string
		want string
	}{
		{path: "/partners/dictionaries/countries", want: `"code":"RU"`},
		{path: "/partners/dictionaries/registration-regions?country_code=RU", want: `"country_code":"RU"`},
		{path: "/partners/dictionaries/cooperation-forms", want: `"Self-employed"`},
		{path: "/partners/dictionaries/payout-currencies", want: `"Ruble"`},
		{path: "/partners/dictionaries/block-types", want: `"platforms":["web","mobile"]`},
		{path: "/partners/dictionaries/geo-tree", want: `"children":[{"code":"MSK"`},
		{path: "/partners/income/stats?from=2026-05-01&to=2026-05-02", want: `"ecpm":500`},
	}

	for _, tc := range tests {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		req.AddCookie(sess)
		req.AddCookie(csrf)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("%s: expected 200 got %d body=%s", tc.path, rr.Code, rr.Body.String())
		}
		if !strings.Contains(rr.Body.String(), tc.want) {
			t.Fatalf("%s: body %s does not contain %s", tc.path, rr.Body.String(), tc.want)
		}
	}
}

func TestPartnerIncomeStats_InvalidPeriod(t *testing.T) {
	ac := newStubAuthClient()
	r := newTestRouter(ac, &stubService{})
	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, ac, 7)

	req := httptest.NewRequest(http.MethodGet, "/partners/income/stats?from=bad-date", nil)
	req.AddCookie(sess)
	req.AddCookie(csrf)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d body=%s", rr.Code, rr.Body.String())
	}
}
