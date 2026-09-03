package cache

import (
	"context"
	"time"
)

type BalanceSource interface {
	Balance(context.Context, string) (int64, error)
}

type Store interface {
	Get(context.Context, string) (int64, bool, error)
	Set(context.Context, string, int64, time.Duration) error
	Delete(context.Context, string) error
}

type CachedReader struct {
	source BalanceSource
	cache  Store
	ttl    time.Duration
}

func NewCachedReader(source BalanceSource, cache Store, ttl time.Duration) *CachedReader {
	if source == nil || cache == nil || ttl <= 0 {
		panic("cache: invalid dependency or TTL")
	}
	return &CachedReader{source: source, cache: cache, ttl: ttl}
}

func (r *CachedReader) Balance(ctx context.Context, accountID string) (int64, error) {
	key := "bank:v1:balance:" + accountID
	if value, found, err := r.cache.Get(ctx, key); err == nil && found {
		return value, nil
	}
	value, err := r.source.Balance(ctx, accountID)
	if err != nil {
		return 0, err
	}
	_ = r.cache.Set(ctx, key, value, r.ttl)
	return value, nil
}

func (r *CachedReader) Invalidate(ctx context.Context, accountID string) error {
	return r.cache.Delete(ctx, "bank:v1:balance:"+accountID)
}
