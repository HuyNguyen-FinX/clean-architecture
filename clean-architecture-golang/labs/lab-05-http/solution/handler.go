package solution

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
)

var ErrInsufficientBalance = errors.New("insufficient balance")

type TransferCommand struct {
	From     string
	To       string
	Amount   int64
	Currency string
}

type Transfer interface {
	Execute(context.Context, TransferCommand) error
}

type Handler struct {
	transfer Transfer
}

func New(transfer Transfer) *Handler {
	if transfer == nil {
		panic("http: nil transfer use case")
	}
	return &Handler{transfer: transfer}
}

type transferRequest struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{"method_not_allowed", "method not allowed"})
		return
	}
	if !isJSON(r.Header.Get("Content-Type")) {
		writeJSON(w, http.StatusUnsupportedMediaType, errorResponse{"unsupported_media_type", "content type must be application/json"})
		return
	}
	var request transferRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{"invalid_request", "body must be one valid JSON object"})
		return
	}
	command := TransferCommand{
		From: request.From, To: request.To, Amount: request.Amount, Currency: request.Currency,
	}
	if err := h.transfer.Execute(r.Context(), command); err != nil {
		if errors.Is(err, ErrInsufficientBalance) {
			writeJSON(w, http.StatusConflict, errorResponse{"insufficient_balance", "balance is insufficient"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, errorResponse{"internal_error", "an internal error occurred"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}

func isJSON(value string) bool {
	mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(value))
	return err == nil && mediaType == "application/json"
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
