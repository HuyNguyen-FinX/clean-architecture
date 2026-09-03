package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"example.com/cleanarch/lab06/solution/application"
)

var _ application.BalanceStore = (*Store)(nil)

type Store struct {
	mu       sync.RWMutex
	balances map[string]int64
	closed   bool
}

func New(seed map[string]int64) *Store {
	balances := make(map[string]int64, len(seed))
	for id, balance := range seed {
		balances[id] = balance
	}
	return &Store{balances: balances}
}

func (s *Store) Balance(ctx context.Context, id string) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return 0, errors.New("store closed")
	}
	balance, ok := s.balances[id]
	if !ok {
		return 0, fmt.Errorf("%q: %w", id, application.ErrNotFound)
	}
	return balance, nil
}

func (s *Store) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
}

func (s *Store) Closed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.closed
}
