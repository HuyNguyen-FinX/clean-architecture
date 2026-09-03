package starter

import (
	"context"
	"encoding/json"
	"net/http"
)

type Transfer interface {
	Execute(context.Context, int64) error
}

type Handler struct{ transfer Transfer }

func New(transfer Transfer) *Handler { return &Handler{transfer: transfer} }

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Amount int64 `json:"amount"`
	}
	_ = json.NewDecoder(r.Body).Decode(&request)
	if err := h.transfer.Execute(context.Background(), request.Amount); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
