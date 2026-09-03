package memory

import (
	"context"
	"sync"

	"example.com/cleanarch/lab12/solution/application"
	"example.com/cleanarch/lab12/solution/domain"
)

var (
	_ application.UnitOfWork    = (*Store)(nil)
	_ application.HistoryReader = (*Store)(nil)
)

type state struct {
	accounts    map[string]*domain.Account
	transfers   []application.TransferRecord
	outbox      []application.OutboxEvent
	idempotency map[string]application.IdempotencyRecord
}

type Store struct {
	mu    sync.RWMutex
	state state
}

func New(accounts ...*domain.Account) *Store {
	s := &Store{state: state{
		accounts: make(map[string]*domain.Account), idempotency: make(map[string]application.IdempotencyRecord),
	}}
	for _, account := range accounts {
		s.state.accounts[account.ID()] = account.Clone()
	}
	return s
}

func (s *Store) WithinTransaction(ctx context.Context, fn func(application.Transaction) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	staged := cloneState(s.state)
	if err := fn((*transaction)(&staged)); err != nil {
		return err
	}
	s.state = staged
	return nil
}

func (s *Store) ByAccount(ctx context.Context, id string) ([]application.TransferRecord, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]application.TransferRecord, 0)
	for _, record := range s.state.transfers {
		if record.From == id || record.To == id {
			result = append(result, record)
		}
	}
	return result, nil
}

func (s *Store) Balance(id string) (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	account, ok := s.state.accounts[id]
	if !ok {
		return 0, false
	}
	return account.Balance(), true
}

func (s *Store) Outbox() []application.OutboxEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]application.OutboxEvent(nil), s.state.outbox...)
}

type transaction state

func (tx *transaction) FindAccount(id string) (*domain.Account, error) {
	account, ok := tx.accounts[id]
	if !ok {
		return nil, application.ErrNotFound
	}
	return account.Clone(), nil
}

func (tx *transaction) SaveAccount(account *domain.Account) error {
	tx.accounts[account.ID()] = account.Clone()
	return nil
}

func (tx *transaction) FindIdempotency(key string) (application.IdempotencyRecord, bool) {
	record, ok := tx.idempotency[key]
	return record, ok
}

func (tx *transaction) SaveIdempotency(record application.IdempotencyRecord) error {
	tx.idempotency[record.Key] = record
	return nil
}

func (tx *transaction) AppendTransfer(record application.TransferRecord) error {
	tx.transfers = append(tx.transfers, record)
	return nil
}

func (tx *transaction) AppendOutbox(event application.OutboxEvent) error {
	event.Payload = append([]byte(nil), event.Payload...)
	tx.outbox = append(tx.outbox, event)
	return nil
}

func cloneState(source state) state {
	target := state{
		accounts:    make(map[string]*domain.Account, len(source.accounts)),
		transfers:   append([]application.TransferRecord(nil), source.transfers...),
		outbox:      append([]application.OutboxEvent(nil), source.outbox...),
		idempotency: make(map[string]application.IdempotencyRecord, len(source.idempotency)),
	}
	for id, account := range source.accounts {
		target.accounts[id] = account.Clone()
	}
	for key, record := range source.idempotency {
		target.idempotency[key] = record
	}
	for i := range target.outbox {
		target.outbox[i].Payload = append([]byte(nil), target.outbox[i].Payload...)
	}
	return target
}
