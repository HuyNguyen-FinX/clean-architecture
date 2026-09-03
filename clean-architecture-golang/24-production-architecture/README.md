# Production Architecture: lifecycle và reliability là một phần thiết kế

Tutorial thường dừng ở handler → use case → repository. Production phải sống qua rollout, overload, dependency outage, partial failure, schema change và shutdown. Clean boundaries giúp đặt policy, nhưng không thay timeout/idempotency/outbox/operations.

## Kết quả học tập

- thiết kế startup/shutdown/resource ownership;
- phân biệt liveness/readiness/startup;
- phân bổ timeout/retry/concurrency budgets;
- quản config/secrets/migrations;
- triển khai idempotency/outbox/reconciliation;
- lập SLO, capacity, deployment và incident runbook.

## 1. Ba level

### Level 1

Production-ready nghĩa process khởi động đúng, xử lý đúng khi dependency lỗi và dừng không làm mất work.

### Level 2

Engineer cấu hình server/pool/client, probes, signal, retry, telemetry, migrations và CI/CD.

### Level 3

Reliability là end-to-end guarantee gắn product semantics. Local transaction, async delivery, retry và status API phải tạo một state machine có thể vận hành/reconcile.

## 2. Production topology

~~~mermaid
flowchart LR
    LB["Load Balancer"] --> API["API replicas"]
    API --> PG[("PostgreSQL")]
    API --> REDIS[("Redis")]
    API --> OUTBOX[("Outbox table")]
    WORKER["Outbox workers"] --> OUTBOX
    WORKER --> KAFKA[("Kafka")]
    CONSUMER["Consumers"] --> KAFKA
    API --> OTEL["Telemetry collector"]
    WORKER --> OTEL
~~~

Mỗi arrow là failure/latency/backpressure boundary.

## 3. Configuration

Load một lần:

~~~go
type Config struct {
	HTTPAddress     string
	DatabaseURL     string
	ShutdownTimeout time.Duration
	MaxDBConns      int32
}
~~~

Validate:

- required;
- ranges/durations;
- cross-field budgets;
- environment identity;
- secret references.

Không log DSN/token. Không mutable global config. Dynamic config cần version/validation/rollback riêng.

## 4. Startup

~~~text
load config
→ logger/telemetry
→ pools/clients
→ ping critical dependencies
→ migrations check
→ build object graph
→ start workers/listeners
→ readiness true
~~~

Fail-fast dependency bắt buộc. Không fallback durable DB sang memory. Startup có timeout và logs đủ để sửa config.

## 5. Resource ownership

Composition root tạo pool/client/server, nên đóng chúng. Repository không close shared pool. Cleanup chạy reverse dependency order.

Tránh log.Fatal sau defers vì os.Exit bỏ defer. main gọi run và log/exit sau run cleanup.

## 6. Graceful shutdown

~~~go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()

go func() {
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}()

err := server.ListenAndServe()
~~~

Sequence:

1. readiness false;
2. stop fetch/new work;
3. shutdown HTTP;
4. drain workers;
5. flush producer/telemetry;
6. close clients/pool;
7. bounded hard stop.

## 7. Health semantics

- Liveness: restart có giúp không? Event loop/process alive.
- Readiness: instance có nhận work đúng guarantee không?
- Startup probe: app còn initializing/migrating.

DB required cho Transfer: DB down → unready. Redis performance cache down → vẫn ready degraded. Kafka down + outbox durable → API có thể ready, nhưng outbox-age alert.

Không ping dependencies quá nặng mỗi probe.

## 8. Timeout budget

Client 2s:

~~~text
LB/network 200ms
HTTP/app 200ms
DB lock/query 800ms
external 500ms
response margin 300ms
~~~

Parent context là end-to-end budget. Mỗi adapter per-attempt cap nhỏ hơn remaining. Timeout layers không được mâu thuẫn làm proxy cắt trước app diagnostics.

## 9. Retry budget

Retry chỉ transient/idempotent. Max attempts/backoff/jitter trong deadline. Retry amplification:

~~~text
client 3 × API 3 × SDK 3 = 27 attempts
~~~

Chọn một layer chủ retry hoặc phối hợp budget/header. Metrics count original requests và attempts riêng.

## 10. Idempotency

POST Transfer cần durable key:

~~~text
new key + request hash
→ processing
→ completed + stable response
~~~

Atomic cùng business effect. Same key khác hash → conflict. Concurrent duplicate phải lock/wait/replay. Retention theo client retry window/compliance.

## 11. Transaction + outbox

Trong một DB transaction:

~~~text
lock accounts
update balances
insert transfer
complete idempotency
insert outbox
commit
~~~

Sau commit, worker publish at-least-once. Consumer inbox idempotent. Reconciliation kiểm transfer/outbox/consumer effect.

## 12. Rate limit và load shedding

Rate limit theo tenant/actor/IP tùy threat/product. Load shedding khi queue/pool saturation để giữ tail latency và recovery.

429 cho quota; 503 cho capacity/dependency. Retry-After có policy. Không nhận vô hạn rồi timeout.

