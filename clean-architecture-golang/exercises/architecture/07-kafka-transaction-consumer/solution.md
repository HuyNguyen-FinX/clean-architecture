# Solution tham khảo: Kafka Consumer Xử Lý Transaction

## Boundary

Kafka consumer is driving adapter:

~~~text
record → strict envelope/schema mapper → TransferMoneyCommand → use case
~~~

Kafka message is integration DTO, not command/domain event. Topic/partition/offset stay adapter diagnostics.

## Idempotency/inbox

Event envelope has stable EventID. Within same PostgreSQL transaction as transfer:

~~~sql
INSERT INTO processed_messages(consumer_name, event_id)
VALUES ($1,$2) ON CONFLICT DO NOTHING;
~~~

If already completed, return success. If new, apply transfer + mark/record. A separate Redis check is not atomic with Postgres.

## Offset policy

- commit after DB transaction success/known duplicate;
- retryable error: no commit or retry-topic publish durably;
- permanent decode/version: DLQ/quarantine then commit according to policy;
- shutdown: stop fetch, drain, commit completed only.

Auto-commit before effect risks message loss.

## Retry

Bounded in-process for brief transient; retry topic prevents head-of-line but can reorder. Business reject may be permanent outcome, not infrastructure retry. Deadlock/serialization retry reloads Aggregates.

## Transaction nesting

Consumer/inbox and Transfer use case both wanting transactions can nest incorrectly. Options:

- message application handler owns UoW and calls transaction-aware transfer operation;
- Transfer supports caller transaction contract;
- inbox integrated into same use-case idempotency.

Document one boundary.

## DLQ

Keep original topic/partition/offset, EventID/schema, error category, attempts/timestamps and protected payload. Provide repair/replay owner.

## Tests

- duplicate delivers once;
- crash-after-DB-before-offset redelivery;
- malformed permanent;
- transient retry;
- DB rollback;
- broker offset/rebalance integration;
- out-of-order/version;
- graceful shutdown.

## Outgoing event

If transfer publishes MoneyTransferred, write outbox in same DB transaction. Worker duplicate remains; downstream idempotent.
