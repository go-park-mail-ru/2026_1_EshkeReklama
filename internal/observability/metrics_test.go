package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestHTTPMiddlewareTracksInFlightRequests(t *testing.T) {
	metrics := NewMetrics("app")
	started := make(chan struct{})
	release := make(chan struct{})

	router := mux.NewRouter()
	router.Use(metrics.HTTPMiddleware(HTTPMiddlewareConfig{}))
	router.HandleFunc("/widgets/{id}", func(w http.ResponseWriter, _ *http.Request) {
		close(started)
		<-release
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodGet)

	done := make(chan struct{})
	go func() {
		defer close(done)
		req := httptest.NewRequest(http.MethodGet, "/widgets/42", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("handler did not start")
	}

	inFlight := testutil.ToFloat64(metrics.httpRequestsInFlight.WithLabelValues("app", http.MethodGet, "/widgets/{id}"))
	if inFlight != 1 {
		t.Fatalf("unexpected in-flight gauge during request: got %v want 1", inFlight)
	}

	close(release)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("handler did not finish")
	}

	inFlight = testutil.ToFloat64(metrics.httpRequestsInFlight.WithLabelValues("app", http.MethodGet, "/widgets/{id}"))
	if inFlight != 0 {
		t.Fatalf("unexpected in-flight gauge after request: got %v want 0", inFlight)
	}
}

func TestUnaryServerInterceptorTracksInFlightRequests(t *testing.T) {
	metrics := NewMetrics("auth")
	interceptor := metrics.UnaryServerInterceptor()
	started := make(chan struct{})
	release := make(chan struct{})

	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = interceptor(
			context.Background(),
			struct{}{},
			&grpc.UnaryServerInfo{FullMethod: "/auth.v1.AuthService/Login"},
			func(_ context.Context, _ any) (any, error) {
				close(started)
				<-release
				return struct{}{}, nil
			},
		)
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("grpc handler did not start")
	}

	inFlight := testutil.ToFloat64(metrics.grpcRequestsInFlight.WithLabelValues("auth", "auth.v1.AuthService", "Login"))
	if inFlight != 1 {
		t.Fatalf("unexpected grpc in-flight gauge during request: got %v want 1", inFlight)
	}

	close(release)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("grpc handler did not finish")
	}

	inFlight = testutil.ToFloat64(metrics.grpcRequestsInFlight.WithLabelValues("auth", "auth.v1.AuthService", "Login"))
	if inFlight != 0 {
		t.Fatalf("unexpected grpc in-flight gauge after request: got %v want 0", inFlight)
	}
}
