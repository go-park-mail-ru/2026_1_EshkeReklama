package observability

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const namespace = "eshkere"

var durationBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5}

type Metrics struct {
	service string

	registry *prometheus.Registry

	httpRequestsTotal      *prometheus.CounterVec
	httpRequestErrorsTotal *prometheus.CounterVec
	httpRequestDuration    *prometheus.HistogramVec
	grpcRequestsTotal      *prometheus.CounterVec
	grpcRequestErrorsTotal *prometheus.CounterVec
	grpcRequestDuration    *prometheus.HistogramVec
	serviceInfo            *prometheus.GaugeVec
}

type HTTPMiddlewareConfig struct {
	SkipPaths []string
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (r *responseWriter) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func NewMetrics(service string) *Metrics {
	registry := prometheus.NewRegistry()
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	metrics := &Metrics{
		service:  service,
		registry: registry,
		httpRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests handled by the service.",
		}, []string{"service", "method", "route", "status"}),
		httpRequestErrorsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_request_errors_total",
			Help:      "Total number of failed HTTP requests handled by the service.",
		}, []string{"service", "method", "route", "status"}),
		httpRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "Latency distribution of HTTP requests handled by the service.",
			Buckets:   durationBuckets,
		}, []string{"service", "method", "route"}),
		grpcRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "grpc_requests_total",
			Help:      "Total number of gRPC requests handled by the service.",
		}, []string{"service", "grpc_service", "grpc_method", "code"}),
		grpcRequestErrorsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "grpc_request_errors_total",
			Help:      "Total number of failed gRPC requests handled by the service.",
		}, []string{"service", "grpc_service", "grpc_method", "code"}),
		grpcRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "grpc_request_duration_seconds",
			Help:      "Latency distribution of gRPC requests handled by the service.",
			Buckets:   durationBuckets,
		}, []string{"service", "grpc_service", "grpc_method"}),
		serviceInfo: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "service_info",
			Help:      "Static information about the running service.",
		}, []string{"service"}),
	}

	registry.MustRegister(
		metrics.httpRequestsTotal,
		metrics.httpRequestErrorsTotal,
		metrics.httpRequestDuration,
		metrics.grpcRequestsTotal,
		metrics.grpcRequestErrorsTotal,
		metrics.grpcRequestDuration,
		metrics.serviceInfo,
	)
	metrics.serviceInfo.WithLabelValues(service).Set(1)

	return metrics
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

func (m *Metrics) HTTPMiddleware(cfg HTTPMiddlewareConfig) func(http.Handler) http.Handler {
	skipPaths := make(map[string]struct{}, len(cfg.SkipPaths))
	for _, path := range cfg.SkipPaths {
		skipPaths[path] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, skip := skipPaths[r.URL.Path]; skip {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			rec := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rec, r)

			route := routeTemplate(r)
			statusCode := httpStatusLabel(rec.statusCode)

			m.httpRequestsTotal.WithLabelValues(
				m.service,
				r.Method,
				route,
				statusCode,
			).Inc()
			m.httpRequestDuration.WithLabelValues(m.service, r.Method, route).Observe(time.Since(start).Seconds())

			if rec.statusCode >= http.StatusInternalServerError {
				m.httpRequestErrorsTotal.WithLabelValues(
					m.service,
					r.Method,
					route,
					statusCode,
				).Inc()
			}
		})
	}
}

func (m *Metrics) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		start := time.Now()

		resp, err = handler(ctx, req)

		grpcService, grpcMethod := splitGRPCMethod(info.FullMethod)
		code := status.Code(err).String()

		m.grpcRequestsTotal.WithLabelValues(m.service, grpcService, grpcMethod, code).Inc()
		m.grpcRequestDuration.WithLabelValues(m.service, grpcService, grpcMethod).Observe(time.Since(start).Seconds())
		if code != "OK" {
			m.grpcRequestErrorsTotal.WithLabelValues(m.service, grpcService, grpcMethod, code).Inc()
		}

		return resp, err
	}
}

func NewMetricsServer(addr string, metrics *Metrics) *http.Server {
	if addr == "" || metrics == nil {
		return nil
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", metrics.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	return &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func routeTemplate(r *http.Request) string {
	route := mux.CurrentRoute(r)
	if route == nil {
		return "unmatched"
	}

	if template, err := route.GetPathTemplate(); err == nil && template != "" {
		return template
	}
	if regexp, err := route.GetPathRegexp(); err == nil && regexp != "" {
		return regexp
	}

	return "unmatched"
}

func splitGRPCMethod(fullMethod string) (string, string) {
	trimmed := strings.TrimPrefix(fullMethod, "/")
	service, method, found := strings.Cut(trimmed, "/")
	if !found || service == "" || method == "" {
		return "unknown", "unknown"
	}
	return service, method
}

func httpStatusLabel(statusCode int) string {
	if statusCode == 0 {
		return "unknown"
	}
	return strconv.Itoa(statusCode)
}
