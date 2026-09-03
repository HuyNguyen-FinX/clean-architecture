package application

import (
	"context"
	"errors"
	"time"

	"example.com/cleanarch/lab12/solution/domain"
)

var (
	ErrNotFound            = errors.New("account not found")
	ErrInvalidCommand      = errors.New("invalid transfer command")
	ErrIdempotencyConflict = errors.New("idempotency key conflicts with another request")
)

type TransferRecord struct {
	ID        string
	From      string
	To        string
	Amount    int64
	CreatedAt time.Time
}

type OutboxEvent struct {
	ID        string
	Type      string
	Key       string
	Payload   []byte
	CreatedAt time.Time
}

type IdempotencyRecord struct {
	Key         string
	RequestHash string
	TransferID  string
}

type Transaction interface {
	FindAccount(string) (*domain.Account, error)
	SaveAccount(*domain.Account) error
	FindIdempotency(string) (IdempotencyRecord, bool)
	SaveIdempotency(IdempotencyRecord) error
	AppendTransfer(TransferRecord) error
	AppendOutbox(OutboxEvent) error
}

type UnitOfWork interface {
	WithinTransaction(context.Context, func(Transaction) error) error
}

type HistoryReader interface {
	ByAccount(context.Context, string) ([]TransferRecord, error)
}

type IDGenerator interface {
	NewID() string
}

type Clock interface {
	Now() time.Time
}
