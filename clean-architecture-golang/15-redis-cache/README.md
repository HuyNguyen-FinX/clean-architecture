# Redis và Cache Boundary: latency đổi lấy consistency complexity

Redis có thể làm cache, rate limiter, ephemeral coordination hoặc idempotency store. Cùng technology nhưng guarantee khác nhau; “Redis là infrastructure” chưa trả lời cache nằm ở đâu, fail-open hay fail-closed và stale data có chấp nhận không.

## Kết quả học tập

- chọn cache placement theo use case;
- implement cache-aside/decorator;
- thiết kế key/version/TTL/invalidation;
- xử lý stampede, stale data và cache failure;
- phân biệt distributed lock với DB transaction;
- test bằng fake clock và Redis integration.

## 1. Ba level

### Level 1

Cache giữ bản sao để đọc nhanh. Bản sao có thể cũ.

### Level 2

Engineer quản key, TTL, serialization, miss, invalidation, stampede, timeout và pool.

### Level 3

Cache thay consistency model. Placement quyết định layer nào biết stale/fallback semantics. Không có cache “trong suốt” nếu caller dựa vào freshness.

## 2. Bốn vị trí

### Repository decorator

~~~text
Application → CachedRepository → PostgresRepository
~~~

Hợp khi cache gần như thay thế read của repository và semantics đủ tương đương.

### Application service

Hợp khi workflow quyết định stale tolerance, bypass sau write hoặc fallback.

### Dedicated cache port

Hợp khi cache là capability rõ, ví dụ exchange-rate snapshot.

### HTTP/proxy cache

Hợp public GET theo ETag/Cache-Control, không cần domain biết.

Không có vị trí đúng cho mọi use case.

## 3. Cache-aside

~~~go
func (r *CachedRepository) FindByID(ctx context.Context, id AccountID) (*Account, error) {
	account, ok, err := r.cache.Get(ctx, id)
	if err == nil && ok {
		return account, nil
	}
	account, err = r.next.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	_ = r.cache.Set(ctx, account, r.ttl)
	return account, nil
}
~~~

Đoạn trên cố ý fail-open cho cache read/write. Với balance decision, stale Account có thể nguy hiểm; cached read projection phù hợp hơn write Aggregate.

## 4. Key design

~~~text
bank:v2:account-summary:{tenant}:{accountID}
~~~

Key cần namespace, schema version, tenant và canonical ID. Không dùng raw PII nếu key bị expose trong metrics/debug. Hash khi cần nhưng giữ khả năng điều tra có kiểm soát.

Version prefix giúp deploy schema mới không decode bytes cũ; phải plan cleanup memory.

## 5. TTL

TTL là product consistency decision:

- shorter: fresher, miss/load cao;
- longer: stale lâu, rẻ hơn;
- random jitter: tránh nhiều key expire đồng thời;
- no TTL: invalidation phải hoàn hảo.

Không copy TTL 5 phút cho mọi data. Account balance và country list có volatility/risk khác nhau.

## 6. Invalidation

Write-through:

~~~text
DB write → cache update
~~~

Nếu cache update fail sau DB commit, cache stale.

Delete-on-write:

~~~text
DB write → delete cache
~~~

Race reader có thể repopulate old value giữa DB read và delete. Double-delete/delay không là proof tuyệt đối.

Versioned/event invalidation có thể mạnh hơn nhưng thêm complexity. Chọn stale tolerance và recovery.

## 7. Stampede

Hot key expire, hàng nghìn request cùng query DB. Mitigation:

- singleflight trong process;
- distributed lease;
- TTL jitter;
- stale-while-revalidate;
- proactive refresh;
- request coalescing/load shedding.

singleflight một process không coalesce giữa replicas. Distributed lock cần timeout/token ownership.

## 8. Negative caching

Cache not-found ngắn hạn giảm penetration nhưng account vừa tạo có thể vẫn 404. Phải phân biệt cache miss với cached negative sentinel và chọn TTL nhỏ.

Không negative-cache transient DB error.

