package httpadapter

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/application"
	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/domain"
)

func TestTransferEndpointReturnsCreatedAndMapsCommand(t *testing.T) {
	transfer := &recordingTransferUseCase{
		result: application.TransferMoneyResult{TransferID: "T-100"},
	}
	handler := NewHandler(transfer, &recordingHistoryUseCase{})
	req := transferRequestFor(`{
		"from_account_id":"A-100",
		"to_account_id":"B-200",
		"amount":500000,
		"currency":"VND"
	}`, "KEY-1")
	rec := httptest.NewRecorder()

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated || !strings.Contains(rec.Body.String(), `"transfer_id":"T-100"`) {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
	if transfer.lastCommand.Amount != 500_000 || transfer.lastCommand.IdempotencyKey != "KEY-1" {
		t.Fatalf("handler did not map request: %+v", transfer.lastCommand)
	}
}

func TestTransferEndpointReturnsOKForIdempotentReplay(t *testing.T) {
	transfer := &recordingTransferUseCase{result: application.TransferMoneyResult{
		TransferID: "T-100", Replayed: true,
	}}
	rec := httptest.NewRecorder()
	NewHandler(transfer, &recordingHistoryUseCase{}).Routes().ServeHTTP(
		rec,
		transferRequestFor(
			`{"from_account_id":"A","to_account_id":"B","amount":1,"currency":"VND"}`,
			"KEY-1",
		),
	)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"replayed":true`) {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestTransferEndpointMapsStableErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code int
		body string
	}{
		{"insufficient", domain.ErrInsufficientBalance, http.StatusConflict, "insufficient_balance"},
		{"frozen", domain.ErrAccountFrozen, http.StatusConflict, "account_frozen"},
		{"idempotency", application.ErrIdempotencyConflict, http.StatusConflict, "idempotency_conflict"},
		{"missing key", application.ErrInvalidCommand, http.StatusBadRequest, "invalid_transfer"},
		{"internal", errors.New("password=secret database down"), http.StatusInternalServerError, "internal_error"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := NewHandler(&recordingTransferUseCase{err: test.err}, &recordingHistoryUseCase{})
			rec := httptest.NewRecorder()
			handler.Routes().ServeHTTP(rec, transferRequestFor(
				`{"from_account_id":"A","to_account_id":"B","amount":1,"currency":"VND"}`,
				"KEY-1",
			))
			if rec.Code != test.code || !strings.Contains(rec.Body.String(), test.body) {
				t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
			}
			if bytes.Contains(rec.Body.Bytes(), []byte("secret")) {
				t.Fatalf("response leaked internal error: %s", rec.Body.String())
			}
		})
	}
}

func TestTransferEndpointUsesStrictSingleJSONDocument(t *testing.T) {
	tests := []string{
		`{`,
		`{"from_account_id":"A","to_account_id":"B","amount":1,"currency":"VND"} {}`,
		`{"from_account_id":"A","to_account_id":"B","amount":1,"currency":"VND","extra":true}`,
	}
	for _, body := range tests {
		rec := httptest.NewRecorder()
		handler := NewHandler(&recordingTransferUseCase{}, &recordingHistoryUseCase{})
		handler.Routes().ServeHTTP(rec, transferRequestFor(body, "KEY-1"))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("body=%q code=%d", body, rec.Code)
		}
	}
}

func TestTransferHistoryMapsApplicationView(t *testing.T) {
	history := &recordingHistoryUseCase{items: []application.TransferView{{
		ID: "T-100", FromAccountID: "A-100", ToAccountID: "B-200",
		Amount: 500_000, Currency: "VND", CreatedAt: time.Unix(1_000, 0).UTC(),
	}}}
	handler := NewHandler(&recordingTransferUseCase{}, history)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/accounts/A-100/transfers?limit=25", nil)

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"transfer_id":"T-100"`) {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
	if history.accountID != "A-100" || history.limit != 25 {
		t.Fatalf("account=%q limit=%d", history.accountID, history.limit)
	}
}

func TestTransferHistoryRejectsInvalidLimit(t *testing.T) {
	handler := NewHandler(&recordingTransferUseCase{}, &recordingHistoryUseCase{})
	rec := httptest.NewRecorder()
	handler.Routes().ServeHTTP(
		rec,
		httptest.NewRequest(http.MethodGet, "/accounts/A/transfers?limit=101", nil),
	)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestNewHandlerRequiresUseCases(t *testing.T) {
	tests := []struct {
		transfer TransferUseCase
		history  TransferHistoryUseCase
	}{
		{nil, &recordingHistoryUseCase{}},
		{&recordingTransferUseCase{}, nil},
	}
	for _, test := range tests {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("expected panic")
				}
			}()
			NewHandler(test.transfer, test.history)
		}()
	}
}

func transferRequestFor(body, key string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/transfers", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", key)
	return req
}

type recordingTransferUseCase struct {
	err         error
	result      application.TransferMoneyResult
	lastCommand application.TransferMoneyCommand
}

func (uc *recordingTransferUseCase) Execute(
	_ context.Context,
	cmd application.TransferMoneyCommand,
) (application.TransferMoneyResult, error) {
	uc.lastCommand = cmd
	return uc.result, uc.err
}

type recordingHistoryUseCase struct {
	err       error
	items     []application.TransferView
	accountID string
	limit     int
}

func (uc *recordingHistoryUseCase) Execute(
	_ context.Context,
	accountID string,
	limit int,
) ([]application.TransferView, error) {
	uc.accountID = accountID
	uc.limit = limit
	return uc.items, uc.err
}
