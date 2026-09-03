package application

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("account not found")

type BalanceStore interface {
	Balance(context.Context, string) (int64, error)
}

type GetBalance struct {
	store BalanceStore
}

func NewGetBalance(store BalanceStore) *GetBalance {
	if store == nil {
		panic("application: nil balance store")
	}
	return &GetBalance{store: store}
}

func (uc *GetBalance) Execute(ctx context.Context, id string) (int64, error) {
	if id == "" {
		return 0, ErrNotFound
	}
	return uc.store.Balance(ctx, id)
}
