package memory

import (
	"context"
	"sync"
)

type Inbox struct {
	mu        sync.Mutex
	completed map[string]struct{}
}

func NewInbox() *Inbox {
	return &Inbox{completed: make(map[string]struct{})}
}

func (i *Inbox) ProcessOnce(
	ctx context.Context,
	eventID string,
	fn func(context.Context) error,
) (bool, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if _, exists := i.completed[eventID]; exists {
		return false, nil
	}
	if err := fn(ctx); err != nil {
		return false, err
	}
	i.completed[eventID] = struct{}{}
	return true, nil
}
