package memory

import (
	"context"
	"sync"

	"example.com/cleanarch/lab11/solution/application"
	"example.com/cleanarch/lab11/solution/domain"
)

type Store struct {
	mu       sync.Mutex
	accounts map[string]*domain.Account
	outbox   []string
}

func New(accounts ...*domain.Account) *Store {
	store := &Store{accounts: make(map[string]*domain.Account)}
	for _, account := range accounts {
		store.accounts[account.ID()] = account.Clone()
	}
	return store
}

func (s *Store) Within(ctx context.Context, fn func(application.Transaction) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tx := &transaction{accounts: clone(s.accounts), outbox: append([]string(nil), s.outbox...)}
	if err := fn(tx); err != nil {
		return err
	}
	s.accounts, s.outbox = tx.accounts, tx.outbox
	return nil
}

func (s *Store) Snapshot() (map[string]*domain.Account, []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return clone(s.accounts), append([]string(nil), s.outbox...)
}

type transaction struct {
	accounts map[string]*domain.Account
	outbox   []string
}

func (tx *transaction) Find(id string) (*domain.Account, error) {
	account, ok := tx.accounts[id]
	if !ok {
		return nil, application.ErrNotFound
	}
	return account.Clone(), nil
}
func (tx *transaction) Save(account *domain.Account) error {
	tx.accounts[account.ID()] = account.Clone()
	return nil
}
func (tx *transaction) AppendOutbox(event string) error {
	tx.outbox = append(tx.outbox, event)
	return nil
}

func clone(source map[string]*domain.Account) map[string]*domain.Account {
	target := make(map[string]*domain.Account, len(source))
	for id, account := range source {
		target[id] = account.Clone()
	}
	return target
}
