package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type spyTransfer struct {
	from, to string
	amount   int64
}

func (s *spyTransfer) Execute(_ context.Context, from, to string, amount int64) error {
	s.from, s.to, s.amount = from, to, amount
	return nil
}

func TestHandlerMapsRequest(t *testing.T) {
	spy := &spyTransfer{}
	request := httptest.NewRequest(http.MethodPost, "/transfers",
		strings.NewReader(`{"from":"A","to":"B","amount":40}`))
	response := httptest.NewRecorder()
	New(spy).ServeHTTP(response, request)
	if response.Code != http.StatusCreated || spy.from != "A" || spy.to != "B" || spy.amount != 40 {
		t.Fatalf("code=%d spy=%+v", response.Code, spy)
	}
}
