package v1

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"
	"time"

	"eshkere/internal/handler/v1/dto"
	"eshkere/internal/models"
)

func TestAdCampaign_UpdateListDelete(t *testing.T) {
	sm := newTestSessionManager()
	svc := &stubService{}
	r := newTestRouter(sm, svc)

	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, sm, 1)

	svc.updateAdCampaignFn = func(_ context.Context, campaignID int, req *dto.UpdateAdCampaignRequest) error {
		if campaignID != 5 || req == nil || req.Name == nil || *req.Name != "new" {
			t.Fatalf("unexpected update args: campaignID=%d req=%+v", campaignID, req)
		}
		return nil
	}
	svc.listAdCampaignsFn = func(_ context.Context, advertiserID int) ([]*models.AdCampaign, error) {
		if advertiserID != 1 {
			t.Fatalf("unexpected advertiser id: %d", advertiserID)
		}
		return []*models.AdCampaign{{ID: 5, AdvertiserID: 1, Name: "camp", Status: models.AdStatusWorking, DailyBudget: 42}}, nil
	}
	svc.deleteAdCampaignFn = func(_ context.Context, campaignID int) error {
		if campaignID != 5 {
			t.Fatalf("unexpected delete id: %d", campaignID)
		}
		return nil
	}

	updateReq := httptest.NewRequest(http.MethodPut, "/ad_campaigns/5", bytes.NewBufferString(`{"name":"new"}`))
	updateReq.AddCookie(sess)
	updateReq.AddCookie(csrf)
	updateReq.Header.Set("X-CSRF-Token", csrf.Value)
	updateRR := httptest.NewRecorder()
	r.ServeHTTP(updateRR, updateReq)
	if updateRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", updateRR.Code, updateRR.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/ad_campaigns", nil)
	listReq.AddCookie(sess)
	listReq.AddCookie(csrf)
	listRR := httptest.NewRecorder()
	r.ServeHTTP(listRR, listReq)
	if listRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", listRR.Code, listRR.Body.String())
	}
	var listEnvelope struct {
		Data dto.ListAdCampaignsResponse `json:"data"`
	}
	if err := json.Unmarshal(listRR.Body.Bytes(), &listEnvelope); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(listEnvelope.Data.Campaigns) != 1 || listEnvelope.Data.Campaigns[0].ID != 5 {
		t.Fatalf("unexpected campaigns: %+v", listEnvelope.Data.Campaigns)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/ad_campaigns/5", nil)
	deleteReq.AddCookie(sess)
	deleteReq.AddCookie(csrf)
	deleteReq.Header.Set("X-CSRF-Token", csrf.Value)
	deleteRR := httptest.NewRecorder()
	r.ServeHTTP(deleteRR, deleteReq)
	if deleteRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", deleteRR.Code, deleteRR.Body.String())
	}
}

func TestAdGroup_UpdateListDelete(t *testing.T) {
	sm := newTestSessionManager()
	svc := &stubService{}
	r := newTestRouter(sm, svc)

	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, sm, 1)

	svc.updateAdGroupFn = func(_ context.Context, groupID int, req *dto.UpdateAdGroupRequest) error {
		if groupID != 3 || req == nil || req.Name == nil || *req.Name != "new-group" {
			t.Fatalf("unexpected update args: groupID=%d req=%+v", groupID, req)
		}
		return nil
	}
	svc.listAdGroupsFn = func(_ context.Context, campaignID int) ([]*models.AdGroup, error) {
		if campaignID != 9 {
			t.Fatalf("unexpected campaign id: %d", campaignID)
		}
		return []*models.AdGroup{{ID: 3, AdCampaignID: 9, Name: "g", TopicID: 1, RegionID: 2, AgeFrom: 18, AgeTo: 30, Gender: models.GenderAny}}, nil
	}
	svc.deleteAdGroupFn = func(_ context.Context, groupID int) error {
		if groupID != 3 {
			t.Fatalf("unexpected delete id: %d", groupID)
		}
		return nil
	}

	updateReq := httptest.NewRequest(http.MethodPut, "/ad_campaigns/9/ad_groups/3", bytes.NewBufferString(`{"name":"new-group"}`))
	updateReq.AddCookie(sess)
	updateReq.AddCookie(csrf)
	updateReq.Header.Set("X-CSRF-Token", csrf.Value)
	updateRR := httptest.NewRecorder()
	r.ServeHTTP(updateRR, updateReq)
	if updateRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", updateRR.Code, updateRR.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/ad_campaigns/9/ad_groups", nil)
	listReq.AddCookie(sess)
	listReq.AddCookie(csrf)
	listRR := httptest.NewRecorder()
	r.ServeHTTP(listRR, listReq)
	if listRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", listRR.Code, listRR.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/ad_campaigns/9/ad_groups/3", nil)
	deleteReq.AddCookie(sess)
	deleteReq.AddCookie(csrf)
	deleteReq.Header.Set("X-CSRF-Token", csrf.Value)
	deleteRR := httptest.NewRecorder()
	r.ServeHTTP(deleteRR, deleteReq)
	if deleteRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", deleteRR.Code, deleteRR.Body.String())
	}
}