## 9. Serialization

Không gob raw domain struct nếu private fields/version evolution khó. Dùng cache DTO:

~~~go
type accountCacheV2 struct {
	ID           string
	BalanceMinor int64
	Currency     string
	Version      uint64
}
~~~

Decode rồi validate/rehydrate. Corrupt cache thường delete + fallback; metric để phát hiện deploy incompatibility.

## 10. Failure policy

| Redis use | Down thì? |
|---|---|
| performance cache | thường fail-open sang DB |
| rate limit security | có thể fail-closed hoặc local degraded |
| durable idempotency | không bypass nếu sẽ double charge |
| session/auth revocation | theo threat model |
| distributed coordination | operation có thể phải dừng |

Technology không quyết định criticality.

## 11. Distributed lock

Redis SET NX PX có thể tạo lease, nhưng:

- holder pause quá TTL;
- network partition;
- clock/timing;
- unlock nhầm owner nếu không token;
- failover semantics.

Fencing token ở protected resource có thể cần. Không dùng Redis lock thay DB row lock cho invariant nằm trong PostgreSQL nếu DB mới là source of truth.

## 12. Rate limiting

Token bucket/sliding window là application/delivery policy. Lua script giúp atomic multi-command. Key theo actor/IP/tenant; metric label tránh high cardinality.

429 response và Retry-After là HTTP adapter responsibility; limit policy có thể ở application nếu theo subscription/business tier.

## 13. Context và timeout

Cache phải nhanh hơn primary path. Per-attempt timeout ngắn, không retry nhiều làm cache chậm hơn DB. Truyền request context và phân biệt miss với timeout.

## 14. Testing

- decorator unit test hit/miss/fallback/invalidation;
- fake clock TTL;
- concurrent stampede test;
- serialization version/corrupt payload;
- Redis integration cho Lua/TTL/atomicity;
- load test hot key;
- chaos test Redis down.

Fake map không chứng minh expiry/cluster/failover.

## 15. Production scenario

Hot account summary key hết hạn cùng lúc trên 50 replicas:

- 5.000 requests xuyên DB;
- pool cạn;
- latency timeout;
- retries khuếch đại.

TTL jitter + distributed/singleflight coalescing + stale-while-revalidate và DB protection cần phối hợp. Cache hit ratio cao không đủ; đo origin load và stale/error.

## 16. Debug

1. hit/miss/error/latency theo cache name;
2. key version/TTL;
3. serialization failures;
4. origin query volume;
5. eviction/memory policy;
6. hot keys;
7. connection pool/timeouts;
8. invalidation timeline.

Không chạy KEYS trên production dataset lớn; dùng scan/metrics/tooling phù hợp.

## 17. Khi nào không cache?

Query đã nhanh, data ít, write-heavy, freshness nghiêm ngặt hoặc team chưa vận hành invalidation tốt: cache có thể giảm reliability. Tối ưu query/index trước. Cache là data system thứ hai.

## 18. Lab

Làm [Lab 10: Redis](../labs/lab-10-redis/README.md): cache-aside decorator, TTL fake clock, fail-open, invalidation và stampede challenge.

## 19. Mastery questions

1. Cache placement đổi ownership gì?
2. Vì sao balance Aggregate không nên cache mù?
3. TTL là business trade-off ra sao?
4. singleflight thiếu guarantee nào?
5. Redis down khi cache và idempotency khác gì?
6. Distributed lock khác DB transaction?
7. Negative cache gây stale 404 thế nào?
8. Hit ratio chưa đủ để đánh giá gì?

## Further reading

- Redis documentation về expiration, transactions, Lua và distributed locks.
- Cache-aside và cache stampede literature.
- Go x/sync/singleflight documentation.

## Quality gate

- [x] Placement và consistency model
- [x] Cache-aside code/key/TTL/invalidation
- [x] Stampede/negative/serialization/failure policy
- [x] Lock/rate-limit/context
- [x] Tests, production scenario, debug
- [x] Trade-off, lab, mastery
