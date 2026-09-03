package review

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

type TransferRequest struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int64  `json:"amount"`
}

type Account struct {
	ID      string
	Balance int64
	Frozen  bool
}

type Repository interface {
	Find(context.Context, string) (*Account, error)
	Save(context.Context, *Account) error
}

type Publisher interface {
	Publish(context.Context, TransferRequest) error
}

type DomainError struct {
	Status int
	Cause  error
}

func (e *DomainError) Error() string { return e.Cause.Error() }
func (e *DomainError) Unwrap() error { return e.Cause }

type Service struct {
	db        *sql.DB
	accounts  Repository
	publisher Publisher
}

func NewService(db *sql.DB, accounts Repository, publisher Publisher) *Service {
	return &Service{db: db, accounts: accounts, publisher: publisher}
}

func (s *Service) Transfer(ctx context.Context, request TransferRequest) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	from, err := s.accounts.Find(ctx, request.From)
	if err != nil {
		return err
	}
	to, err := s.accounts.Find(ctx, request.To)
	if err != nil {
		return err
	}
	if from.Frozen || from.Balance < request.Amount {
		return &DomainError{Status: http.StatusConflict, Cause: fmt.Errorf("transfer rejected")}
	}

	from.Balance -= request.Amount
	to.Balance += request.Amount
	if err := s.accounts.Save(ctx, from); err != nil {
		return err
	}
	if err := s.accounts.Save(ctx, to); err != nil {
		return err
	}
	go func() {
		_ = s.publisher.Publish(context.Background(), request)
	}()
	return tx.Commit()
}

type Handler struct{ service *Service }

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var request TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.service.Transfer(r.Context(), request); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
