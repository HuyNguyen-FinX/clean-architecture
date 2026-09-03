# Lab 10: Cache-Aside Decorator và Consistency

Thời lượng: 75-120 phút. Memory cache có fake clock làm executable model; challenge thay adapter Redis thật.

## Mục tiêu

- đặt cache sau read port;
- phân biệt hit/miss/error;
- dùng TTL + schema key;
- fail-open cache error;
- invalidate sau write;
- test stale window không dùng sleep.

## Kiến thức cần

- [Redis Cache](../../15-redis-cache/README.md)
- Repository decorator, Clock port, race testing.

## Diagram

~~~mermaid
flowchart LR
    APP["GetBalance"] --> CACHED["CachedBalanceReader"]
    CACHED --> CACHE["Cache port"]
    CACHED --> SOURCE["Primary source"]
    REDIS["Redis/Memory adapter"] -.implements.-> CACHE
~~~

## Problem

Starter map cache không TTL và không invalidation. Primary balance đổi nhưng read mãi trả giá trị cũ.

## Yêu cầu

1. Cache.Get trả value, found, error riêng.
2. Cache key có version.
3. Miss/error fallback primary.
4. Set cache sau primary success.
5. Fake Clock test expiry deterministic.
6. Returned values không share mutable pointers.
7. Concurrent use pass race.
8. Ghi rõ fail-open chỉ phù hợp performance cache.

## Các bước

1. Tái hiện stale starter.
2. Thiết kế BalanceSource và Cache ports.
3. Implement decorator.
4. Implement expiring memory adapter.
5. Inject Clock.
6. Test hit/miss/expiry/cache failure.
7. Thiết kế write invalidation.

## Expected behavior

Lần đầu đọc primary và populate; lần hai hit cache; trước TTL có thể stale theo contract; sau TTL reload; cache failure không làm read fail nếu primary khỏe.

## Test

~~~bash
cd starter && go test ./...
cd ../solution && go test -race ./... && go vet ./...
~~~

## Questions

1. TTL là product trade-off gì?
2. Vì sao Redis idempotency không nên fail-open?
3. Cache Aggregate balance khác read projection ra sao?
4. singleflight một process còn thiếu gì?
5. Invalidation race hình thành thế nào?

## Challenge

- Adapter go-redis + Testcontainers.
- TTL jitter.
- singleflight chống stampede.
- negative cache.
- version/fencing cho stale write.

## Solution explanation

Solution dùng read projection int64 để tránh cache mutable Aggregate. Fake clock làm expiry test tức thời. Memory adapter minh họa contract, không chứng minh Redis cluster/eviction/Lua semantics.
