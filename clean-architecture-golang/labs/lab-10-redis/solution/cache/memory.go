package cache

import (
	"context"
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
}

type entry struct {
	value     int64
	expiresAt time.Time
}

type ExpiringCache struct {
	mu      sync.Mutex
	clock   Clock
	entries map[string]entry
}

func NewExpiringCache(clock Clock) *ExpiringCache {
	return &ExpiringCache{clock: clock, entries: make(map[string]entry)}
}

func (c *ExpiringCache) Get(ctx context.Context, key string) (int64, bool, error) {
	if err := ctx.Err(); err != nil {
		return 0, false, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	item, ok := c.entries[key]
	if !ok {
		return 0, false, nil
	}
	if !c.clock.Now().Before(item.expiresAt) {
		delete(c.entries, key)
		return 0, false, nil
	}
	return item.value, true, nil
}

func (c *ExpiringCache) Set(ctx context.Context, key string, value int64, ttl time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = entry{value: value, expiresAt: c.clock.Now().Add(ttl)}
	return nil
}

func (c *ExpiringCache) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
	return nil
}
