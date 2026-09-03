package httpmiddleware

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestObserveRequestsPropagatesAndLogsRequestID(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	var contextID string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextID = RequestID(r.Context())
		w.WriteHeader(http.StatusCreated)
	})
	request := httptest.NewRequest(http.MethodPost, "/transfers", nil)
	request.Header.Set("X-Request-ID", "REQ-123")
	response := httptest.NewRecorder()

	ObserveRequests(logger, next).ServeHTTP(response, request)

	if response.Code != http.StatusCreated || response.Header().Get("X-Request-ID") != "REQ-123" {
		t.Fatalf("code=%d request-id=%q", response.Code, response.Header().Get("X-Request-ID"))
	}
	if contextID != "REQ-123" || !strings.Contains(output.String(), `"status":201`) {
		t.Fatalf("context-id=%q log=%s", contextID, output.String())
	}
}

func TestObserveRequestsPropagatesTraceAndExportsBoundedMetrics(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	metrics := NewMetrics()
	var gotTraceID string
	mux := http.NewServeMux()
	mux.HandleFunc("GET /accounts/{id}/transfers", func(w http.ResponseWriter, r *http.Request) {
		gotTraceID = TraceID(r.Context())
		w.WriteHeader(http.StatusNoContent)
	})
	request := httptest.NewRequest(http.MethodGet, "/accounts/A-100/transfers", nil)
	request.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	response := httptest.NewRecorder()

	ObserveRequestsWithMetrics(logger, metrics, mux).ServeHTTP(response, request)

	if gotTraceID != "4bf92f3577b34da6a3ce929d0e0e4736" {
		t.Fatalf("trace id=%q", gotTraceID)
	}
	metricsResponse := httptest.NewRecorder()
	metrics.ServeHTTP(metricsResponse, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := metricsResponse.Body.String()
	if !strings.Contains(body, `route="/accounts/{accountID}/transfers"`) ||
		strings.Contains(body, "A-100") {
		t.Fatalf("metrics labels are wrong: %s", body)
	}
}

func TestObserveRequestsRejectsOversizedCallerRequestID(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	var got string
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got = RequestID(r.Context())
	})
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("X-Request-ID", strings.Repeat("x", 129))
	ObserveRequests(logger, next).ServeHTTP(httptest.NewRecorder(), request)
	if got == "" || len(got) > 128 {
		t.Fatalf("generated request id=%q", got)
	}
}

func TestRequestIDMissingFromPlainContext(t *testing.T) {
	if got := RequestID(context.Background()); got != "" {
		t.Fatalf("got %q", got)
	}
}
