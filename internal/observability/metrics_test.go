package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestHTTPMiddlewareUsesRouteTemplate(t *testing.T) {
	metrics := NewMetrics("app")
	router := mux.NewRouter()
	router.Use(metrics.HTTPMiddleware(HTTPMiddlewareConfig{}))
	router.HandleFunc("/widgets/{id}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}).Methods(http.MethodPost)

	req := httptest.NewRequest(http.MethodPost, "/widgets/42", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("unexpected status code: got %d want %d", rec.Code, http.StatusCreated)
	}

	counter := testutil.ToFloat64(metrics.httpRequestsTotal.WithLabelValues("app", http.MethodPost, "/widgets/{id}", "201"))
	if counter != 1 {
		t.Fatalf("unexpected request counter: got %v want 1", counter)
	}

	histogramMetric, err := metrics.httpRequestDuration.GetMetricWithLabelValues("app", http.MethodPost, "/widgets/{id}")
	if err != nil {
		t.Fatalf("get histogram metric: %v", err)
	}

	promMetric, ok := histogramMetric.(interface{ Write(*dto.Metric) error })
	if !ok {
		t.Fatalf("histogram metric does not expose prometheus write interface")
	}

	metricProto := &dto.Metric{}
	if err := promMetric.Write(metricProto); err != nil {
		t.Fatalf("write histogram metric: %v", err)
	}
	if metricProto.GetHistogram().GetSampleCount() != 1 {
		t.Fatalf("unexpected histogram sample count: got %d want 1", metricProto.GetHistogram().GetSampleCount())
	}
}

func TestUnaryServerInterceptorTracksErrors(t *testing.T) {
	metrics := NewMetrics("auth")
	interceptor := metrics.UnaryServerInterceptor()

	_, err := interceptor(
		context.Background(),
		struct{}{},
		&grpc.UnaryServerInfo{FullMethod: "/auth.v1.AuthService/Login"},
		func(_ context.Context, _ any) (any, error) {
			return nil, status.Error(codes.Internal, "boom")
		},
	)

	if status.Code(err) != codes.Internal {
		t.Fatalf("unexpected grpc code: got %s want %s", status.Code(err), codes.Internal)
	}

	total := testutil.ToFloat64(metrics.grpcRequestsTotal.WithLabelValues("auth", "auth.v1.AuthService", "Login", codes.Internal.String()))
	if total != 1 {
		t.Fatalf("unexpected request counter: got %v want 1", total)
	}

	errors := testutil.ToFloat64(metrics.grpcRequestErrorsTotal.WithLabelValues("auth", "auth.v1.AuthService", "Login", codes.Internal.String()))
	if errors != 1 {
		t.Fatalf("unexpected error counter: got %v want 1", errors)
	}
}
