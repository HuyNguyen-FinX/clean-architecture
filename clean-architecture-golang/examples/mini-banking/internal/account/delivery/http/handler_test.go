package httpadapter

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/application"
	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/domain"
)

func TestTransferEndpointReturnsCreated(t *testing.T) {
	useCase := &recordingTransferUseCase{}
	handler := NewHandler(useCase)

	req := httptest.NewRequest(http.MethodPost, "/transfers", bytes.NewBufferString(`{
		"from_account_id": "A-100",
		"to_account_id": "B-200",
		"amount": 500000,
		"currency": "VND"
	}`))
	rec := httptest.NewRecorder()

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	if useCase.lastCommand.Amount != 500_000 {
		t.Fatalf("handler did not map request to command")
	}
}

func TestTransferEndpointMapsDomainError(t *testing.T) {
	useCase := &recordingTransferUseCase{err: domain.ErrInsufficientBalance}
	handler := NewHandler(useCase)

	req := httptest.NewRequest(http.MethodPost, "/transfers", bytes.NewBufferString(`{
		"from_account_id": "A-100",
		"to_account_id": "B-200",
		"amount": 500000,
		"currency": "VND"
	}`))
	rec := httptest.NewRecorder()

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

func TestTransferEndpointRejectsInvalidJSON(t *testing.T) {
	handler := NewHandler(&recordingTransferUseCase{})

	req := httptest.NewRequest(http.MethodPost, "/transfers", bytes.NewBufferString(`{`))
	rec := httptest.NewRecorder()

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

type recordingTransferUseCase struct {
	err         error
	lastCommand application.TransferMoneyCommand
}

func (uc *recordingTransferUseCase) Execute(_ context.Context, cmd application.TransferMoneyCommand) error {
	uc.lastCommand = cmd
	if uc.err != nil {
		return uc.err
	}

	if cmd.Amount == 0 {
		return errors.New("test command was not mapped")
	}

	return nil
}
