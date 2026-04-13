package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"eshkere/internal/handler/dto"
	"eshkere/internal/models"
	"eshkere/internal/session"

	"go.uber.org/mock/gomock"
)

func createSessionCookie(t *testing.T, sm *session.Manager, advertiserID int) *http.Cookie {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rr := httptest.NewRecorder()
	if err := sm.Create(rr, req, advertiserID); err != nil {
		t.Fatalf("Create session: %v", err)
	}
	cookies := rr.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected session cookie")
	}
	return cookies[0]
}

func TestAdCampaign_CRUD(t *testing.T) {
	sm := newTestSessionManager()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := NewMockService(ctrl)
	r := newTestRouter(sm, svc)

	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, sm, 1)

	svc.EXPECT().
		CreateAdCampaign(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, c *models.AdCampaign) (*models.AdCampaign, error) {
			if c.AdvertiserID != 1 || c.Name != "camp" || c.DailyBudget != 10 {
				t.Fatalf("unexpected campaign: %+v", c)
			}
			c.ID = 42
			return c, nil
		})

	createReq := httptest.NewRequest(http.MethodPost, "/ad_campaigns", bytes.NewBufferString(`{"name":"camp","daily_budget":10}`))
	createReq.AddCookie(sess)
	createReq.AddCookie(csrf)
	createReq.Header.Set("X-CSRF-Token", csrf.Value)
	createRR := httptest.NewRecorder()
	r.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", createRR.Code, createRR.Body.String())
	}

	svc.EXPECT().
		ListAdCampaigns(gomock.Any(), 1).
		Return([]*models.AdCampaign{{ID: 1, AdvertiserID: 1, Status: models.AdStatusWorking, Name: "c", DailyBudget: 7}}, nil)

	listReq := httptest.NewRequest(http.MethodGet, "/ad_campaigns", nil)
	listReq.AddCookie(sess)
	listReq.AddCookie(csrf)
	listRR := httptest.NewRecorder()
	r.ServeHTTP(listRR, listReq)
	if listRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", listRR.Code, listRR.Body.String())
	}

	newName := "new"
	budget := int64(99)
	svc.EXPECT().
		UpdateAdCampaign(gomock.Any(), 42, dto.UpdateAdCampaignRequest{Name: &newName, DailyBudget: &budget}).
		Return(nil)

	updReq := httptest.NewRequest(http.MethodPut, "/ad_campaigns/42", bytes.NewBufferString(`{"name":"new","daily_budget":99}`))
	updReq.AddCookie(sess)
	updReq.AddCookie(csrf)
	updReq.Header.Set("X-CSRF-Token", csrf.Value)
	updRR := httptest.NewRecorder()
	r.ServeHTTP(updRR, updReq)
	if updRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", updRR.Code, updRR.Body.String())
	}

	svc.EXPECT().DeleteAdCampaign(gomock.Any(), 42).Return(nil)

	delReq := httptest.NewRequest(http.MethodDelete, "/ad_campaigns/42", nil)
	delReq.AddCookie(sess)
	delReq.AddCookie(csrf)
	delReq.Header.Set("X-CSRF-Token", csrf.Value)
	delRR := httptest.NewRecorder()
	r.ServeHTTP(delRR, delReq)
	if delRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", delRR.Code, delRR.Body.String())
	}
}

