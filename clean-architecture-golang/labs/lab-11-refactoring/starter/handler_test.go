package starter

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTransferCharacterization(t *testing.T) {
	store := &Store{Accounts: map[string]*Account{
		"A": {ID: "A", Balance: 100}, "B": {ID: "B", Balance: 0},
	}}
	producer := &Producer{}
	handler := &Handler{Store: store, Producer: producer}
	request := httptest.NewRequest(http.MethodPost, "/transfers",
		strings.NewReader(`{"from":"A","to":"B","amount":40}`))
	response := httptest.NewRecorder()
	handler.Transfer(response, request)
	if response.Code != http.StatusCreated || store.Accounts["A"].Balance != 60 ||
		store.Accounts["B"].Balance != 40 || len(producer.Events) != 1 {
		t.Fatalf("code=%d accounts=%v events=%v", response.Code, store.Accounts, producer.Events)
	}
}
