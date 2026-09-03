package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example.com/cleanarch/lab11/solution/application"
	"example.com/cleanarch/lab11/solution/domain"
	"example.com/cleanarch/lab11/solution/httpapi"
	"example.com/cleanarch/lab11/solution/memory"
)

func TestRefactoredSlicePreservesBehaviorAndAddsAtomicOutbox(t *testing.T) {
	store := memory.New(domain.NewAccount("A", 100), domain.NewAccount("B", 0))
	handler := httpapi.New(application.NewTransferMoney(store))
	request := httptest.NewRequest(http.MethodPost, "/transfers",
		strings.NewReader(`{"from":"A","to":"B","amount":40}`))
	response := httptest.NewRecorder()
	handler.Transfer(response, request)
	accounts, outbox := store.Snapshot()
	if response.Code != http.StatusCreated || accounts["A"].Balance() != 60 ||
		accounts["B"].Balance() != 40 || len(outbox) != 1 {
		t.Fatalf("code=%d A=%d B=%d outbox=%v",
			response.Code, accounts["A"].Balance(), accounts["B"].Balance(), outbox)
	}
}

func TestRejectedTransferDoesNotWrite(t *testing.T) {
	store := memory.New(domain.NewAccount("A", 10), domain.NewAccount("B", 0))
	handler := httpapi.New(application.NewTransferMoney(store))
	request := httptest.NewRequest(http.MethodPost, "/transfers",
		strings.NewReader(`{"from":"A","to":"B","amount":11}`))
	response := httptest.NewRecorder()
	handler.Transfer(response, request)
	accounts, outbox := store.Snapshot()
	if response.Code != http.StatusConflict || accounts["A"].Balance() != 10 || len(outbox) != 0 {
		t.Fatalf("code=%d balance=%d outbox=%v", response.Code, accounts["A"].Balance(), outbox)
	}
}
