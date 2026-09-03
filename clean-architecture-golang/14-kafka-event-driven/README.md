# Kafka và Event-Driven Architecture: delivery semantics trước sơ đồ

Kafka không làm hệ thống decoupled chỉ vì service không gọi HTTP trực tiếp. Consumer duplicate, ordering theo partition, schema evolution, retry và DB-event atomicity quyết định correctness.

## Kết quả học tập

- đặt Kafka consumer/publisher đúng adapter side;
- phân biệt Domain Event và Integration Event;
- implement idempotent consumer;
- thiết kế retry/DLQ/order;
- hiểu transactional outbox và failure matrix;
- test mapper/policy bằng fake và broker integration.

## 1. Hai chiều dependency

Consumer là driving adapter:

~~~text
Kafka record → decode/map → Application Use Case
~~~

Publisher là driven adapter:

~~~text
Application EventPublisher port → Kafka producer adapter → broker
~~~

~~~mermaid
flowchart LR
    BROKER[("Kafka")]
    CONSUMER["Kafka Consumer Adapter"] --> APP["Application"]
    APP --> DOMAIN["Domain"]
    APP --> PORT["Event Publisher Port"]
    KPROD["Kafka Producer Adapter"] -.implements.-> PORT
    BROKER --> CONSUMER
    KPROD --> BROKER
~~~

Domain không biết topic, partition, header hoặc Kafka library.

## 2. Ba level

### Level 1

Producer ghi event; consumer đọc và gọi use case. Một message có thể đến lại nên processing phải chịu duplicate.

### Level 2

Engineer quản key/partition, offset, batching, retry, backoff, DLQ, rebalance, schema và shutdown.

### Level 3

Event là contract giữa bounded contexts. Asynchrony đổi consistency model: caller không có immediate success toàn workflow. Outbox liên kết local state với publication nhưng vẫn thường at-least-once.

## 3. Domain Event khác Integration Event

Domain Event:

~~~go
type MoneyTransferred struct {
	TransferID TransferID
	From       AccountID
	To         AccountID
	Amount     Money
}
~~~

Nó diễn đạt fact trong domain, có thể chứa Value Object.

Integration Event:

~~~go
type moneyTransferredV1 struct {
	EventID      string
	TransferID   string
	AmountMinor  int64
	Currency     string
	OccurredAt   string
	SchemaVersion int
}
~~~

Nó là external compatibility schema, không expose mọi domain field. Adapter/application event mapper redacts, flattens và versions.

Không phải mọi Domain Event đều phải publish; một số chỉ điều phối nội bộ.

## 4. Consumer loop

Conceptual interface độc lập client:

~~~go
type Message struct {
	Key     []byte
	Value   []byte
	Headers map[string][]byte
}

type Handler interface {
	Handle(context.Context, Message) error
}
~~~

Kafka adapter:

1. fetch record;
2. parse envelope/schema;
3. map command;
4. gọi use case;
5. classify outcome;
6. commit/retry/DLQ.

Business logic không nằm trong callback.

## 5. At-least-once và duplicate

Consumer xử lý DB commit rồi crash trước offset commit:

~~~text
process message
→ DB commit
→ crash
→ offset chưa commit
→ record được giao lại
~~~

Duplicate là normal operation. Consumer idempotency:

