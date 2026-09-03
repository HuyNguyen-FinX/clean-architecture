# System Design: từ package boundary tới data ownership

Clean Architecture chủ yếu kiểm source dependency trong một deployable unit. System design thêm capacity, service/data boundaries, network, consistency, availability và operations.

## Kết quả học tập

- phân biệt module/service/deployment boundary;
- estimate traffic/storage/connections;
- chọn data ownership và consistency;
- thiết kế transfer path, idempotency/outbox/reconcile;
- phân tích partition/hot key/multi-region;
- trình bày design theo trade-off.

## 1. Clarify requirements

Transfer system:

Functional:

- deposit/withdraw/transfer;
- status/history;
- idempotent retry;
- notification.

Non-functional:

- 5.000 TPS average, burst 15.000;
- p99 POST < 500ms region-local;
- no money creation/loss;
- audit/reconciliation;
- RPO/RTO;
- compliance/data residency.

Không vẽ Kafka trước requirements.

## 2. Capacity rough math

5.000 TPS × 86.400 ≈ 432M transfers/day nếu liên tục. Nếu record + indexes/outbox ~2KB, ~864GB/day trước replication/overhead. Con số này buộc partition/retention/archival assumptions.

20 API replicas × 50 DB connections = 1.000. DB connection/process limits phải budget.

Estimation không cần chính xác tuyệt đối; cần lộ bottleneck/assumption.

## 3. Monolith, modular monolith, microservices

Clean architecture chạy trong cả ba.

Start modular monolith khi:

- team nhỏ;
- transactions cross modules quan trọng;
- scale profile tương tự;
- operational platform hạn chế.

Split service khi:

- ownership/team autonomy;
- independent scale/release;
- security/isolation;
- context language/data rõ;
- distributed cost đáng trả.

Không split handlers/services/repositories thành services.

## 4. Data ownership

Account context sở hữu balances/ledger. Notification không query accounts table trực tiếp; nhận published event/read API.

Shared database có thể pragmatic nhưng schema/table ownership và access rules rõ. Database-per-service không tự ngăn semantic coupling qua synchronous chains.

## 5. High-level transfer

~~~mermaid
flowchart LR
    C["Client"] --> API["Transfer API"]
    API --> DB[("PostgreSQL/Ledger")]
    DB --> OUT["Outbox"]
    WORKER["Publisher"] --> OUT
    WORKER --> K[("Kafka")]
    K --> HIST["History projection"]
    K --> NOTIFY["Notification"]
    RECON["Reconciliation"] --> DB
~~~

Synchronous path ngắn giữ money consistency. Async side effects không nằm trong DB transaction network call.

## 6. Balance vs ledger

Mutable balance row:

- read nhanh;
- hot account contention;
- audit cần transfer table.

Append-only double-entry ledger:

- every transfer debit/credit entries sum zero;
- audit/rebuild mạnh;
- current balance cần materialized snapshot/projection;
- posting/idempotency/concurrency phức tạp.

Core banking thường cần ledger model; tutorial Account balance là bước học, không final universal architecture.

## 7. Transaction boundary

Single database:

~~~text
idempotency insert/check
lock/reserve accounts
ledger entries/transfer
balance snapshots
outbox
commit
~~~

Không gọi Kafka/payment remote trong transaction. Lock IDs deterministic. Hot key may serialize.

## 8. Idempotency

Key scoped tenant+operation; request hash; states processing/completed; stable result. Unique constraint handles concurrent duplicates. Status lookup supports ambiguous client timeout.

Multi-region key ownership must be globally consistent or routed to home region.

## 9. Partitioning

Partition by AccountID keeps account-local operations but transfer spans two partitions. Choices:

- database supports cross-partition transaction;
- route by source and asynchronously credit;
- reservation/saga;
- ledger service central sequencing.

Trade product consistency for scale only explicitly. Hash partition spreads load except celebrity/hot account.

## 10. Hot account

Merchant account receives huge credits. Options:

- append ledger entries without locking receiver balance synchronously;
- sharded counters/buckets with reconciled total;
- asynchronous receiver projection;
- queue/serialize per account;
- separate available vs posted.

Each changes read freshness/withdraw semantics.

## 11. Consistency

- Account withdrawal invariant: strong/local.
- Notification: eventual.
- History projection: bounded eventual.
- Fraud/risk: pending workflow if remote.
- Analytics: eventual/batch.