func TestAdGroup_And_Ads_CRUD(t *testing.T) {
	sm := newTestSessionManager()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := NewMockService(ctrl)
	r := newTestRouter(sm, svc)

	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, sm, 1)

	svc.EXPECT().
		CreateAdGroup(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, g *models.AdGroup) (*models.AdGroup, error) {
			if g.AdCampaignID != 10 || g.Name != "g" {
				t.Fatalf("unexpected group: %+v", g)
			}
			g.ID = 5
			return g, nil
		})

	createGroupReq := httptest.NewRequest(http.MethodPost, "/ad_campaigns/10/ad_groups", bytes.NewBufferString(`{"topic_id":1,"region_id":2,"name":"g","age_from":18,"age_to":25,"gender":"any"}`))
	createGroupReq.AddCookie(sess)
	createGroupReq.AddCookie(csrf)
	createGroupReq.Header.Set("X-CSRF-Token", csrf.Value)
	createGroupRR := httptest.NewRecorder()
	r.ServeHTTP(createGroupRR, createGroupReq)
	if createGroupRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", createGroupRR.Code, createGroupRR.Body.String())
	}

	svc.EXPECT().ListAdGroups(gomock.Any(), 10).Return([]*models.AdGroup{}, nil)
	listGroupReq := httptest.NewRequest(http.MethodGet, "/ad_campaigns/10/ad_groups", nil)
	listGroupReq.AddCookie(sess)
	listGroupReq.AddCookie(csrf)
	listGroupRR := httptest.NewRecorder()
	r.ServeHTTP(listGroupRR, listGroupReq)
	if listGroupRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", listGroupRR.Code, listGroupRR.Body.String())
	}

	svc.EXPECT().UpdateAdGroup(gomock.Any(), 5, gomock.Any()).Return(nil)
	updGroupReq := httptest.NewRequest(http.MethodPut, "/ad_campaigns/10/ad_groups/5", bytes.NewBufferString(`{"name":"g2"}`))
	updGroupReq.AddCookie(sess)
	updGroupReq.AddCookie(csrf)
	updGroupReq.Header.Set("X-CSRF-Token", csrf.Value)
	updGroupRR := httptest.NewRecorder()
	r.ServeHTTP(updGroupRR, updGroupReq)
	if updGroupRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", updGroupRR.Code, updGroupRR.Body.String())
	}

	svc.EXPECT().DeleteAdGroup(gomock.Any(), 5).Return(nil)
	delGroupReq := httptest.NewRequest(http.MethodDelete, "/ad_campaigns/10/ad_groups/5", nil)
	delGroupReq.AddCookie(sess)
	delGroupReq.AddCookie(csrf)
	delGroupReq.Header.Set("X-CSRF-Token", csrf.Value)
	delGroupRR := httptest.NewRecorder()
	r.ServeHTTP(delGroupRR, delGroupReq)
	if delGroupRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", delGroupRR.Code, delGroupRR.Body.String())
	}

	svc.EXPECT().
		CreateAd(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, ad *models.Ad) (*models.Ad, error) {
			if ad.AdGroupID != 2 || ad.Title != "t" {
				t.Fatalf("unexpected ad: %+v", ad)
			}
			ad.ID = 9
			return ad, nil
		})

	createAdReq := httptest.NewRequest(http.MethodPost, "/ad_campaigns/1/ad_groups/2/ads", bytes.NewBufferString(`{"title":"t","short_desc":"s","image_url":"i","target_url":"u"}`))
	createAdReq.AddCookie(sess)
	createAdReq.AddCookie(csrf)
	createAdReq.Header.Set("X-CSRF-Token", csrf.Value)
	createAdRR := httptest.NewRecorder()
	r.ServeHTTP(createAdRR, createAdReq)
	if createAdRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", createAdRR.Code, createAdRR.Body.String())
	}

	svc.EXPECT().UpdateAd(gomock.Any(), 9, gomock.Any()).Return(nil)
	updAdReq := httptest.NewRequest(http.MethodPut, "/ad_campaigns/1/ad_groups/2/ads/9", bytes.NewBufferString(`{"title":"t2"}`))
	updAdReq.AddCookie(sess)
	updAdReq.AddCookie(csrf)
	updAdReq.Header.Set("X-CSRF-Token", csrf.Value)
	updAdRR := httptest.NewRecorder()
	r.ServeHTTP(updAdRR, updAdReq)
	if updAdRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", updAdRR.Code, updAdRR.Body.String())
	}

	svc.EXPECT().DeleteAd(gomock.Any(), 9).Return(nil)
	delAdReq := httptest.NewRequest(http.MethodDelete, "/ad_campaigns/1/ad_groups/2/ads/9", nil)
	delAdReq.AddCookie(sess)
	delAdReq.AddCookie(csrf)
	delAdReq.Header.Set("X-CSRF-Token", csrf.Value)
	delAdRR := httptest.NewRecorder()
	r.ServeHTTP(delAdRR, delAdReq)
	if delAdRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", delAdRR.Code, delAdRR.Body.String())
	}
}
