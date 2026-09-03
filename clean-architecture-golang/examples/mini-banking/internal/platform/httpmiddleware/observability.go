package httpmiddleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type requestIDKey struct{}
type traceIDKey struct{}

func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

func TraceID(ctx context.Context) string {
	id, _ := ctx.Value(traceIDKey{}).(string)
	return id
}

type metricValue struct {
	count       uint64
	durationSum time.Duration
}

type Metrics struct {
	mu     sync.RWMutex
	values map[string]metricValue
}

func NewMetrics() *Metrics {
	return &Metrics{values: make(map[string]metricValue)}
}

func (m *Metrics) record(method, route string, status int, duration time.Duration) {
	if m == nil {
		return
	}
	key := method + "\x00" + route + "\x00" + strconv.Itoa(status)
	m.mu.Lock()
	value := m.values[key]
	value.count++
	value.durationSum += duration
	m.values[key] = value
	m.mu.Unlock()
}

func (m *Metrics) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	m.mu.RLock()
	keys := make([]string, 0, len(m.values))
	for key := range m.values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	values := make(map[string]metricValue, len(m.values))
	for _, key := range keys {
		values[key] = m.values[key]
	}
	m.mu.RUnlock()

	_, _ = fmt.Fprintln(w, "# TYPE http_server_requests_total counter")
	_, _ = fmt.Fprintln(w, "# TYPE http_server_request_duration_seconds_sum counter")
	for _, key := range keys {
		parts := strings.Split(key, "\x00")
		value := values[key]
		labels := fmt.Sprintf(
			"method=%q,route=%q,status=%q",
			parts[0], parts[1], parts[2],
		)
		_, _ = fmt.Fprintf(w, "http_server_requests_total{%s} %d\n", labels, value.count)
		_, _ = fmt.Fprintf(
			w,
			"http_server_request_duration_seconds_sum{%s} %.6f\n",
			labels,
			value.durationSum.Seconds(),
		)
	}
}

func ObserveRequests(logger *slog.Logger, next http.Handler) http.Handler {
	return ObserveRequestsWithMetrics(logger, nil, next)
}

func ObserveRequestsWithMetrics(
	logger *slog.Logger,
	metrics *Metrics,
	next http.Handler,
) http.Handler {
	if logger == nil || next == nil {
		panic("http middleware: nil dependency")
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := acceptedRequestID(r.Header.Get("X-Request-ID"))
		if requestID == "" {
			requestID = randomHex(16)
		}
		traceID := traceIDFromTraceparent(r.Header.Get("traceparent"))
		if traceID == "" {
			traceID = randomHex(16)
		}
		w.Header().Set("X-Request-ID", requestID)
		w.Header().Set("X-Trace-ID", traceID)
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		ctx := context.WithValue(r.Context(), requestIDKey{}, requestID)
		ctx = context.WithValue(ctx, traceIDKey{}, traceID)
		request := r.WithContext(ctx)
		next.ServeHTTP(recorder, request)

		duration := time.Since(start)
		route := boundedRoute(request.URL.Path)
		metrics.record(r.Method, route, recorder.status, duration)
		logger.InfoContext(ctx, "HTTP request completed",
			"request_id", requestID,
			"trace_id", traceID,
			"method", r.Method,
			"route", route,
			"status", recorder.status,
			"response_bytes", recorder.bytes,
			"duration_ms", duration.Milliseconds(),
		)
	})
}

func boundedRoute(path string) string {
	switch path {
	case "/healthz", "/transfers", "/metrics":
		return path
	}
	if strings.HasPrefix(path, "/accounts/") && strings.HasSuffix(path, "/transfers") {
		return "/accounts/{accountID}/transfers"
	}
	return "unmatched"
}

type statusRecorder struct {
	http.ResponseWriter
	status      int
	bytes       int
	wroteHeader bool
}

func (w *statusRecorder) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func (w *statusRecorder) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusRecorder) Write(body []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(body)
	w.bytes += n
	return n, err
}

func acceptedRequestID(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > 128 {
		return ""
	}
	return raw
}

func traceIDFromTraceparent(value string) string {
	parts := strings.Split(strings.TrimSpace(value), "-")
	if len(parts) != 4 || parts[0] != "00" || len(parts[1]) != 32 {
		return ""
	}
	if _, err := hex.DecodeString(parts[1]); err != nil || parts[1] == strings.Repeat("0", 32) {
		return ""
	}
	return strings.ToLower(parts[1])
}

func randomHex(size int) string {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		return "correlation-id-unavailable"
	}
	return hex.EncodeToString(value)
}
