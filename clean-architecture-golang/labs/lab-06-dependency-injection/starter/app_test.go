package starter

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBalanceHandler(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/balance?id=A", nil)
	response := httptest.NewRecorder()
	BalanceHandler(response, request)
	if response.Code != http.StatusOK || response.Body.String() != "100" {
		t.Fatalf("code=%d body=%q", response.Code, response.Body.String())
	}
}
