package memory

import (
	"context"
	"sync"

	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/domain"
)

type Repository struct {
	mu       sync.RWMutex
	accounts map[domain.AccountID]*domain.Account
}

func NewRepository(accounts ...*domain.Account) *Repository {
	repo := &Repository{
		accounts: make(map[domain.AccountID]*domain.Account, len(accounts)),
	}

	for _, account := range accounts {
		repo.accounts[account.ID()] = account.Clone()
	}

	return repo
}

func (r *Repository) FindByID(ctx context.Context, id domain.AccountID) (*domain.Account, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	account, ok := r.accounts[id]
	if !ok {
		return nil, domain.ErrAccountNotFound
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

func (r *Repository) Snapshot() map[domain.AccountID]*domain.Account {
	r.mu.RLock()
	defer r.mu.RUnlock()

	snapshot := make(map[domain.AccountID]*domain.Account, len(r.accounts))
	for id, account := range r.accounts {
		snapshot[id] = account.Clone()
	}

	return snapshot
}
