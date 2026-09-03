package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"example.com/cleanarch/lab07/solution/domain"
)

type Transfer interface {
	Execute(context.Context, string, string, int64) error
}

type Handler struct{ transfer Transfer }

func New(transfer Transfer) *Handler { return &Handler{transfer: transfer} }

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var request struct {
		From   string `json:"from"`
		To     string `json:"to"`
		Amount int64  `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if err := h.transfer.Execute(r.Context(), request.From, request.To, request.Amount); err != nil {
		if errors.Is(err, domain.ErrInsufficient) {
			http.Error(w, "insufficient balance", http.StatusConflict)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
