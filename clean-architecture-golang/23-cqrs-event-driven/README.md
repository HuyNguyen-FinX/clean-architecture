# CQRS, Event Sourcing và Saga: tách vì model khác, không vì trend

CQRS chỉ yêu cầu tách responsibility đọc và ghi. Nó không bắt buộc hai service, Kafka hay Event Sourcing. Mỗi mức tách thêm consistency/operations cost và chỉ đáng dùng khi read/write forces thật sự khác.

## Kết quả học tập

- áp dụng CQRS nhẹ bằng read port/projection;
- phân biệt CQRS, Event-Driven và Event Sourcing;
- thiết kế projection/idempotency/rebuild;
- model Saga/process manager;
- phân tích eventual consistency/compensation;
- biết khi không dùng.

## 1. Problem

Transaction history list cần join/filter/pagination. Dùng Aggregate Repository:

~~~text
load 100 Transfer Aggregates
load Accounts
map to display
~~~

Write model bảo vệ invariant nhưng read cần projection nhanh. Ép một model phục vụ cả hai tạo N+1 và public getters vô nghĩa.

## 2. Ba level

### Level 1

Command thay đổi state; Query đọc state. Hai model có thể tách khi nhu cầu khác.

### Level 2

Engineer tạo command handler, query handler/read model, projection và consistency contract.

### Level 3

Physical separation, events và Event Sourcing chỉ là lựa chọn nâng cao. Product phải chấp nhận stale reads, rebuild/operations và duplicate.

## 3. CQRS nhẹ trong một process/database

~~~go
type TransferHistory interface {
	ListByAccount(
		context.Context,
		AccountID,
		Page,
	) ([]TransferItem, NextCursor, error)
}
~~~

Postgres adapter query projection trực tiếp. Command side dùng Aggregate Repository. Cùng DB, cùng deploy, ít operational cost nhưng model đọc/ghi rõ.

## 4. Command

Command là intent:

~~~go
type TransferMoney struct {
	From, To AccountID
	Amount   Money
	IdempotencyKey string
}
~~~

Handler validate actor/workflow, transaction và trả result. Không trả mutable Aggregate.

## 5. Query

Query không thay state:

~~~go
type GetTransferStatus struct{ TransferID TransferID }
~~~

Read handler có thể dùng replica/cache/projection. “Không thay state” không có nghĩa miễn auth/rate-limit/observability.

## 6. Tách physical

Read database riêng:

~~~text
Command DB commit + outbox
→ event
→ projector
→ Read DB
→ Query API
~~~

Read có lag. API/UI cần trạng thái pending, last-updated, read-your-write token hoặc fallback nếu guarantee cần.

## 7. Projection

Projector idempotent:

~~~sql
INSERT INTO account_history(event_id, account_id, ...)
VALUES (...)
ON CONFLICT (event_id) DO NOTHING;
~~~

Nếu event update snapshot theo aggregate version, ignore duplicate/older và detect gaps. Projection code/version/migration cần test.

## 8. Rebuild

Read model có thể rebuild từ source events/outbox/history:

- new table/version;
- replay in order per key;
- monitor lag/error;
- dual-read compare;
- cut traffic;
- retire old projection.

Replay không được kích external side effects. Namespace inbox cho replay.

## 9. Event-Driven không đồng nghĩa CQRS

App có thể publish events mà read/write model chung. CQRS có thể synchronous cùng DB không Kafka. Hai patterns orthogonal.

## 10. Event Sourcing

State được derive từ event stream:

~~~text
AccountOpened
MoneyDeposited
MoneyWithdrawn
AccountFrozen
~~~

Repository append events với expected version; Aggregate fold events.

~~~go
func Rehydrate(events []Event) (*Account, error) {
	account := &Account{}
	for _, event := range events {
		account.apply(event)
	}
	return account, nil
}
~~~

## 11. Event store concurrency

~~~sql
INSERT INTO account_events(aggregate_id, version, type, payload)
VALUES ($1, $2, $3, $4);
~~~

Unique(aggregate_id, version) phát hiện concurrent append. Conflict reload/retry theo policy.

Event schema là forever-ish data; evolution/upcasting khó hơn row migration.

## 12. Snapshot

Stream dài replay chậm. Snapshot tại version N rồi apply events N+1. Snapshot là cache, phải rebuild được và verify version/hash.

## 13. Event Sourcing audit không tự đủ compliance

