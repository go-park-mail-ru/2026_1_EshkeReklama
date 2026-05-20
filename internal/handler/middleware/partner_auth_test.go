package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	errs "eshkere/internal/errors"
	"eshkere/pkg/ctxutils"
)

func TestPartnerAuth(t *testing.T) {
	type testCase struct {
		name       string
		cookie     *http.Cookie
		validate   func(context.Context, string) (int64, error)
		wantStatus int
		wantID     int
	}

	cases := []testCase{
		{
			name:       "missing cookie",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:   "session not found",
			cookie: &http.Cookie{Name: "sid", Value: "bad"},
			validate: func(context.Context, string) (int64, error) {
				return 0, errs.ErrSessionNotFound
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:   "internal error",
			cookie: &http.Cookie{Name: "sid", Value: "bad"},
			validate: func(context.Context, string) (int64, error) {
				return 0, errors.New("boom")
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:   "success",
			cookie: &http.Cookie{Name: "sid", Value: "ok"},
			validate: func(_ context.Context, sessionID string) (int64, error) {
				if sessionID != "ok" {
					t.Fatalf("unexpected session id: %s", sessionID)
				}
				return 17, nil
			},
			wantStatus: http.StatusOK,
			wantID:     17,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			validator := &partnerStubValidator{validate: tc.validate}
			handler := PartnerAuth(validator, "sid")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				id, err := ctxutils.PartnerIDFromContext(r.Context())
				if err != nil {
					t.Fatalf("partner id from context: %v", err)
				}
				_ = json.NewEncoder(w).Encode(map[string]int{"id": id})
			}))

			req := httptest.NewRequest(http.MethodGet, "/partners/me", nil)
			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Fatalf("unexpected status: got %d want %d", rr.Code, tc.wantStatus)
			}
			if tc.wantStatus == http.StatusOK {
				var body map[string]int
				if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if body["id"] != tc.wantID {
					t.Fatalf("unexpected id: %v", body)
				}
			}
		})
	}
}

type partnerStubValidator struct {
	validate func(context.Context, string) (int64, error)
}

func (s *partnerStubValidator) ValidateSession(ctx context.Context, sid string) (int64, error) {
	if s.validate == nil {
		return 0, nil
	}
	return s.validate(ctx, sid)
}
