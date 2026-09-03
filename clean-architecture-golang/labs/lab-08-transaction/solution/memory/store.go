package memory

import (
	"context"
	"errors"
	"sync"

	"example.com/cleanarch/lab08/solution/application"
	"example.com/cleanarch/lab08/solution/domain"
)

var (
	_                    application.AccountRepository = (*Store)(nil)
	_                    application.Transactor        = (*Store)(nil)
	ErrNotFound                                        = errors.New("account not found")
	ErrWrongTransaction                                = errors.New("transaction belongs to another store")
	ErrNestedTransaction                               = errors.New("nested transaction is not supported")
)

type txKey struct{}

type transaction struct {
	owner    *Store
	accounts map[string]*domain.Account
}

type Store struct {
	mu       sync.RWMutex
	accounts map[string]*domain.Account
}

func New(accounts ...*domain.Account) *Store {
	store := &Store{accounts: make(map[string]*domain.Account, len(accounts))}
	for _, account := range accounts {
		store.accounts[account.ID()] = account.Clone()
	}
	return store
}

func (s *Store) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	if ctx.Value(txKey{}) != nil {
		return ErrNestedTransaction
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	staged := cloneMap(s.accounts)
	tx := &transaction{owner: s, accounts: staged}
	if err := fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		return err
	}
	s.accounts = staged
	return nil
}

func (s *Store) FindByID(ctx context.Context, id string) (*domain.Account, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if tx, ok := ctx.Value(txKey{}).(*transaction); ok {
		if tx.owner != s {
			return nil, ErrWrongTransaction
		}
		return find(tx.accounts, id)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return find(s.accounts, id)
}

func (s *Store) Save(ctx context.Context, account *domain.Account) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if tx, ok := ctx.Value(txKey{}).(*transaction); ok {
		if tx.owner != s {
			return ErrWrongTransaction
		}
		tx.accounts[account.ID()] = account.Clone()
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accounts[account.ID()] = account.Clone()
	return nil
}

func (s *Store) Snapshot() map[string]*domain.Account {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneMap(s.accounts)
}

func find(accounts map[string]*domain.Account, id string) (*domain.Account, error) {
	account, ok := accounts[id]
	if !ok {
		return nil, ErrNotFound
	}
	return account.Clone(), nil
}

func cloneMap(source map[string]*domain.Account) map[string]*domain.Account {
	result := make(map[string]*domain.Account, len(source))
	for id, account := range source {
		result[id] = account.Clone()
	}
	return result
}
