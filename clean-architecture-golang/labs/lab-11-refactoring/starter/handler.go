package starter

import (
	"encoding/json"
	"net/http"
)

type Account struct {
	ID      string
	Balance int64
	Frozen  bool
}

type Store struct{ Accounts map[string]*Account }

type Producer struct{ Events []string }

type Handler struct {
	Store    *Store
	Producer *Producer
}

func (h *Handler) Transfer(w http.ResponseWriter, r *http.Request) {
	var input struct {
		From   string `json:"from"`
		To     string `json:"to"`
		Amount int64  `json:"amount"`
	}
	if json.NewDecoder(r.Body).Decode(&input) != nil || input.Amount <= 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	from, to := h.Store.Accounts[input.From], h.Store.Accounts[input.To]
	if from == nil || to == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if from.Frozen || from.Balance < input.Amount {
		http.Error(w, "transfer rejected", http.StatusConflict)
		return
	}
	from.Balance -= input.Amount
	to.Balance += input.Amount
	h.Producer.Events = append(h.Producer.Events, input.From+"->"+input.To)
	w.WriteHeader(http.StatusCreated)
}
