# Infrastructure Layer: detail production sau inward-facing ports

Infrastructure chứa PostgreSQL, Redis, Kafka publisher, HTTP clients, clocks, file systems và vendor SDK. Gọi chúng là “detail” không có nghĩa code adapter được phép sơ sài; production reliability thường sống chính ở đây.

## Kết quả học tập

- implement port mà không làm driver leak vào core;
- phân biệt driving/driven adapter;
- thiết kế mapping, timeout, retry, pool và health;
- giữ adapter semantics đồng nhất bằng contract test;
- đặt observability tại I/O boundary;
- tránh abstraction giả và vendor-agnostic fantasy.

## 1. Problem

~~~go
type TransferMoney struct {
	pool   *pgxpool.Pool
	redis  *redis.Client
	writer *kafka.Writer
}
~~~

Application bị buộc vào protocol/API của ba vendor. Unit test orchestration phải mock concrete clients; driver upgrade kéo core thay đổi; business vocabulary biến thành Exec/Get/WriteMessages.

Port diễn đạt capability:

~~~go
type AccountRepository interface {
	FindByID(context.Context, domain.AccountID) (*domain.Account, error)
	Save(context.Context, *domain.Account) error
}

type EventPublisher interface {
	Publish(context.Context, application.Event) error
}
~~~

Adapter dịch capability sang driver.

## 2. Ba level

### Level 1: trực giác

Core nói “tôi cần lưu Account”; adapter biết phải chạy SQL nào.

### Level 2: Backend Engineer

Infrastructure sở hữu:

- connection/client lifecycle;
- serialization/mapping;
- timeout/retry/backoff;
- error classification;
- protocol-specific consistency;
- health/metrics/tracing;
- integration tests.

### Level 3: Architecture

Port ổn định theo consumer intent; adapter hấp thụ volatility của provider. Nhưng abstraction không thể xóa semantic differences: Redis cache không giống PostgreSQL, Kafka at-least-once không giống function call. Contract phải công bố guarantee thay vì giả vờ implementations tương đương.

## 3. Compile-time và runtime

~~~mermaid
flowchart LR
    APP["Application"] --> PORT["Port"]
    PG["Postgres adapter"] -.implements.-> PORT
    REDIS["Redis adapter"] -.implements.-> PORT
    EXT["Payment HTTP adapter"] -.implements.-> PORT
~~~

Adapter import application/domain. Runtime application gọi injected adapter qua port. Composition root import cả hai.

## 4. Driving vs driven

- Driving: HTTP handler, gRPC server, Kafka consumer, cron.
- Driven: PostgreSQL Repository, Redis cache, Kafka publisher, external gateway.

Một technology có thể ở hai phía. Kafka consumer kích use case; Kafka producer được use case gọi. Gộp tất cả vào infrastructure folder vẫn được nếu team hiểu direction.

## 5. Adapter phải map model

Postgres row, Redis bytes, vendor response và domain object có schema/lifecycle khác nhau. Mapping boundary:

~~~text
provider data
→ validate/normalize
→ adapter model
→ domain/application model
~~~

Không tạo invalid Entity bằng struct literal vì field private. Rehydrate qua factory và trả corrupt-data error.

Mapping error không phải business rejection. Nó là signal dữ liệu/provider không thỏa contract, cần telemetry và thường không retry mù.

## 6. Error taxonomy

Adapter nên giữ cause:

~~~go
if errors.Is(err, pgx.ErrNoRows) {
	return nil, application.ErrAccountNotFound
}
return nil, fmt.Errorf("find account %q: %w", id, err)
~~~

Chỉ map khi consumer cần semantics. Đừng tạo một enum “infrastructure error” mất toàn bộ cause. HTTP response sanitization xảy ra ở delivery; retry classification có thể dùng typed wrapper/SQLSTATE.

## 7. Timeout và budget

Context deadline từ caller là budget toàn operation. Adapter có thể đặt per-attempt timeout ngắn hơn:

~~~go
attemptCtx, cancel := context.WithTimeout(ctx, 800*time.Millisecond)
defer cancel()
response, err := client.Do(request.WithContext(attemptCtx))
~~~

Không tạo timeout dài hơn parent. Tổng retries + backoff phải nằm trong budget. Database query, Redis và external HTTP có latency/failure semantics khác; policy không nhất thiết giống.

## 8. Retry ở đúng nơi

Retry adapter khi:

- operation idempotent hoặc có idempotency key;
- error transient được phân loại;
- caller budget cho phép;
- có max attempt/backoff/jitter/metrics.

Không retry:

- validation/business rejection;
- non-idempotent call không key;
- context canceled;
- mọi 5xx vô hạn;
- bên trong và bên ngoài cùng retry nhân bùng số attempt.

Application có thể sở hữu workflow retry; adapter sở hữu transport attempt. Tránh “retry amplification”.

## 9. Pool/client lifecycle

Composition root tạo shared pool/client và đóng khi process shutdown. Adapter không tự tạo client mỗi call và không close shared resource.