~~~sql
BEGIN;
INSERT INTO processed_messages(consumer, message_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;
-- nếu inserted: apply business change
COMMIT;
~~~

Processed marker phải cùng transaction với business effect. Check Redis rồi write Postgres không atomic và có race.

## 6. Idempotency state machine

Một boolean “seen” đôi khi chưa đủ. Với operation dài:

~~~text
new → processing → completed
              ↘ failed/retryable
~~~

Phải định nghĩa:

- identity từ producer event ID, không chỉ partition/offset nếu event có thể republish;
- request/event hash;
- concurrent duplicate;
- lease cho processing bị crash;
- retention;
- replay/backfill namespace.

## 7. Offset commit

Auto-commit có thể commit trước business effect. Manual commit sau success thường rõ hơn.

Policy:

| Outcome | Offset |
|---|---|
| success/duplicate completed | commit |
| transient dependency error | chưa commit hoặc retry topic |
| permanent schema/validation | DLQ/quarantine rồi commit |
| process shutdown | dừng fetch, drain, commit completed |

Không commit chỉ vì handler trả nil nếu side effect chạy goroutine chưa kết thúc.

## 8. Ordering

Kafka chỉ bảo đảm order trong một partition. Chọn key AccountID giúp events cùng account vào một partition, nhưng transfer có hai accounts. Không thể đồng thời co-locate mọi cặp account.

Nếu consumer cần version/order:

- event có aggregate ID + sequence;
- detect gap/duplicate;
- buffer/retry hoặc rebuild;
- partition count change và key strategy được quản.

Global ordering làm throughput giảm mạnh và thường không cần.

## 9. Retry

### Blocking retry

Consumer giữ partition, sleep/retry. Dễ nhưng head-of-line blocking.

### Retry topic

Publish sang topic retry với attempt/not-before metadata. Tăng throughput nhưng order với record sau có thể đổi.

### In-process bounded retry

Phù hợp transient ngắn, phải có jitter/budget.

Không retry permanent decode/business rejection. Retry policy cần metric attempt/age, không chỉ count.

## 10. DLQ không phải bãi rác

DLQ record cần:

- original topic/partition/offset;
- event ID/schema version;
- sanitized failure category;
- attempt/timestamps;
- original payload theo data policy.

Phải có owner, alert, inspection, repair và replay tool. Đẩy DLQ rồi quên là data loss có tên đẹp.

## 11. Producer semantics

Kafka producer có retry/batching/idempotent mode tùy client/config. Broker ack không chứng minh downstream consumer applied.

Key quyết định partition/order; header mang trace/schema metadata; topic là adapter config.

Application port nên nói intent:

~~~go
type TransferEvents interface {
	PublishTransferred(context.Context, TransferredEvent) error
}
~~~

Hoặc generic EventPublisher nếu envelope thực sự ổn định. Tránh port nhận kafka.Message.

## 12. Dual write problem

Sai:

~~~text
UPDATE database commit
publish Kafka
~~~

Nếu publish fail, state có nhưng event mất.

Đảo:

~~~text
publish Kafka
UPDATE database
~~~

Nếu DB fail, event nói điều chưa xảy ra.

Không có try/catch nào tạo atomicity giữa hai hệ thống độc lập.

## 13. Transactional Outbox

~~~sql
BEGIN;
UPDATE accounts ...;
INSERT INTO transfers ...;
INSERT INTO outbox(
    event_id, aggregate_id, event_type, payload, occurred_at
) VALUES (...);
COMMIT;
~~~

Worker:

~~~text
claim pending rows
→ publish Kafka
→ mark published
~~~

DB transaction bảo đảm business state và intent-to-publish cùng tồn tại. Worker có thể publish duplicate nếu crash sau publish trước mark; consumer vẫn phải idempotent.

## 14. Claim outbox rows

Nhiều worker có thể dùng:

~~~sql
SELECT id, payload
FROM outbox
WHERE published_at IS NULL
ORDER BY id
FOR UPDATE SKIP LOCKED
LIMIT 100;
~~~

Không giữ transaction mở trong network publish nếu có thể. Một thiết kế claim lease/status rồi commit, publish ngoài tx, sau đó mark. Phải recover lease hết hạn và chấp nhận duplicate.

Alternative CDC/Debezium đọc WAL giảm polling code nhưng thêm platform dependency/operations. Không có free exactly-once.

## 15. Outbox failure matrix

| Failure | Kết quả | Recovery |
|---|---|---|
| Business tx rollback | không state/outbox | retry request |
| Commit OK, worker down | pending durable | restart |
| Publish fail | pending/leased | backoff |
| Publish OK, mark fail | duplicate | idempotent consumer |
| Payload incompatible | stuck/DLQ | repair/upcaster |
| Kafka prolonged down | backlog | capacity/alert/load policy |

Theo dõi oldest pending age, backlog count, attempts và publish latency.

## 16. Schema evolution

JSON, Avro hay Protobuf đều cần compatibility:

- event type/version;
- additive fields/defaults;
- không đổi meaning field cũ;
- consumer tolerant;
- upcaster khi cần;
- contract tests/schema registry;
- retention dài nghĩa event cũ còn xuất hiện.

Domain model có thể đổi nhanh hơn integration contract; mapper là anti-corruption boundary.

## 17. Transaction consumer với database

Handler muốn atomically process + marker:

~~~go
return tx.WithinTransaction(ctx, func(txCtx context.Context) error {
	inserted, err := inbox.TryStart(txCtx, event.ID)
	if err != nil || !inserted {
		return err
	}
	if err := useCase.Execute(txCtx, cmd); err != nil {
		return err
	}
	return inbox.Complete(txCtx, event.ID)
})
~~~

Use case architecture phải tránh nested transaction. Có thể application handler sở hữu transaction, hoặc use case nhận idempotency capability trong cùng UoW.

## 18. Context, rebalance và shutdown

Consumer cần:

1. stop fetch/new messages;
2. cancel/drain handlers theo timeout;
3. commit only completed records;
4. close consumer/producer;
5. flush telemetry.

Rebalance có thể revoke partition khi handler còn chạy. Client library contract quyết định cancellation/commit; test bằng broker thật.

## 19. Observability

Fields/metrics:

- topic/partition/offset (không high-cardinality metric labels tùy tiện);
- consumer group;
- event type/schema;
- event/trace ID;
- processing duration/outcome;
- lag;
- retry/DLQ/outbox age.

Propagate trace context qua headers nhưng không coi trace delivery là business guarantee.

## 20. Testing

- mapper/envelope unit tests;
- fake publisher verifies application intent;
- duplicate consumer test;
- retry classifier and DLQ policy;
- outbox repository integration with PostgreSQL;
- broker integration for partition/offset/rebalance;
- compatibility fixtures for old schemas;
- end-to-end state → outbox → Kafka → consumer.

Mock Kafka client không chứng minh broker ordering/offset.

## 21. Production scenario

Kafka down 45 phút khi Transfer vẫn chạy:

- DB transaction tiếp tục ghi outbox;
- backlog tăng;
- disk/table/index pressure tăng;
- worker backoff, không hammer broker;
- oldest-age alert;
- capacity plan cho outage window;
- consumer nhận burst sau recovery;
- duplicates vẫn có.

Nếu product không chấp nhận delayed notification, application phải expose pending state/SLO; không fake synchronous guarantee.

## 22. Debug

### Duplicate effect

Kiểm event ID, inbox marker cùng transaction, producer republish, handler goroutine/offset commit và retention.

### Missing event

Đi từ business row → outbox row → worker attempts → broker topic/partition → consumer offset → inbox/business effect.

### Ordering bug

Kiểm key, partition count, multiple topics, retry topic reorder, sequence/gap handling.

## 23. Khi nào không dùng Kafka?

Request-response đơn giản, traffic nhỏ, consistency synchronous hoặc team không vận hành broker tốt có thể dùng HTTP/DB job. Kafka thêm operational/data-contract cost. Outbox polling với DB hoặc direct call đôi khi đủ.

## 24. Lab và bài tập

Làm [Lab 09: Kafka](../labs/lab-09-kafka/README.md):

1. consumer mapper;
2. duplicate-safe inbox;
3. retry/permanent classification;
4. fake publisher;
5. outbox failure simulation.

## 25. Mastery questions

1. Kafka consumer là adapter phía nào?
2. Domain Event khác Integration Event?
3. At-least-once tạo duplicate thế nào?
4. Offset commit trước DB gây gì?
5. Outbox giải dual write nhưng còn duplicate ở đâu?
6. Retry topic phá order thế nào?
7. DLQ cần operation nào?
8. Kafka ack chứng minh guarantee gì?
9. Event key cho transfer hai account chọn thế nào?
10. Khi nào Kafka là over-engineering?

## Further reading

- Apache Kafka documentation: design, producer/consumer configs, delivery semantics.
- Confluent/Schema Registry compatibility documentation.
- Martin Fowler, Event-Driven Architecture và Event Sourcing distinctions.
- Transactional Outbox pattern và CDC literature.

## Quality gate

- [x] Producer/consumer dependency directions
- [x] Domain vs integration events
- [x] Duplicate, idempotency, offset, ordering, retry, DLQ
- [x] Full outbox flow/failure matrix
- [x] Schema, lifecycle, observability, tests
- [x] Production scenario/debug/trade-off
- [x] Lab, exercises, mastery
