package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"example.com/cleanarch/lab11/solution/application"
	"example.com/cleanarch/lab11/solution/domain"
)

type Transfer interface {
	Execute(context.Context, application.Command) error
}

type Handler struct{ transfer Transfer }

func New(transfer Transfer) *Handler { return &Handler{transfer: transfer} }

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
	err := h.transfer.Execute(r.Context(), application.Command(input))
	switch {
	case errors.Is(err, application.ErrNotFound):
		http.Error(w, "not found", http.StatusNotFound)
	case errors.Is(err, domain.ErrRejected):
		http.Error(w, "transfer rejected", http.StatusConflict)
	case err != nil:
		http.Error(w, "internal error", http.StatusInternalServerError)
	default:
		w.WriteHeader(http.StatusCreated)
	}
}
