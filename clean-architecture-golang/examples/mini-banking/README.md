# Mini Banking: Vertical Slice V1-V11

Đây là project xuyên suốt curriculum. Nó không cố mô phỏng toàn bộ core banking; nó dùng một flow đủ khó để chứng minh boundary, transaction, idempotency, messaging và lifecycle bằng code Go chạy được.

## Guarantee Đã Implement

| Guarantee | Cơ chế | Test chính |
|---|---|---|
| Money hợp lệ | Currency/overflow/equality trong Value Object | domain table tests |
| Account invariant | private state, overdraft và frozen transition | domain tests |
| Transfer atomic | Transactor bao hai Account + Transfer + key + outbox | memory rollback và PostgreSQL integration |
| Chống lost update | `SELECT FOR UPDATE`, lock Account theo ID ổn định | opposite-direction integration test |
| Retry idempotent | claim key + canonical request hash trong cùng transaction | replay/conflict/concurrent duplicate tests |
| History bền | immutable Transfer record cùng commit với balance | application/HTTP/PostgreSQL tests |
| DB/event intent atomic | transactional outbox | rollback và relay failure tests |
| At-least-once publish | relay mark chỉ sau Kafka ack | relay test; downstream vẫn phải deduplicate |
| Kafka input an toàn | event ID thành idempotency key, commit offset sau result/DLQ | decoder/adapter tests |
| Strict HTTP | body limit, one JSON value, unknown-field rejection, safe errors | `httptest` suite |
| Observability | structured logs, request/trace correlation, bounded metrics | middleware tests |
| Lifecycle | timeouts, SIGTERM cancellation, drain và resource close | composition code + unit suites |
| Dependency direction | AST import fitness test | `internal/architecture` |

PostgreSQL integration tests chỉ chạy khi có `TEST_DATABASE_URL`; không có biến này chúng vẫn compile rồi skip có thông báo. Kafka adapter compile/test không cần broker, còn broker integration là quality gate của môi trường deployment.

## Evolution

~~~text
V1  Money + Account domain/invariants
V2  TransferMoney application use case
V3  consumer-owned ports + memory adapter
V4  PostgreSQL mapping/migration
V5  transaction + deterministic row locking
V6  strict HTTP contract + stable error mapping
V7  domain/use-case/adapter/architecture test portfolio
V8  durable idempotency claim, hash conflict và replay
V9  Kafka producer + TransferRequested consumer adapters
V10 Transfer history + transactional outbox + relay
V11 structured logs + bounded metrics + trace correlation + graceful lifecycle
~~~

Mỗi version thêm một guarantee quan sát được. Folder hoặc interface mới không được tính là tiến hóa nếu không đổi failure semantics và test.

## Kiến Trúc

~~~mermaid
flowchart LR
    HTTP["HTTP adapter"] --> APP["Transfer/List/Relay use cases"]
    KCON["Kafka consumer"] --> APP
    APP --> DOMAIN["Account, Money, Transfer"]
    APP --> PORTS["Store, Transactor, Publisher ports"]
    MEM["Memory transactional adapter"] -.implements.-> PORTS
    PG["PostgreSQL adapter"] -.implements.-> PORTS
    KPUB["Kafka publisher"] -.implements.-> PORTS
~~~

Compile-time dependency:

~~~text
delivery/http -> application -> domain
infrastructure/memory|postgres|kafka -> application/domain
cmd/api -> all adapters for composition only
domain -> standard library
~~~

Runtime use case vẫn gọi PostgreSQL/Kafka object. Dependency Inversion nói về import/ownership: application sở hữu capability contract, adapter phụ thuộc ngược vào contract đó.

## Transaction Và Idempotency

Một request hợp lệ chạy như sau:

~~~text
BEGIN
  INSERT idempotency claim ON CONFLICT DO NOTHING
  SELECT claim FOR UPDATE
  nếu complete: replay transfer_id
  nếu hash khác: conflict
  SELECT hai Account theo stable ID order FOR UPDATE
  Account.Withdraw / Account.Deposit
  UPDATE hai Account
  INSERT Transfer
  INSERT Outbox(MoneyTransferred.v1)
  complete idempotency record
COMMIT
~~~

Claim trước business effect làm hai request đồng thời cùng key serialize trên một durable row. Nếu transaction fail, claim rollback theo. Commit xong nhưng HTTP response mất thì retry trả cùng `transfer_id`; dùng cùng key cho body khác trả conflict.

Memory adapter cũng clone toàn state trong `WithinTransaction`, nên test có rollback thật. Nó vẫn không có durability hoặc cross-process semantics như PostgreSQL.

## Outbox Và Kafka

Outbox relay đọc message chưa publish, gọi Kafka producer với `RequiredAcks=all`, rồi mới ghi `published_at`. Crash sau publish/trước mark tạo duplicate; đây là at-least-once có chủ đích, không phải exactly-once. Consumer downstream phải có inbox/idempotent effect.

`TransferRequested.v1` consumer:

