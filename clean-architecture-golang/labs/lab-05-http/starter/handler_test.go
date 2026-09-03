package starter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeTransfer struct{ amount int64 }

func (f *fakeTransfer) Execute(_ context.Context, amount int64) error {
	f.amount = amount
	return nil
}

func TestHandlerHappyPath(t *testing.T) {
	fake := &fakeTransfer{}
	request := httptest.NewRequest(http.MethodPost, "/transfers", strings.NewReader(`{"amount":100}`))
	response := httptest.NewRecorder()
	New(fake).ServeHTTP(response, request)
	if response.Code != http.StatusCreated || fake.amount != 100 {
		t.Fatalf("code=%d amount=%d", response.Code, fake.amount)
	}
}