## 13. Concurrency budgets

Align:

- HTTP max/in-flight;
- DB pool;
- worker count;
- external connection/quota;
- Kafka partitions;
- CPU/memory.

Tăng replica/conns/workers có thể overload shared DB. Capacity planning nhìn tổng fleet.

## 14. Circuit breaker/bulkhead

External dependency chậm: bounded concurrency, queue, timeout, breaker. Cache/notification có thể degraded; money authorization có Pending state.

Không để breaker state ẩn trong one-off client không observable.

## 15. Database migrations

Rolling deploy nghĩa old/new app cùng schema. Expand-contract:

1. additive schema;
2. compatible code;
3. backfill;
4. enforce;
5. switch reads;
6. cleanup later.

Migration lock/table rewrite được đo; backup/recovery rehearsal; schema version readiness.

## 16. Deployment

Canary/progressive rollout:

- health + SLI gates;
- backward compatible events/schema;
- feature flag kill switch;
- rollback app không rollback data tự động;
- outbox/consumer versions coexist.

Deploy worker trước/after producer theo schema compatibility.

## 17. Secrets/security

- secret manager/file/env injection, không repo;
- least privilege DB/Kafka;
- TLS/mTLS;
- rotate credential;
- dependency/supply-chain scanning;
- request/body limits;
- audit sensitive operations;
- redaction.

Clean layer không tự tạo security.

## 18. Observability/SLO

SLIs:

- transfer API availability/latency;
- idempotency conflict/duplicate;
- transaction retry/lock wait;
- outbox oldest age;
- consumer lag/DLQ;
- reconciliation mismatch.

Alert theo user/business impact. Dashboard có deploy markers. Trace HTTP → DB và outbox worker → Kafka.

## 19. Reconciliation

Money system cần periodic independent check:

- double-entry ledger sums;
- transfer vs account movements;
- idempotency records;
- outbox missing/stuck;
- provider settlement.

Repair là audited command, không manual SQL không dấu vết.

## 20. Failure scenarios

### PostgreSQL down

Fail requests bounded, readiness false, no memory fallback, retries controlled.

### Kafka down

Commit transfer + outbox, backlog alert/capacity, publish after recovery.

### Redis down

Performance cache fail-open; durable idempotency fail-closed nếu Redis là source (tốt hơn dùng DB).

### Commit OK, response lost

Client retry key, replay completed response/status.

### Process SIGTERM giữa message

Stop fetch, finish/rollback, commit offset only completed.

### Region/network partition

Không multi-primary money write nếu conflict model chưa chứng minh. RTO/RPO và failover runbook.

## 21. Production checklist theo boundary

| Boundary | Checks |
|---|---|
| Domain | invariant, overflow, state transitions |
| Application | transaction, idempotency, authorization |
| HTTP/gRPC | limits, auth, timeout, safe errors |
| PostgreSQL | pool, lock, migration, backup |
| Kafka/outbox | duplicate, lag, lag, DLQ |
| Redis | stale/failure/eviction |
| External | unknown outcome/reconcile |
| Process | signal, drain, probes, telemetry |

## 22. Testing/release gates

- unit/race/fuzz;
- adapter integration real dependencies;
- migration compatibility;
- E2E duplicate/retry;
- load hot account;
- chaos dependency outage;
- restore/reconciliation drill;
- architecture/security scanning.

## 23. Incident response

1. detect SLO burn/business mismatch;
2. establish impact/timeline;
3. mitigate load/deploy/dependency;
4. preserve evidence;
5. reconcile/repair;
6. root cause across boundaries;
7. add guard/test/signal/runbook.

## 24. Khi nào production stack đơn giản hơn?

Low-risk internal tool không cần Kafka/Redis/OTel collector riêng. Managed DB + HTTP + structured logs/backups có thể đủ. Production-ready là phù hợp risk/SLO, không phải đủ logo.

## 25. Lab

Làm [Lab 12: Full Application](../labs/lab-12-full-application/README.md): ghép domain/use case/adapters, graceful lifecycle, idempotency/outbox design và test portfolio.

## 26. Mastery questions

1. Liveness/readiness khác nhau?
2. Vì sao log.Fatal phá deferred cleanup?
3. Retry amplification hình thành thế nào?
4. Kafka down nhưng outbox hoạt động có unready không?
5. Deploy rollback khác data rollback?
6. Idempotency record nằm transaction nào?
7. Capacity budget liên hệ fleet/DB?
8. Reconciliation vì sao bắt buộc trong money system?

## Further reading

- Go net/http Server, os/signal và context docs.
- Kubernetes probes/termination lifecycle docs.
- Google SRE books/workbook.
- PostgreSQL operations/migration/backup docs.
- OpenTelemetry/Prometheus/Kafka official docs.

## Quality gate

- [x] Lifecycle/config/resources/probes
- [x] Timeout/retry/concurrency/load policies
- [x] Idempotency/transaction/outbox/reconcile
- [x] Migration/deployment/security/SLO
- [x] Failure/incident/test scenarios
- [x] Trade-off, lab, mastery
