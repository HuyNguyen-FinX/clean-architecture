package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"example.com/cleanarch/lab12/solution/application"
	"example.com/cleanarch/lab12/solution/domain"
)

type Transfer interface {
	Execute(context.Context, application.TransferCommand) (application.TransferResult, error)
}

type Handler struct {
	transfer Transfer
	history  application.HistoryReader
}

func New(transfer Transfer, history application.HistoryReader) *Handler {
	return &Handler{transfer: transfer, history: history}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /transfers", h.createTransfer)
	mux.HandleFunc("GET /accounts/{id}/transfers", h.listTransfers)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	return mux
}

func (h *Handler) createTransfer(w http.ResponseWriter, r *http.Request) {
	var request struct {
		From   string `json:"from"`
		To     string `json:"to"`
		Amount int64  `json:"amount"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{"invalid_request", "invalid JSON body"})
		return
	}
	result, err := h.transfer.Execute(r.Context(), application.TransferCommand{
		IdempotencyKey: r.Header.Get("Idempotency-Key"),
		From:           request.From, To: request.To, Amount: request.Amount,
	})
	if err != nil {
		status, body := mapError(err)
		writeJSON(w, status, body)
		return
	}
	status := http.StatusCreated
	if result.Replayed {
		status = http.StatusOK
	}
	writeJSON(w, status, map[string]any{"transfer_id": result.ID, "replayed": result.Replayed})
}

func (h *Handler) listTransfers(w http.ResponseWriter, r *http.Request) {
	records, err := h.history.ByAccount(r.Context(), r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorBody{"internal_error", "an internal error occurred"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": records})
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func mapError(err error) (int, errorBody) {
	switch {
	case errors.Is(err, application.ErrInvalidCommand), errors.Is(err, domain.ErrInvalidAmount):
		return http.StatusBadRequest, errorBody{"invalid_transfer", "transfer request is invalid"}
	case errors.Is(err, application.ErrNotFound):
		return http.StatusNotFound, errorBody{"account_not_found", "account not found"}
	case errors.Is(err, domain.ErrInsufficient):
		return http.StatusConflict, errorBody{"insufficient_balance", "balance is insufficient"}
	case errors.Is(err, application.ErrIdempotencyConflict):
		return http.StatusConflict, errorBody{"idempotency_conflict", "key was used for another request"}
	default:
		return http.StatusInternalServerError, errorBody{"internal_error", "an internal error occurred"}
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
