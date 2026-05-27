package v1

import (
	"bytes"
	"context"
	"database/sql"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
)

func TestAppealHandlers(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{
		createAppealFn: func(_ context.Context, in *serviceinput.CreateAppeal) (*models.Appeal, error) {
			if in.AdvertiserID == nil || *in.AdvertiserID != 1 {
				t.Fatalf("expected optional advertiser binding, got %+v", in.AdvertiserID)
			}
			return &models.Appeal{ID: 11}, nil
		},
		listAppealsFn: func(_ context.Context, advertiserID int) ([]*models.Appeal, error) {
			return []*models.Appeal{{
				ID:           11,
				AdvertiserID: sql.NullInt64{Int64: int64(advertiserID), Valid: true},
				Status:       models.AppealStatusOpen,
				Category:     models.AppealCategoryBug,
				Title:        "Bug",
				Description:  "Broken",
				CreatedAt:    time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
			}}, nil
		},
		getAppealByIDFn: func(_ context.Context, appealID int) (*models.Appeal, error) {
			return &models.Appeal{
				ID:           appealID,
				AdvertiserID: sql.NullInt64{Int64: 1, Valid: true},
				Status:       models.AppealStatusOpen,
				Category:     models.AppealCategoryBug,
				Title:        "Bug",
				Description:  "Broken",
				CreatedAt:    time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
			}, nil
		},
	}
	r := newTestRouter(ac, svc)
	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, ac, 1)

	var form bytes.Buffer
	writer := multipart.NewWriter(&form)
	_ = writer.WriteField("category", "bug")
	_ = writer.WriteField("title", "Bug title")
	_ = writer.WriteField("description", "Broken thing")
	_ = writer.WriteField("name", "Ivan")
	_ = writer.WriteField("email", "ivan@test.dev")
	_ = writer.Close()

	createReq := httptest.NewRequest(http.MethodPost, "/appeals", &form)
	createReq.Header.Set("Content-Type", writer.FormDataContentType())
	createReq.AddCookie(sess)
	createReq.AddCookie(csrf)
	createReq.Header.Set("X-CSRF-Token", csrf.Value)
	createRR := httptest.NewRecorder()
	r.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated || !strings.Contains(createRR.Body.String(), `"id":11`) {
		t.Fatalf("create appeal failed code=%d body=%s", createRR.Code, createRR.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/appeals", nil)
	listReq.AddCookie(sess)
	listReq.AddCookie(csrf)
	listRR := httptest.NewRecorder()
	r.ServeHTTP(listRR, listReq)
	if listRR.Code != http.StatusOK || !strings.Contains(listRR.Body.String(), `"advertiser_id":1`) {
		t.Fatalf("list appeals failed code=%d body=%s", listRR.Code, listRR.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/appeals/11", nil)
	getReq.AddCookie(sess)
	getReq.AddCookie(csrf)
	getRR := httptest.NewRecorder()
	r.ServeHTTP(getRR, getReq)
	if getRR.Code != http.StatusOK || !strings.Contains(getRR.Body.String(), `"title":"Bug"`) {
		t.Fatalf("get appeal failed code=%d body=%s", getRR.Code, getRR.Body.String())
	}
}

func TestGetAppealByID_NotOwned(t *testing.T) {
	ac := newStubAuthClient()
	svc := &stubService{
		getAppealByIDFn: func(_ context.Context, appealID int) (*models.Appeal, error) {
			return &models.Appeal{ID: appealID, AdvertiserID: sql.NullInt64{Int64: 99, Valid: true}}, nil
		},
	}
	r := newTestRouter(ac, svc)
	csrf := getCSRF(t, r)
	sess := createSessionCookie(t, ac, 1)

	req := httptest.NewRequest(http.MethodGet, "/appeals/11", nil)
	req.AddCookie(sess)
	req.AddCookie(csrf)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 got %d body=%s", rr.Code, rr.Body.String())
	}
}