Không chọn one consistency model toàn system.

## 12. Availability and partitions

CAP không phải chọn CA/AP một lần. Trong network partition, operation cụ thể chọn:

- reject/queue writes để tránh double spend;
- serve stale read với timestamp;
- route home region;
- accept pending command.

Money writes thường prefer consistency; read/notification có thể degrade.

## 13. Multi-region

Active-passive:

- simpler write ownership;
- failover RTO;
- replica lag/RPO;
- split-brain fencing.

Active-active:

- lower local latency;
- conflict/global ordering hard;
- account home-region routing;
- idempotency/global ledger.

Không multi-primary mutable balance bằng last-write-wins.

## 14. Messaging

Kafka partitions/order/at-least-once. Outbox, inbox, schema registry, DLQ/replay. Broker outage capacity measured bằng outbox growth. Consumers idempotent.

## 15. Cache

Cache history/account summary with stale contract, not authoritative withdrawal decision. Redis down fail-open for read cache. Key/version/TTL/stampede.

## 16. API

POST /transfers + Idempotency-Key returns TransferID/status. GET /transfers/{id}. 201 for posted synchronous; 202 for pending distributed workflow.

Pagination cursor. Auth/tenant/rate-limit/body limits. Stable error codes.

## 17. Reliability

- timeouts budgets;
- bounded retries;
- circuit/bulkhead for remotes;
- load shedding;
- graceful shutdown;
- migrations expand-contract;
- backups/restore drills;
- reconciliation.

## 18. Observability

SLIs:

- posted transfer correctness/latency;
- rejected reason;
- DB lock/pool;
- idempotency replay/conflict;
- outbox age;
- consumer lag;
- ledger imbalance.

Trace helps latency; ledger/reconciliation proves financial correctness.

## 19. Security

- authn/authz account ownership;
- encryption transit/at rest;
- least privilege;
- PII/token redaction;
- audit actor/action;
- fraud limits;
- replay protection;
- key rotation;
- tamper-evident records where required.

## 20. Failure walkthrough

### DB primary fails during commit

Client outcome unknown. Retry with key/status. Failover must avoid split-brain; reconcile.

### Kafka down

Transfer commits with outbox; notification/history lag; capacity/age alert.

### Region partition

Home-region account writes unavailable/queued rather than conflicting.

### Consumer duplicate

Inbox/event ID; same effect once; offset after transaction.

## 21. Testing

- domain/property;
- transaction/concurrency Postgres;
- idempotency E2E;
- outbox/consumer broker;
- load hot keys;
- failover/chaos;
- restore/reconcile;
- compatibility.

## 22. Evolution plan

~~~text
V1 modular monolith + Postgres
V2 outbox/workers
V3 read projections/cache
V4 ledger partition/archive
V5 split contexts only with ownership pressure
V6 multi-region after RTO/latency demands
~~~

Avoid future-proof distributed complexity before measurements.

## 23. Interview design flow

1. clarify;
2. estimate;
3. define invariants;
4. high-level components/data;
5. critical write path;
6. failure/concurrency;
7. scale bottleneck;
8. operations/security;
9. trade-offs/evolution.

## 24. Khi nào một service đủ?

Most systems at early scale: one Go deploy + Postgres + background worker can handle much. Clear modules/outbox and observability create evolution path. Microservices are organizational/distributed tool, not seniority badge.

## 25. Bài tập

1. Estimate 5k TPS storage.
2. Design hot merchant credit.
3. Compare balance row vs ledger.
4. Design active-passive failover.
5. Define consistency per feature.

## 26. Mastery questions

1. Clean Architecture không giải network issue nào?
2. Transfer partition key conflict?
3. Why no LWW balance?
4. 202 changes product semantics?
5. Outbox outage capacity calculate?
6. Reconciliation vs tracing?
7. When split service?
8. Shared DB can still have ownership?

## Further reading

- Designing Data-Intensive Applications.
- Google SRE books.
- PostgreSQL/Kafka official docs.
- DDD strategic design/Team Topologies.
- Building Microservices, read with monolith trade-offs.

## Quality gate

- [x] Requirements/capacity/data ownership
- [x] Transaction/ledger/idempotency
- [x] Partition/hot key/multi-region/consistency
- [x] API/cache/messaging/reliability/security
- [x] Failures/testing/evolution
- [x] Interview flow/trade-off/mastery