Events immutable giúp timeline nhưng cần:

- actor/source;
- clock/order/integrity;
- PII deletion/encryption policy;
- authorization;
- schema evolution;
- tamper evidence;
- reconciliation.

Đừng chọn chỉ vì “cần audit log”.

## 14. Saga/process manager

Workflow qua services:

~~~text
TransferRequested
→ ReserveSource
→ CreditDestination
→ CompleteTransfer
~~~

Failure credit có thể ReleaseReservation. Saga state:

~~~text
pending_source → pending_destination → completed
              ↘ compensating → compensated/manual
~~~

Process manager là application component phản ứng event, phát command, persist state/idempotency.

## 15. Choreography vs orchestration

Choreography: services react events, coupling trực tiếp thấp nhưng flow ẩn, cyclic surprises.

Orchestration: coordinator nói bước tiếp, flow/timeout rõ nhưng coordinator có thể phình.

Chọn theo workflow complexity/ownership. Cả hai cần idempotency, correlation và observability.

## 16. Compensation không phải rollback

Credit đã xảy ra, compensation debit/refund là business operation mới có thể fail hoặc bị policy chặn. State phải giữ pending/manual; không xóa lịch sử.

## 17. Specification

Specification có thể diễn đạt reusable query/business predicates, nhưng đừng build universal AST để Repository translate mọi DB. Query object/read port cụ thể thường rõ hơn Go.

## 18. Consistency timeline

~~~text
t0 command accepted
t1 write DB commit
t2 outbox publish
t3 projector receives
t4 read model updated
~~~

GET ở t1.5 có thể stale. Define SLA lag, không giả vờ immediate. Client có thể query command status source.

## 19. Failure matrix

| Failure | Effect |
|---|---|
| outbox delayed | projection stale |
| duplicate event | projector idempotent |
| out-of-order | version/gap policy |
| poison event | projection stalls/quarantine |
| projector bug | rebuild new version |
| saga timeout | retry/compensate/manual |
| compensation fail | escalated state |

## 20. Testing

- command/domain unit;
- query adapter integration;
- event fold/given-when-then;
- projection duplicate/order/version;
- rebuild fixture from old schemas;
- saga transition table/timeouts;
- E2E eventual assertion với bounded polling;
- chaos broker/projector down.

Không sleep fixed 5s; poll with deadline and diagnostic.

## 21. Production scenario

History projection lag 20 phút nhưng transfers vẫn commit:

- POST returns transfer ID/status source;
- GET projection exposes last-updated/stale contract;
- lag SLO alert;
- projector scale/fix;
- no replay external notification;
- reconcile projection count/hash;
- product decides whether to pause writes.

## 22. Debug

Missing read item:

1. command/transfer source record;
2. outbox;
3. broker event/partition;
4. consumer offset/inbox;
5. projection version/error;
6. schema mapping;
7. rebuild/repair.

Saga stuck: correlate saga ID, expected next event, timeout job, duplicate/gap và compensation attempts.

## 23. Khi nào không dùng?

CRUD, same read/write shape, moderate scale, strong read-after-write: one model/DB often best. Event Sourcing/CQRS physical adds data stores, lag, replay, schemas và on-call burden.

## 24. Bài tập

1. Tạo TransferHistory read port.
2. Vẽ consistency timeline/UI states.
3. Implement idempotent projector.
4. Model transfer saga + compensation failure.
5. So sánh row audit table với Event Sourcing.

## 25. Mastery questions

1. CQRS có bắt Kafka/hai DB?
2. Event-Driven và CQRS khác nhau?
3. Projection rebuild không được làm gì?
4. Event Sourcing concurrency dùng expected version ra sao?
5. Compensation khác rollback?
6. Product phải biết eventual consistency gì?
7. Snapshot có là source of truth?
8. Khi nào choreography khó debug?

## Further reading

- Martin Fowler, CQRS và Event Sourcing.
- Microsoft/AWS architecture guidance về CQRS/Saga, đọc cùng trade-offs.
- DDD literature về Domain Events.
- Kafka/outbox chapter của curriculum.

## Quality gate

- [x] CQRS từ nhẹ đến physical
- [x] Command/query/projection/rebuild
- [x] Event Sourcing/version/snapshot/evolution
- [x] Saga/choreography/compensation
- [x] Failure/testing/production/debug
- [x] Trade-off, exercises, mastery
