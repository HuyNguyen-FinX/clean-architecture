package starter

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTransfer(t *testing.T) {
	Accounts = map[string]int64{"A": 1_000, "B": 100}
	History = nil
	req := httptest.NewRequest(http.MethodPost, "/transfers",
		strings.NewReader(`{"from":"A","to":"B","amount":300}`))
	rec := httptest.NewRecorder()
	Transfer(rec, req)
	if rec.Code != http.StatusCreated || Accounts["A"] != 700 || Accounts["B"] != 400 {
		t.Fatalf("code=%d accounts=%v", rec.Code, Accounts)
	}
}
