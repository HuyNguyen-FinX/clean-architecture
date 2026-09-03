package solution

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type contextKey struct{}

type fakeTransfer struct {
	command TransferCommand
	context context.Context
	err     error
}

func (f *fakeTransfer) Execute(ctx context.Context, command TransferCommand) error {
	f.context = ctx
	f.command = command
	return f.err
}

func request(body string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/transfers", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestHandlerMapsCommandAndContext(t *testing.T) {
	fake := &fakeTransfer{}
	req := request(`{"from":"A","to":"B","amount":100,"currency":"VND"}`)
	req = req.WithContext(context.WithValue(req.Context(), contextKey{}, "marker"))
	rec := httptest.NewRecorder()

	New(fake).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated || fake.command.Amount != 100 {
		t.Fatalf("code=%d command=%+v", rec.Code, fake.command)
	}
	if fake.context.Value(contextKey{}) != "marker" {
		t.Fatal("request context was not propagated")
	}
}

func TestHandlerRejectsInvalidBodies(t *testing.T) {
	bodies := []string{
		`{"from":"A","unknown":true}`,
		`{"from":"A"} {}`,
		`{`,
	}
	for _, body := range bodies {
		rec := httptest.NewRecorder()
		New(&fakeTransfer{}).ServeHTTP(rec, request(body))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("body=%q code=%d", body, rec.Code)
		}
	}
}

func TestHandlerMapsKnownAndHidesUnknownErrors(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		status  int
		notBody string
	}{
		{"known", ErrInsufficientBalance, http.StatusConflict, ""},
		{"unknown", errors.New("password=secret"), http.StatusInternalServerError, "secret"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			New(&fakeTransfer{err: test.err}).ServeHTTP(rec,
				request(`{"from":"A","to":"B","amount":100,"currency":"VND"}`))
			if rec.Code != test.status || strings.Contains(rec.Body.String(), test.notBody) && test.notBody != "" {
				t.Fatalf("code=%d body=%q", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestHandlerRequiresJSONContentType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/transfers", strings.NewReader("{}"))
	rec := httptest.NewRecorder()
	New(&fakeTransfer{}).ServeHTTP(rec, req)
	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("code=%d", rec.Code)
	}
}
