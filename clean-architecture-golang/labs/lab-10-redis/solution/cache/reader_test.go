package cache

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type fakeClock struct{ now time.Time }

func (c *fakeClock) Now() time.Time { return c.now }

type source struct {
	mu    sync.Mutex
	value int64
	calls int
}

func (s *source) Balance(context.Context, string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	return s.value, nil
}

type failingCache struct{}

func (failingCache) Get(context.Context, string) (int64, bool, error) {
	return 0, false, errors.New("redis down")
}
func (failingCache) Set(context.Context, string, int64, time.Duration) error {
	return errors.New("redis down")
}
func (failingCache) Delete(context.Context, string) error { return errors.New("redis down") }

func TestCachedReaderHitAndExpiry(t *testing.T) {
	clock := &fakeClock{now: time.Unix(1_000, 0)}
	primary := &source{value: 100}
	reader := NewCachedReader(primary, NewExpiringCache(clock), time.Minute)

	first, _ := reader.Balance(context.Background(), "A")
	primary.value = 200
	cached, _ := reader.Balance(context.Background(), "A")
	if first != 100 || cached != 100 || primary.calls != 1 {
		t.Fatalf("first=%d cached=%d calls=%d", first, cached, primary.calls)
	}

	clock.now = clock.now.Add(time.Minute)
	fresh, _ := reader.Balance(context.Background(), "A")
	if fresh != 200 || primary.calls != 2 {
		t.Fatalf("fresh=%d calls=%d", fresh, primary.calls)
	}
}

func TestCachedReaderFailsOpen(t *testing.T) {
	reader := NewCachedReader(&source{value: 100}, failingCache{}, time.Minute)
	got, err := reader.Balance(context.Background(), "A")
	if err != nil || got != 100 {
		t.Fatalf("got=%d err=%v", got, err)
	}
}

func TestInvalidateForcesReload(t *testing.T) {
	clock := &fakeClock{now: time.Unix(1_000, 0)}
	primary := &source{value: 100}
	reader := NewCachedReader(primary, NewExpiringCache(clock), time.Hour)
	_, _ = reader.Balance(context.Background(), "A")
	primary.value = 200
	if err := reader.Invalidate(context.Background(), "A"); err != nil {
		t.Fatal(err)
	}
	got, _ := reader.Balance(context.Background(), "A")
	if got != 200 {
		t.Fatalf("got=%d", got)
	}
}
