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

func TestParseStatsPeriod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/stats?from=2026-05-01&to=2026-05-03", nil)
	from, to, err := parseStatsPeriod(req)
	if err != nil {
		t.Fatalf("parseStatsPeriod: %v", err)
	}
	if from.Format("2006-01-02") != "2026-05-01" || to.Format("2006-01-02") != "2026-05-03" {
		t.Fatalf("unexpected period: %s..%s", from, to)
	}

	badReq := httptest.NewRequest(http.MethodGet, "/stats?from=2026-05-03&to=2026-05-01", nil)
	if _, _, err := parseStatsPeriod(badReq); err == nil {
		t.Fatalf("expected invalid reversed period")
	}
}

func TestAdvertiserStatsHandlers_Success(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{
		getCampaignStatsFn: func(_ context.Context, advertiserID, campaignID int, from, to time.Time) (*service.CampaignStats, error) {
			if advertiserID != 1 || campaignID != 9 {
				t.Fatalf("unexpected ids: advertiser=%d campaign=%d", advertiserID, campaignID)
			}
			if from.Format("2006-01-02") != "2026-05-01" || to.Format("2006-01-02") != "2026-05-03" {
				t.Fatalf("unexpected range: %s..%s", from, to)
			}
			return &service.CampaignStats{
				Period:         service.StatsPeriod{From: from, To: to},
				Totals:         service.StatsMetric{Impressions: 100, CTR: 2.5},
				PreviousTotals: service.StatsMetric{Impressions: 80},
				Timeline:       []service.StatsPoint{{Date: from, StatsMetric: service.StatsMetric{Impressions: 40}}},
				Groups:         []service.StatsEntityRow{{ID: 3, Name: "Warm"}},
				Placements:     []service.StatsEntityRow{{ID: 7, Name: "Site #7"}},
			}, nil
		},
		getGroupStatsFn: func(_ context.Context, advertiserID, campaignID, groupID int, from, to time.Time) (*service.GroupStats, error) {
			if advertiserID != 1 || campaignID != 9 || groupID != 4 {
				t.Fatalf("unexpected ids for group stats: %d %d %d", advertiserID, campaignID, groupID)
			}
			return &service.GroupStats{
				Period:     service.StatsPeriod{From: from, To: to},
				Totals:     service.StatsMetric{Clicks: 8},
				Timeline:   []service.StatsPoint{{Date: from, StatsMetric: service.StatsMetric{Clicks: 3}}},
				Ads:        []service.StatsEntityRow{{ID: 8, Name: "Ad #8"}},
				Placements: []service.StatsEntityRow{{ID: 1, Name: "Placement #1"}},
			}, nil
		},
		getAdStatsFn: func(_ context.Context, advertiserID, campaignID, groupID, adID int, from, to time.Time) (*service.AdStats, error) {
			if advertiserID != 1 || campaignID != 9 || groupID != 4 || adID != 2 {
				t.Fatalf("unexpected ids for ad stats: %d %d %d %d", advertiserID, campaignID, groupID, adID)
			}
			return &service.AdStats{
				Period:     service.StatsPeriod{From: from, To: to},
				Totals:     service.StatsMetric{Spend: 77},
				Timeline:   []service.StatsPoint{{Date: from, StatsMetric: service.StatsMetric{Spend: 10}}},
				Placements: []service.StatsEntityRow{{ID: 12, Name: "Placement #12"}},
			}, nil
		},
	}
	r := newTestRouter(ac, svc)

	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, ac, 1)

	for _, tc := range []struct {
		path       string
		wantStatus int
		wantBody   string
	}{
		{path: "/ad_campaigns/9/stats?from=2026-05-01&to=2026-05-03", wantStatus: http.StatusOK, wantBody: `"groups":[{"id":3,"name":"Warm"`},
		{path: "/ad_campaigns/9/ad_groups/4/stats?from=2026-05-01&to=2026-05-03", wantStatus: http.StatusOK, wantBody: `"ads":[{"id":8,"name":"Ad #8"`},
		{path: "/ad_campaigns/9/ad_groups/4/ads/2/stats?from=2026-05-01&to=2026-05-03", wantStatus: http.StatusOK, wantBody: `"spend":77`},
	} {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		req.AddCookie(sess)
		req.AddCookie(csrf)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if rr.Code != tc.wantStatus {
			t.Fatalf("%s: expected %d got %d body=%s", tc.path, tc.wantStatus, rr.Code, rr.Body.String())
		}
		if !strings.Contains(rr.Body.String(), tc.wantBody) {
			t.Fatalf("%s: body %s does not contain %s", tc.path, rr.Body.String(), tc.wantBody)
		}
	}
}

func TestAdvertiserStatsHandlers_InvalidInput(t *testing.T) {
	ac := newStubAuthClient()
	r := newTestRouter(ac, &stubService{})
	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, ac, 1)

	for _, tc := range []struct {
		path string
		want int
	}{
		{path: "/ad_campaigns/not-int/stats", want: http.StatusInternalServerError},
		{path: "/ad_campaigns/9/ad_groups/nope/stats", want: http.StatusInternalServerError},
		{path: "/ad_campaigns/9/ad_groups/4/ads/oops/stats", want: http.StatusInternalServerError},
		{path: "/ad_campaigns/9/stats?from=bad-date", want: http.StatusBadRequest},
	} {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		req.AddCookie(sess)
		req.AddCookie(csrf)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if rr.Code != tc.want {
			t.Fatalf("%s: expected %d got %d body=%s", tc.path, tc.want, rr.Code, rr.Body.String())
		}
	}
}
