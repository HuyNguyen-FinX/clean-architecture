package httpadapter

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/application"
	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/domain"
)

type TransferUseCase interface {
	Execute(context.Context, application.TransferMoneyCommand) (application.TransferMoneyResult, error)
}

type TransferHistoryUseCase interface {
	Execute(context.Context, string, int) ([]application.TransferView, error)
}

type Handler struct {
	transfer TransferUseCase
	history  TransferHistoryUseCase
}

func NewHandler(transfer TransferUseCase, history TransferHistoryUseCase) *Handler {
	if transfer == nil || history == nil {
		panic("http adapter: nil use case")
	}
	return &Handler{transfer: transfer, history: history}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.handleHealth)
	mux.HandleFunc("POST /transfers", h.handleTransfer)
	mux.HandleFunc("GET /accounts/{accountID}/transfers", h.handleTransferHistory)
	return mux
}

func (h *Handler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleTransfer(w http.ResponseWriter, r *http.Request) {
	var req transferRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Code: "invalid_request", Message: "request body must be one valid JSON object",
		})
		return
	}

	result, err := h.transfer.Execute(r.Context(), application.TransferMoneyCommand{
		IdempotencyKey: r.Header.Get("Idempotency-Key"),
		FromAccountID:  req.FromAccountID,
		ToAccountID:    req.ToAccountID,
		Amount:         req.Amount,
		Currency:       req.Currency,
	})
	if err != nil {
		status, response := responseFromError(err)
		writeJSON(w, status, response)
		return
	}
	status := http.StatusCreated
	if result.Replayed {
		status = http.StatusOK
	}
	writeJSON(w, status, transferResponse{
		TransferID: string(result.TransferID),
		Replayed:   result.Replayed,
	})
}

func (h *Handler) handleTransferHistory(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 || parsed > 100 {
			writeJSON(w, http.StatusBadRequest, errorResponse{
				Code: "invalid_limit", Message: "limit must be between 1 and 100",
			})
			return
		}
		limit = parsed
	}
	items, err := h.history.Execute(r.Context(), r.PathValue("accountID"), limit)
	if err != nil {
		status, response := responseFromError(err)
		writeJSON(w, status, response)
		return
	}
	response := make([]transferHistoryItem, 0, len(items))
	for _, item := range items {
		response = append(response, transferHistoryItem{
			TransferID: item.ID, FromAccountID: item.FromAccountID,
			ToAccountID: item.ToAccountID, Amount: item.Amount,
			Currency: item.Currency, CreatedAt: item.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": response})
}

type transferRequest struct {
	FromAccountID string `json:"from_account_id"`
	ToAccountID   string `json:"to_account_id"`
	Amount        int64  `json:"amount"`
	Currency      string `json:"currency"`
}

type transferResponse struct {
	TransferID string `json:"transfer_id"`
	Replayed   bool   `json:"replayed"`
}

type transferHistoryItem struct {
	TransferID    string    `json:"transfer_id"`
	FromAccountID string    `json:"from_account_id"`
	ToAccountID   string    `json:"to_account_id"`
	Amount        int64     `json:"amount"`
	Currency      string    `json:"currency"`
	CreatedAt     time.Time `json:"created_at"`
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func responseFromError(err error) (int, errorResponse) {
	switch {
	case errors.Is(err, domain.ErrAccountNotFound):
		return http.StatusNotFound, errorResponse{"account_not_found", "account was not found"}
	case errors.Is(err, domain.ErrInsufficientBalance):
		return http.StatusConflict, errorResponse{"insufficient_balance", "account balance is insufficient"}
	case errors.Is(err, domain.ErrAccountFrozen):
		return http.StatusConflict, errorResponse{"account_frozen", "source account is frozen"}
	case errors.Is(err, application.ErrIdempotencyConflict):
		return http.StatusConflict, errorResponse{"idempotency_conflict", "idempotency key was used for another request"}
	case errors.Is(err, application.ErrInvalidCommand),
		errors.Is(err, domain.ErrInvalidAccountID),
		errors.Is(err, domain.ErrInvalidAmount),
		errors.Is(err, domain.ErrInvalidCurrency),
		errors.Is(err, domain.ErrCurrencyMismatch),
		errors.Is(err, domain.ErrSameAccountTransfer):
		return http.StatusBadRequest, errorResponse{"invalid_transfer", "transfer request is invalid"}
	default:
		return http.StatusInternalServerError, errorResponse{"internal_error", "an internal error occurred"}
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	contentType := strings.TrimSpace(r.Header.Get("Content-Type"))
	if contentType != "" {
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil || mediaType != "application/json" {
			return errors.New("content type must be application/json")
		}
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("request body contains multiple JSON values")
		}
		return err
	}
	return nil
}
