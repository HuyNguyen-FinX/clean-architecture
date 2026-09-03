package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"example.com/cleanarch/lab06/solution/application"
)

type BalanceGetter interface {
	Execute(context.Context, string) (int64, error)
}

type Handler struct {
	get BalanceGetter
}

func New(get BalanceGetter) *Handler {
	if get == nil {
		panic("httpapi: nil balance getter")
	}
	return &Handler{get: get}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /accounts/{id}", h.getBalance)
	return mux
}

func (h *Handler) getBalance(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	balance, err := h.get.Execute(r.Context(), id)
	if errors.Is(err, application.ErrNotFound) {
		http.Error(w, "account not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, balance)
}
