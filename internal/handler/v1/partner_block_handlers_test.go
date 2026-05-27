package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"

	"github.com/gorilla/mux"
)

func TestPartnerBlockHelpers(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/partners/sites/10/blocks/20", nil)
	req = muxSetVars(req, map[string]string{"site_id": "10", "block_id": "20"})
	siteID, blockID, err := parseSiteAndBlockIDs(req)
	if err != nil || siteID != 10 || blockID != 20 {
		t.Fatalf("parseSiteAndBlockIDs mismatch site=%d block=%d err=%v", siteID, blockID, err)
	}

	onlyConfigured, globalCPMV := deriveGeoMeta([]*models.PartnerBlockGeoRule{
		{GeoCode: "all", CPMV: sql.NullInt64{Int64: 150, Valid: true}},
		{GeoCode: "RU-MOW", IsEnabled: true},
	})
	if !onlyConfigured || globalCPMV == nil || *globalCPMV != 150 {
		t.Fatalf("unexpected geo meta: onlyConfigured=%v global=%v", onlyConfigured, globalCPMV)
	}
	if got := supportedPlatformsForBlockType(models.PartnerBlockTypeBanner); len(got) != 3 || got[2] != "amp" {
		t.Fatalf("unexpected banner platforms: %+v", got)
	}
	if got := supportedPlatformsForBlockType(models.PartnerBlockTypeTopAd); len(got) != 1 || got[0] != "mobile" {
		t.Fatalf("unexpected top_ad platforms: %+v", got)
	}
}

func TestPartnerBlockHandlers(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{
		listPartnerBlocksFn: func(_ context.Context, partnerID, siteID int) ([]*models.PartnerBlock, error) {
			return []*models.PartnerBlock{{ID: 2, PartnerSiteID: siteID, Name: "Sidebar", BlockType: models.PartnerBlockTypeBanner, Status: models.PartnerBlockStatusActive, CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}}, nil
		},
		getPartnerBlockFn: func(_ context.Context, partnerID, siteID, blockID int) (*models.PartnerBlock, []*models.PartnerBlockGeoRule, error) {
			return &models.PartnerBlock{
					ID:                blockID,
					PartnerSiteID:     siteID,
					Name:              "Sidebar",
					BlockType:         models.PartnerBlockTypeBanner,
					Status:            models.PartnerBlockStatusActive,
					CPMStrategy:       models.CPMStrategyMaxIncome,
					AmpMode:           models.AmpModeEnabled,
					SizeMode:          models.SizeModeAdaptive,
					BorderMode:        models.BorderModeAuto,
					CornerMode:        models.CornerModeRounded,
					Theme:             models.ThemeModeDark,
					InterscrollerMode: models.InterscrollerModeAuto,
					SelfAdSettings:    json.RawMessage(`{"reserved":true}`),
					CreatedAt:         time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				[]*models.PartnerBlockGeoRule{
					{GeoCode: "all", CPMV: sql.NullInt64{Int64: 200, Valid: true}},
					{GeoCode: "RU-MOW", IsEnabled: true},
				}, nil
		},
		updatePartnerBlockMetaFn: func(_ context.Context, partnerID, siteID int, in *serviceinput.UpdatePartnerBlockMeta) (*models.PartnerBlock, error) {
			return &models.PartnerBlock{ID: in.ID, PartnerSiteID: siteID, Name: "Renamed", BlockType: models.PartnerBlockTypeBanner, Status: models.PartnerBlockStatusActive}, nil
		},
		updatePartnerBlockGeneralFn: func(_ context.Context, partnerID, siteID int, in *serviceinput.UpdatePartnerBlockGeneral) (*models.PartnerBlock, error) {
			return &models.PartnerBlock{
				ID:                           in.ID,
				PartnerSiteID:                siteID,
				CPMStrategy:                  *in.CPMStrategy,
				AmpMode:                      *in.AmpMode,
				SizeMode:                     *in.SizeMode,
				BorderMode:                   *in.BorderMode,
				CornerMode:                   *in.CornerMode,
				Theme:                        *in.Theme,
				InterscrollerMode:            *in.InterscrollerMode,
				InterscrollerBackgroundColor: sql.NullString{String: "#fff", Valid: true},
			}, nil
		},
		updatePartnerBlockGeoFn: func(_ context.Context, partnerID, siteID int, in *serviceinput.UpdatePartnerBlockGeography) ([]*models.PartnerBlockGeoRule, error) {
			return []*models.PartnerBlockGeoRule{
				{GeoCode: "all", CPMV: sql.NullInt64{Int64: 250, Valid: true}},
				{GeoCode: "RU-MOW", IsEnabled: true},
			}, nil
		},
		updatePartnerBlockSelfAdFn: func(_ context.Context, partnerID, siteID int, in *serviceinput.UpdatePartnerBlockSelfAd) (*models.PartnerBlock, error) {
			return &models.PartnerBlock{ID: in.ID, SelfAdSettings: in.Settings}, nil
		},
		deletePartnerBlockFn: func(_ context.Context, partnerID, siteID, blockID int) error { return nil },
	}
	r := newTestRouter(ac, svc)
	csrf := getCSRF(t, r)
	ac.addSession("partner-block-session", 44)
	sess := &http.Cookie{Name: testCookieName, Value: "partner-block-session"}

	tests := []struct {
		method string
		path   string
		body   string
		want   string
	}{
		{method: http.MethodGet, path: "/partners/sites/101/blocks", want: `"name":"Sidebar"`},
		{method: http.MethodGet, path: "/partners/sites/101/blocks/2", want: `"supported_platforms":["desktop","mobile","amp"]`},
		{method: http.MethodPut, path: "/partners/sites/101/blocks/2/meta", body: `{"name":"Renamed","status":"active"}`, want: `"name":"Renamed"`},
		{method: http.MethodPut, path: "/partners/sites/101/blocks/2/general", body: `{"cpm_strategy":"max_income","amp_mode":"enabled","size_mode":"adaptive","border_mode":"auto","corner_mode":"rounded","theme":"dark","interscroller_mode":"auto","interscroller_background_color":"#fff","revenue_share_bps":1000}`, want: `"theme":"dark"`},
		{method: http.MethodPut, path: "/partners/sites/101/blocks/2/geography", body: `{"only_configured":true,"global_cpmv":250,"rules":[{"geo_code":"RU-MOW","is_enabled":true}]}`, want: `"global_cpmv":250`},
		{method: http.MethodPut, path: "/partners/sites/101/blocks/2/self-ad", body: `{"reserved":true}`, want: `"reserved":true`},
		{method: http.MethodDelete, path: "/partners/sites/101/blocks/2", want: `"message":"deleted"`},
	}

	for _, tc := range tests {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		req.AddCookie(sess)
		req.AddCookie(csrf)
		if tc.method != http.MethodGet {
			req.Header.Set("X-CSRF-Token", csrf.Value)
		}
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("%s %s: expected 200 got %d body=%s", tc.method, tc.path, rr.Code, rr.Body.String())
		}
		if !strings.Contains(rr.Body.String(), tc.want) {
			t.Fatalf("%s %s: body %s does not contain %s", tc.method, tc.path, rr.Body.String(), tc.want)
		}
	}
}

func muxSetVars(r *http.Request, val map[string]string) *http.Request {
	return mux.SetURLVars(r, val)
}
