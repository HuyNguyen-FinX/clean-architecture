package application

import (
	"context"

	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/domain"
)

type AccountRepository interface {
	FindByID(ctx context.Context, id domain.AccountID) (*domain.Account, error)
	Save(ctx context.Context, account *domain.Account) error
}

type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
}

type NoopTransactor struct{}

func (NoopTransactor) WithinTransaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	return fn(ctx)
}
