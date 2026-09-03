package application

import (
	"context"
	"errors"
	"time"

	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/domain"
)

var (
	ErrIdempotencyConflict = errors.New("idempotency key conflicts with another request")
	ErrInvalidCommand      = errors.New("transfer command is invalid")
)

type AccountRepository interface {
	FindByID(ctx context.Context, id domain.AccountID) (*domain.Account, error)
	Save(ctx context.Context, account *domain.Account) error
}

type TransferRepository interface {
	SaveTransfer(ctx context.Context, transfer *domain.Transfer) error
	ListTransfersByAccount(
		ctx context.Context,
		accountID domain.AccountID,
		limit int,
	) ([]*domain.Transfer, error)
}

type IdempotencyRecord struct {
	Key         string
	RequestHash string
	TransferID  domain.TransferID
}

type IdempotencyRepository interface {
	ClaimIdempotency(ctx context.Context, key, requestHash string) (domain.TransferID, bool, error)
	CompleteIdempotency(ctx context.Context, key string, transferID domain.TransferID) error
}

type OutboxMessage struct {
	ID          string
	Topic       string
	Key         string
	Payload     []byte
	CreatedAt   time.Time
	PublishedAt *time.Time
}

type OutboxWriter interface {
	AddOutbox(ctx context.Context, message OutboxMessage) error
}

type TransferStore interface {
	AccountRepository
	TransferRepository
	IdempotencyRepository
	OutboxWriter
}

type OutboxStore interface {
	PendingOutbox(ctx context.Context, limit int) ([]OutboxMessage, error)
	MarkOutboxPublished(ctx context.Context, id string, publishedAt time.Time) error
}

type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
}

type EventPublisher interface {
	Publish(ctx context.Context, message OutboxMessage) error
}

type IDGenerator interface {
	NewID() string
}

type Clock interface {
	Now() time.Time
}
