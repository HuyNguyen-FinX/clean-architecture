package memory

import (
	"context"
	"fmt"
	"sync"

	"example.com/cleanarch/lab03/solution/application"
	"example.com/cleanarch/lab03/solution/domain"
)

var _ application.AccountRepository = (*Repository)(nil)

type Repository struct {
	mu       sync.RWMutex
	accounts map[domain.AccountID]*domain.Account
}

func New(accounts ...*domain.Account) (*Repository, error) {
	repo := &Repository{accounts: make(map[domain.AccountID]*domain.Account, len(accounts))}
	for _, account := range accounts {
		if account == nil {
			return nil, fmt.Errorf("seed account: %w", domain.ErrInvalidAccount)
		}
		copied, err := clone(account)
		if err != nil {
			return nil, fmt.Errorf("seed account %q: %w", account.ID(), err)
		}
		repo.accounts[copied.ID()] = copied
	}
	return repo, nil
}

func (r *Repository) FindByID(ctx context.Context, id domain.AccountID) (*domain.Account, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	account, ok := r.accounts[id]
	if !ok {
		return nil, fmt.Errorf("%q: %w", id, application.ErrAccountNotFound)
	}
	copied, err := clone(account)
	if err != nil {
		return nil, fmt.Errorf("clone account %q: %w", id, err)
	}
	return copied, nil
}

func (r *Repository) Save(ctx context.Context, account *domain.Account) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if account == nil {
		return domain.ErrInvalidAccount
	}
	copied, err := clone(account)
	if err != nil {
		return fmt.Errorf("clone account %q: %w", account.ID(), err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.accounts[copied.ID()] = copied
	return nil
}

func clone(account *domain.Account) (*domain.Account, error) {
	return domain.RehydrateAccount(account.ID(), account.Balance())
}
