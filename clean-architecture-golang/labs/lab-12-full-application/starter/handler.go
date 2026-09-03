package starter

import (
	"encoding/json"
	"net/http"
)

var Accounts = map[string]int64{"A": 1_000, "B": 100}
var History []string

func Transfer(w http.ResponseWriter, r *http.Request) {
	var input struct {
		From   string `json:"from"`
		To     string `json:"to"`
		Amount int64  `json:"amount"`
	}
	if json.NewDecoder(r.Body).Decode(&input) != nil || input.Amount <= 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if Accounts[input.From] < input.Amount {
		http.Error(w, "insufficient", http.StatusConflict)
		return
	}
	Accounts[input.From] -= input.Amount
	Accounts[input.To] += input.Amount
	History = append(History, input.From+"->"+input.To)
	w.WriteHeader(http.StatusCreated)
}