Health:

- liveness: process event loop còn sống;
- readiness: instance có nhận work đúng guarantee không;
- dependency health không phải lúc nào cũng ping mọi request;
- circuit open có thể ảnh hưởng readiness tùy capability.

## 10. Database adapter

Nó xử lý row mapping, SQLSTATE, transaction handle, locking. Xem [10 Database](../10-database/README.md). Repository không nên log SQL error rồi application log lại rồi handler log lần ba.

## 11. Redis adapter

Cache adapter cần key version, TTL, serialization, miss/error policy và invalidation. Cache unavailable thường có thể fallback DB; idempotency store unavailable thường không thể bypass an toàn. Cùng Redis technology, criticality khác.

## 12. Kafka adapter

Publisher adapter map application/domain event sang integration schema, chọn key/header/topic, classify broker error. Topic name không đi vào domain.

Producer success có thể chỉ nghĩa broker accepted, không nghĩa consumer applied. Outbox cần khi DB state và event publication phải liên kết.

## 13. External HTTP adapter

Gateway nói bằng intent:

~~~go
type PaymentGateway interface {
	Authorize(ctx context.Context, request Authorization) (AuthorizationResult, error)
}
~~~

Adapter map vendor request/response, status, timeout và idempotency header. Không trả thẳng SDK response.

## 14. Decorator và composition

Infrastructure behavior có thể compose:

~~~text
metrics(repository)
  → cache(repository)
    → postgres repository
~~~

Order ảnh hưởng semantics. Metrics bên ngoài cache đo use-case perceived latency; bên trong chỉ đo DB. Tracing cần span parent đúng. Cache decorator write failure policy phải rõ.

## 15. Contract và integration tests

Contract suite áp cùng stable semantics lên adapters. Integration test dùng real dependency để kiểm protocol.

| Test | Chứng minh |
|---|---|
| Mapper unit | conversion/invariant |
| Contract | port behavior |
| Driver mock | rare branch/call shape |
| Real Postgres/Redis/Kafka | schema/query/offset/protocol |
| End-to-end | wiring + major workflow |

Mock driver không thay real integration test.

## 16. Production failure matrix

| Failure | Adapter concern | Application concern |
|---|---|---|
| PostgreSQL down | classify/cancel/pool | fail/retry policy |
| Redis timeout | bounded attempt | fallback hay fail closed |
| Kafka unavailable | publish error | outbox/pending state |
| Payment timeout | ambiguous result | inquiry/reconciliation |
| malformed provider response | mapping error | safe failure |

## 17. Debug/investigation

1. xác định adapter nào vi phạm latency/error contract;
2. xem trace span I/O và pool/client metrics;
3. giữ wrapped cause/SQLSTATE/status;
4. kiểm timeout/retry layers;
5. replay sanitized fixture;
6. tái hiện với real dependency;
7. so contract tests giữa implementations.

## 18. Wrong patterns

### Generic infrastructure package

Một package infra import mọi domain và trả interface{} trở thành dependency hub.

### Vendor-neutral port quá rộng

~~~go
type UniversalDatabase interface {
	Query(string, ...any) any
}
~~~

Core vẫn nói SQL nhưng mất type safety. Đây không phải inversion có ích.

### Fallback làm mất guarantee

Database down rồi tự ghi memory và trả success là data loss. Fallback phải preserve contract hoặc công bố degraded semantics.

## 19. Khi nào không cần adapter/interface?

Một command migrate chỉ dùng pgx có thể gọi driver trực tiếp. Thin reporting service có thể query SQL thẳng. Wrapper một method cùng signature chỉ thêm navigation nếu không có mapping, policy hoặc inversion need.

## 20. Bài tập

1. Implement Postgres Repository và contract tests.
2. Thiết kế Redis cache decorator với fail-open policy.
3. Thiết kế Payment Gateway phân biệt declined/timeout/unknown.
4. Vẽ retry budget qua HTTP → app → provider.
5. Review một adapter đang trả raw SDK model.

## 21. Mastery questions

1. “Infrastructure là detail” không cho phép điều gì?
2. Kafka consumer/producer nằm hai phía nào?
3. Port xóa được và không xóa được semantic difference nào?
4. Vì sao fallback memory có thể vi phạm contract?
5. Retry amplification hình thành thế nào?
6. Contract test và integration test bổ sung nhau ra sao?
7. Ai sở hữu client Close?
8. Khi nào wrapper driver là abstraction giả?

## Further reading

- Hexagonal Architecture / Ports and Adapters.
- Go context, net/http và database driver docs.
- PostgreSQL, Kafka, Redis official operations guides.
- Release It!, Michael Nygard.

## Quality gate

- [x] Problem, model và dependency analysis
- [x] Mapping/error/timeout/retry/lifecycle
- [x] Postgres/Redis/Kafka/external scenarios
- [x] Contract/integration tests
- [x] Failure/debug/wrong examples
- [x] Trade-off, exercises, mastery
