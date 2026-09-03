# 14 Kafka Event Driven

## Tại sao cần học?

Kafka đưa Clean Architecture vào môi trường asynchronous. Consumer và producer đều là adapter, nhưng event-driven system còn cần idempotency, retry, DLQ và outbox.

## Flow Consumer

```text
Kafka Consumer
↓
Message Adapter
↓
Use Case
↓
Domain
```

## Flow Producer

```text
Use Case / Domain Event
↓
Event Publisher Port
↑
Kafka Producer Adapter
```

## Production Concerns

- Message duplication.
- Ordering.
- Retry và backoff.
- Dead letter queue.
- Transactional outbox.
- Schema evolution.

## Anti-pattern

- Domain entity biết Kafka topic.
- Consumer xử lý toàn bộ business logic trong callback.
- Publish event trong transaction không có outbox nhưng vẫn giả định exactly-once.

## Mastery Check

- [ ] Tôi biết Kafka consumer thuộc delivery/infrastructure adapter.
- [ ] Tôi biết event schema khác domain model.
- [ ] Tôi biết Clean Architecture không tự giải quyết message duplication.
