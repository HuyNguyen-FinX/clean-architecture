# Case Studies: cùng nguyên lý, mức kiến trúc khác nhau

Case study dùng để luyện quyết định theo context. Không bắt đầu bằng folder; bắt đầu bằng actor, invariants, consistency, failure và cost of change.

## Cách phân tích

1. Actors/outcomes.
2. Ubiquitous language/business rules.
3. Data ownership/consistency.
4. Use cases và transaction boundaries.
5. Driving/driven ports.
6. Failure/idempotency/reconciliation.
7. Test/observability/deployment.
8. Simplest architecture đủ dùng.

## So sánh nhanh

| Case | Domain richness | Consistency | Async | Kiến trúc khởi đầu |
|---|---|---|---|---|
| Todo | thấp → vừa | local | không cần | handler/service/store |
| E-commerce Order | cao | local + Saga | events | Aggregate + app workflow |
| Payment | cao/failure-heavy | provider ambiguous | webhook | state machine + gateway |
| Banking | invariant cao | transaction/lock | outbox | rich domain + UoW |
| Loan | policy/lifecycle cao | long-running | workflow | DDD contexts |
| Kafka Worker | workflow | inbox tx | native | adapter + use case |
| Batch | item/chunk | partial | internal queue | job orchestration |
| Microservice | tùy domain | distributed | contracts | start modular ownership |

## Case 1: Todo API

Đọc [Todo API](../case-studies/01-todo-api/README.md).

Điểm quyết định: nếu chỉ CRUD, four-layer mapping là ceremony. Khi xuất hiện transition archived/completed, permissions và scheduling, extract behavior/capability dần.

Failure production: concurrent edits cần version/ETag; không phải cứ Todo là trivial mãi.

## Case 2: E-commerce Order

Đọc [E-commerce Order](../case-studies/02-ecommerce-order/README.md).

Order Aggregate giữ items/total/status. Inventory/Payment là remote contexts; không nằm trong local transaction. Saga/Pending state/compensation và outbox.

## Case 3: Payment

Đọc [Payment Service](../case-studies/03-payment-service/README.md).

Trọng tâm không phải provider interface đẹp mà idempotency, unknown outcome, webhook duplicates/out-of-order, reconciliation và audit.

## Case 4: Banking

Đọc [Banking Account](../case-studies/04-banking-account/README.md).

Money/Account invariant, transfer transaction, lock order, history/ledger, idempotency/outbox. Mini-banking là executable thread của curriculum.

## Case 5: Loan

Đọc [Loan Service](../case-studies/05-loan-service/README.md).

Strategic DDD có leverage: Origination, Risk, Disbursement contexts; language/state machine; external scoring ACL; human/manual path.

## Case 6: Kafka Worker

Đọc [Kafka Worker](../case-studies/06-kafka-worker/README.md).

Consumer là adapter; inbox marker cùng effect; retry/DLQ/offset/rebalance/shutdown; producer outbox.

## Case 7: Batch

Đọc [Batch Processing](../case-studies/07-batch-processing/README.md).

Chunk/checkpoint/partial failure/bounded workers. Domain rule thuần, orchestration ở job application, delivery scheduler/CLI.

## Case 8: Microservice

Đọc [Microservice](../case-studies/08-microservice/README.md).

Service boundary theo data/domain/team ownership, không theo technical layer. Distributed failure/contract/deployment cost trước khi split.

## Cross-case reasoning

### Repository

- Todo CRUD: store concrete/direct SQL có thể đủ.
- Banking: Aggregate Repository + lock/transaction semantics.
- Reporting/batch: query/stream port, không fake Aggregate.

### Transaction

- Todo update: one statement.
- Banking transfer: two accounts + record + idempotency + outbox.
- Order/payment: local transaction + Saga, không distributed ACID.

### Domain model

- Todo CRUD: record.
- Order/Loan/Banking: state transition/Aggregate.
- Kafka Worker: domain có thể ở called use case, consumer không cần invent Entity.

### Testing

- Todo: handler/store integration đủ.
- Banking: domain matrix + concurrent Postgres.
- Payment: provider fixture/timeouts/webhooks/reconcile.
- Kafka: broker/offset/duplicate.
- Batch: restart/checkpoint/property.

## Architecture decision template

~~~text
Context:
Forces and risks:
Decision:
Alternatives:
Consequences:
Guarantees:
Validation/tests:
Revisit trigger:
~~~

ADR không cần dài; phải ghi why/trade-off/trigger.

## Production synthesis: Transfer

~~~text
HTTP idempotency key
→ Application transaction
→ Account invariants
→ Postgres locks
→ Transfer + outbox
→ async publish
→ idempotent consumer
→ history projection
→ observability/reconciliation
~~~

Mỗi boundary giải failure riêng. Bỏ một mảnh không được che bằng từ “Clean”.

## Bài tập

1. Chọn hai case và dùng cùng folder tree, chỉ ra chỗ tree không hợp.
2. Thiết kế minimum architecture cho Todo rồi evolution triggers.
3. Vẽ Payment unknown-outcome state machine.
4. So sánh Bank transaction và Order Saga.
5. Viết ADR monolith vs microservice cho Loan.

## Mastery questions

1. Case nào cần Repository semantic nhất?
2. Vì sao Payment timeout khác rejection?
3. Kafka Worker có cần rich domain riêng không?
4. Todo complexity trigger nào làm tách domain đáng giá?
5. Banking và Order consistency khác nhau?
6. Case study nào cần reconciliation và vì sao?

## Quality gate

- [x] Comparative framework/matrix
- [x] Eight context-specific decisions
- [x] Transactions/repository/testing differences
- [x] Production synthesis
- [x] ADR/exercises/mastery
