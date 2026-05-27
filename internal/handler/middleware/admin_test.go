package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"eshkere/internal/models"
	"eshkere/pkg/ctxutils"
)

type stubAdminChecker struct {
	getByIDFn func(ctx context.Context, id int) (*models.Advertiser, error)
}

func (s stubAdminChecker) GetAdvertiserByID(ctx context.Context, id int) (*models.Advertiser, error) {
	return s.getByIDFn(ctx, id)
}

func TestIsAdminMiddleware(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusNoContent)
	})

	run := func(ctx context.Context, checker AdminChecker) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/admin", nil).WithContext(ctx)
		rr := httptest.NewRecorder()
		IsAdmin(checker)(next).ServeHTTP(rr, req)
		return rr
	}

	if rr := run(context.Background(), nil); rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for nil checker, got %d", rr.Code)
	}
	if rr := run(context.Background(), stubAdminChecker{getByIDFn: func(_ context.Context, _ int) (*models.Advertiser, error) {
		return &models.Advertiser{Role: models.AdvertiserRoleAdmin}, nil
	}}); rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without advertiser id, got %d", rr.Code)
	}

	ctx := context.WithValue(context.Background(), ctxutils.AdvertiserIDKey, 1)
	if rr := run(ctx, stubAdminChecker{getByIDFn: func(_ context.Context, _ int) (*models.Advertiser, error) {
		return nil, errors.New("db down")
	}}); rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 on repo error, got %d", rr.Code)
	}
	if rr := run(ctx, stubAdminChecker{getByIDFn: func(_ context.Context, _ int) (*models.Advertiser, error) {
		return nil, nil
	}}); rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for nil advertiser, got %d", rr.Code)
	}
	if rr := run(ctx, stubAdminChecker{getByIDFn: func(_ context.Context, _ int) (*models.Advertiser, error) {
		return &models.Advertiser{Role: models.AdvertiserRoleUser}, nil
	}}); rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for non-admin, got %d", rr.Code)
	}
	if rr := run(ctx, stubAdminChecker{getByIDFn: func(_ context.Context, _ int) (*models.Advertiser, error) {
		return &models.Advertiser{Role: models.AdvertiserRoleAdmin}, nil
	}}); rr.Code != http.StatusNoContent || !nextCalled {
		t.Fatalf("expected next handler for admin, code=%d next=%v", rr.Code, nextCalled)
	}
}
