package memory

import (
	"context"
	"errors"
	"sync"

	"example.com/cleanarch/lab07/solution/application"
	"example.com/cleanarch/lab07/solution/domain"
)

var _ application.Repository = (*Repository)(nil)

type Repository struct {
	mu       sync.RWMutex
	accounts map[string]*domain.Account
}

func New(accounts ...*domain.Account) *Repository {
	repo := &Repository{accounts: make(map[string]*domain.Account)}
	for _, account := range accounts {
		repo.accounts[account.ID()] = account.Clone()
	}
	return repo
}

func (r *Repository) Find(ctx context.Context, id string) (*domain.Account, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	account, ok := r.accounts[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return account.Clone(), nil
}

func (r *Repository) Save(ctx context.Context, account *domain.Account) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.accounts[account.ID()] = account.Clone()
	return nil
}
