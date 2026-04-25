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
	serviceinput "eshkere/internal/service/input"
)

func TestAdCampaign_UpdateListDelete(t *testing.T) {
	sm := newTestSessionManager()
	svc := &stubService{}
	r := newTestRouter(sm, svc)

	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, sm, 1)

	svc.updateAdCampaignFn = func(_ context.Context, in *serviceinput.UpdateAdCampaign) error {
		if in.ID != 5 || in.Name == nil || *in.Name != "new" {
			t.Fatalf("unexpected update input: %+v", in)
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

	svc.updateAdGroupFn = func(_ context.Context, in *serviceinput.UpdateAdGroup) error {
		if in.ID != 3 || in.Name == nil || *in.Name != "new-group" {
			t.Fatalf("unexpected update input: %+v", in)
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

	svc.updateAdFn = func(_ context.Context, in *serviceinput.UpdateAd) error {
		if in.ID != 8 || in.Title == nil || *in.Title != "renamed" {
			t.Fatalf("unexpected update input: %+v", in)
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

func TestAppeal_CreateMultipartWithoutScreenshot_OK(t *testing.T) {
	sm := newTestSessionManager()
	svc := &stubService{}
	r := newTestRouter(sm, svc)

	csrf := getCSRF(t, r)

	svc.createAppealFn = func(_ context.Context, in *serviceinput.CreateAppeal) (*models.Appeal, error) {
		if in.AdvertiserID != nil {
			t.Fatalf("expected guest appeal, got advertiser id %#v", in.AdvertiserID)
		}
		if in.Category != models.AppealCategoryQuestion || in.Title != "How to top up?" || in.Description != "Need help" {
			t.Fatalf("unexpected appeal payload: %+v", in)
		}
		if in.Name != "Ivan" || in.Email != "ivan@example.com" {
			t.Fatalf("unexpected contact payload: %+v", in)
		}
		if len(in.Image) != 0 {
			t.Fatalf("expected no image, got %d bytes", len(in.Image))
		}
		return &models.Appeal{ID: 21}, nil
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("category", "question"); err != nil {
		t.Fatalf("WriteField category: %v", err)
	}
	if err := writer.WriteField("title", "How to top up?"); err != nil {
		t.Fatalf("WriteField title: %v", err)
	}
	if err := writer.WriteField("description", "Need help"); err != nil {
		t.Fatalf("WriteField description: %v", err)
	}
	if err := writer.WriteField("name", "Ivan"); err != nil {
		t.Fatalf("WriteField name: %v", err)
	}
	if err := writer.WriteField("email", "ivan@example.com"); err != nil {
		t.Fatalf("WriteField email: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/appeal", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(csrf)
	req.Header.Set("X-CSRF-Token", csrf.Value)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 got %d body=%s", rr.Code, rr.Body.String())
	}

	var envelope struct {
		Data dto.CreateAppealResponse `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if envelope.Data.ID != 21 {
		t.Fatalf("expected appeal id 21 got %d", envelope.Data.ID)
	}
}

func TestAppeal_CreateMultipartWithScreenshot_OK(t *testing.T) {
	sm := newTestSessionManager()
	svc := &stubService{}
	r := newTestRouter(sm, svc)

	csrf := getCSRF(t, r)

	svc.createAppealFn = func(_ context.Context, in *serviceinput.CreateAppeal) (*models.Appeal, error) {
		if in.Category != models.AppealCategoryBug || in.Title != "Crash" || in.Description != "Steps to reproduce" {
			t.Fatalf("unexpected appeal payload: %+v", in)
		}
		if in.Name != "Ivan" || in.Email != "ivan@example.com" {
			t.Fatalf("unexpected contact payload: %+v", in)
		}
		if string(in.Image) != "png-bytes" || in.ImageExt != ".png" || in.ImageType != "image/png" {
			t.Fatalf("unexpected image payload: data=%q ext=%s type=%s", string(in.Image), in.ImageExt, in.ImageType)
		}
		return &models.Appeal{ID: 22}, nil
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("category", "bug"); err != nil {
		t.Fatalf("WriteField category: %v", err)
	}
	if err := writer.WriteField("title", "Crash"); err != nil {
		t.Fatalf("WriteField title: %v", err)
	}
	if err := writer.WriteField("description", "Steps to reproduce"); err != nil {
		t.Fatalf("WriteField description: %v", err)
	}
	if err := writer.WriteField("name", "Ivan"); err != nil {
		t.Fatalf("WriteField name: %v", err)
	}
	if err := writer.WriteField("email", "ivan@example.com"); err != nil {
		t.Fatalf("WriteField email: %v", err)
	}

	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "screenshot", "screen.png"))
	header.Set("Content-Type", "image/png")
	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatalf("CreatePart: %v", err)
	}
	if _, err = part.Write([]byte("png-bytes")); err != nil {
		t.Fatalf("Write screenshot: %v", err)
	}
	if err = writer.Close(); err != nil {
		t.Fatalf("Close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/appeal", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(csrf)
	req.Header.Set("X-CSRF-Token", csrf.Value)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 got %d body=%s", rr.Code, rr.Body.String())
	}

	var envelope struct {
		Data dto.CreateAppealResponse `json:"data"`
	}
	if err = json.Unmarshal(rr.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if envelope.Data.ID != 22 {
		t.Fatalf("expected appeal id 22 got %d", envelope.Data.ID)
	}
}

func TestAppeal_CreateJSONRejected(t *testing.T) {
	sm := newTestSessionManager()
	svc := &stubService{}
	r := newTestRouter(sm, svc)

	csrf := getCSRF(t, r)

	req := httptest.NewRequest(http.MethodPost, "/appeal", bytes.NewBufferString(`{"category":"question","title":"How to top up?","description":"Need help","name":"Ivan","email":"ivan@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(csrf)
	req.Header.Set("X-CSRF-Token", csrf.Value)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestAppeal_ListMessages_OK(t *testing.T) {
	sm := newTestSessionManager()
	svc := &stubService{}
	r := newTestRouter(sm, svc)

	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, sm, 1)

	svc.getAppealMessagesFn = func(_ context.Context, advertiserID, appealID int) ([]*models.AppealMessage, error) {
		if advertiserID != 1 || appealID != 7 {
			t.Fatalf("unexpected args: advertiser=%d appeal=%d", advertiserID, appealID)
		}
		return []*models.AppealMessage{{ID: 2, AppealID: 7, Author: models.AppealMessageAuthorAdmin, Text: "reply", CreatedAt: time.Unix(0, 0)}}, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/appeal/7/messages", nil)
	req.AddCookie(sess)
	req.AddCookie(csrf)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}

	var envelope struct {
		Data dto.ListAppealMessagesResponse `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if envelope.Data.AppealID != 7 || len(envelope.Data.Messages) != 1 || envelope.Data.Messages[0].Author != "admin" {
		t.Fatalf("unexpected response: %+v", envelope.Data)
	}
}

func TestAppeal_PostMessage_OK(t *testing.T) {
	sm := newTestSessionManager()
	svc := &stubService{}
	r := newTestRouter(sm, svc)

	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, sm, 1)

	svc.postAppealMessageFn = func(_ context.Context, in *serviceinput.PostAppealMessage) (*models.AppealMessage, error) {
		if in.AdvertiserID != 1 || in.AppealID != 7 || in.Text != "hello" {
			t.Fatalf("unexpected input: %+v", in)
		}
		return &models.AppealMessage{ID: 3, AppealID: 7, Author: models.AppealMessageAuthorAdvertiser, Text: "hello", CreatedAt: time.Unix(0, 0)}, nil
	}

	req := httptest.NewRequest(http.MethodPost, "/appeal/7/messages", bytes.NewBufferString(`{"text":"hello"}`))
	req.AddCookie(sess)
	req.AddCookie(csrf)
	req.Header.Set("X-CSRF-Token", csrf.Value)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 got %d body=%s", rr.Code, rr.Body.String())
	}

	var envelope struct {
		Data dto.AppealMessageResponse `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if envelope.Data.ID != 3 || envelope.Data.Author != "advertiser" {
		t.Fatalf("unexpected response: %+v", envelope.Data)
	}
}