1. Fetch message không auto-commit.
2. Decode envelope và lấy `event_id` làm idempotency key.
3. Gọi cùng `TransferMoneyUseCase` như HTTP.
4. Business rejection/permanent payload được ghi DLQ bền rồi commit offset.
5. Transient error không commit; worker dừng để supervisor/backoff policy xử lý.

Production có thể thay chính sách dừng bằng bounded retry/retry topic. Không nên thêm retry mù trong adapter vì ordering và outage policy là quyết định vận hành.

## Chạy Test

~~~bash
go test -race ./...
go vet ./...
~~~

PostgreSQL thật:

~~~bash
TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable \
  go test -race ./internal/account/infrastructure/postgres
~~~

Integration suite truncate bốn bảng test và kiểm commit/rollback, history/outbox, opposite transfer cùng concurrent duplicate key. Chỉ dùng database disposable.

## Chạy API

Memory mode tự seed hai Account:

~~~bash
go run ./cmd/api
~~~

PostgreSQL mode:

~~~bash
DATABASE_URL=postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable \
AUTO_MIGRATE=1 \
go run ./cmd/api
~~~

`AUTO_MIGRATE=1` chỉ tiện cho học/local. Production chạy migration bằng job có lock, quyền và rollback/roll-forward plan riêng.

Tạo Transfer:

~~~bash
curl -i -X POST http://localhost:8080/transfers \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: demo-transfer-001' \
  -d '{"from_account_id":"A-100","to_account_id":"B-200","amount":500000,"currency":"VND"}'
~~~

Gửi lại đúng request trả `200` và `"replayed": true`; lần đầu trả `201`. Xem history:

~~~bash
curl 'http://localhost:8080/accounts/A-100/transfers?limit=25'
curl 'http://localhost:8080/metrics'
~~~

Metrics chỉ dùng method/route template/status làm labels, không dùng Account ID. `traceparent` hợp lệ được tiếp nhận; response và structured log có `X-Trace-ID`/`trace_id`.

## Chạy Kafka

Khi có broker:

~~~bash
KAFKA_BROKERS=localhost:9092 go run ./cmd/api
~~~

Outbox relay được bật và publish tới topic trong row (`money-transferred-v1`). Bật thêm input consumer:

~~~bash
KAFKA_BROKERS=localhost:9092 KAFKA_CONSUMER_ENABLED=1 go run ./cmd/api
~~~

Topics mặc định: `transfer-requests-v1`, `transfer-requests-dlq-v1`, `money-transferred-v1`. Hạ tầng production cần tạo topic/partition/retention, TLS/SASL, quota và broker integration test; project không tự tạo chúng lúc startup.

Module giữ Go 1.22 nên Kafka client được pin `kafka-go v0.4.47`; [release notes](https://github.com/segmentio/kafka-go/releases) ghi dòng v0.4.49 đã nâng baseline lên Go 1.23. Việc pin là compatibility decision có chủ đích, không phải khuyến nghị dùng version cũ cho mọi project.

## Failure Windows Cần Nhớ

| Failure | Hành vi |
|---|---|
| Receiver save fail | transaction rollback toàn bộ artifacts |
| Commit thành công, response mất | retry cùng key replay |
| Kafka down | transfer vẫn commit, outbox pending |
| Publish xong, mark fail | relay publish lại; duplicate được chấp nhận |
| Duplicate Kafka input | same event ID replay, không chuyển tiền lần hai |
| Invalid Kafka payload | publish DLQ xong mới commit offset |
| SIGTERM | dừng server/consumer, cancel relay, chờ goroutine rồi close clients |

## Giới Hạn Có Chủ Đích

- Model balance mutable phục vụ học invariant; core banking thật nên cân nhắc double-entry ledger và reconciliation.
- Outbox query hiện chấp nhận nhiều relay có thể cùng publish một row; duplicate là an toàn về correctness nhưng production lớn nên claim/lease bằng `SKIP LOCKED`.
- Metrics exporter nhỏ tương thích Prometheus text cho mục đích học. Deployment lớn nên dùng Prometheus/OpenTelemetry SDK, histogram buckets và exporter chuẩn ở outer layer.
- Consumer transient error dừng để supervisor restart. Retry budget/backoff/retry topic phải được chọn theo SLO và ordering của hệ thống thật.
- Không có auth/authorization, fraud, limits, PII policy hoặc multi-region ledger; Clean Architecture không tự cung cấp các guarantee đó.

## Hướng Đọc Code

1. Bắt đầu từ `internal/account/domain` và chạy domain tests.
2. Đọc `application/transfer_money.go`, xác định atomic boundary và port ownership.
3. So sánh transaction memory với PostgreSQL context-bound transaction.
4. Theo event từ `AddOutbox` tới Kafka `Publisher`.
5. Theo request từ HTTP/Kafka vào cùng use case.
6. Đọc architecture fitness test để thấy dependency được kiểm bằng source, không bằng sơ đồ.