func TestAd_UpdateDelete(t *testing.T) {
	sm := newTestSessionManager()
	svc := &stubService{}
	r := newTestRouter(sm, svc)

	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, sm, 1)

	svc.updateAdFn = func(_ context.Context, adID int, req *dto.UpdateAdRequest) error {
		if adID != 8 || req == nil || req.Title == nil || *req.Title != "renamed" {
			t.Fatalf("unexpected update args: adID=%d req=%+v", adID, req)
		}
		return nil
	}
	svc.deleteAdFn = func(_ context.Context, adID int) error {
		if adID != 8 {
			t.Fatalf("unexpected delete id: %d", adID)
		}
		return nil
	}

	updateReq := httptest.NewRequest(http.MethodPut, "/ad_campaigns/1/ad_groups/2/ads/8", bytes.NewBufferString(`{"title":"renamed"}`))
	updateReq.AddCookie(sess)
	updateReq.AddCookie(csrf)
	updateReq.Header.Set("X-CSRF-Token", csrf.Value)
	updateRR := httptest.NewRecorder()
	r.ServeHTTP(updateRR, updateReq)
	if updateRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", updateRR.Code, updateRR.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/ad_campaigns/1/ad_groups/2/ads/8", nil)
	deleteReq.AddCookie(sess)
	deleteReq.AddCookie(csrf)
	deleteReq.Header.Set("X-CSRF-Token", csrf.Value)
	deleteRR := httptest.NewRecorder()
	r.ServeHTTP(deleteRR, deleteReq)
	if deleteRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", deleteRR.Code, deleteRR.Body.String())
	}
}

func TestAdvertiser_UpdateAvatar_OK(t *testing.T) {
	sm := newTestSessionManager()
	svc := &stubService{}
	r := newTestRouter(sm, svc)

	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, sm, 1)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "avatar", "avatar.png"))
	header.Set("Content-Type", "image/png")
	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err = part.Write([]byte("png-bytes")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err = writer.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	svc.updateAdvertiserAvatarFn = func(_ context.Context, advertiserID int, avatar []byte, avatarExt, avatarContentType string) (*models.Advertiser, error) {
		if advertiserID != 1 {
			t.Fatalf("unexpected advertiser id: %d", advertiserID)
		}
		if string(avatar) != "png-bytes" || avatarExt != ".png" || avatarContentType != "image/png" {
			t.Fatalf("unexpected avatar payload: data=%q ext=%s contentType=%s", string(avatar), avatarExt, avatarContentType)
		}
		return &models.Advertiser{
			ID:        1,
			Name:      "name",
			Email:     "a@a.test",
			Phone:     "9001234567",
			AvatarURL: sql.NullString{String: "https://cdn/avatar.png", Valid: true},
			CreatedAt: time.Unix(0, 0),
		}, nil
	}

	req := httptest.NewRequest(http.MethodPut, "/advertiser/me/avatar", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(sess)
	req.AddCookie(csrf)
	req.Header.Set("X-CSRF-Token", csrf.Value)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}
}
