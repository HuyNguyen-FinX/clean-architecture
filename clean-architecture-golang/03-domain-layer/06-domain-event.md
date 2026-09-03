# Domain Event

## Domain Event Là Fact Trong Domain

Domain Event diễn đạt điều có ý nghĩa nghiệp vụ **đã xảy ra**:

```text
AccountFrozen
MoneyWithdrawn
TransferCompleted
OrderConfirmed
```

Tên thường ở past tense vì event là fact, không phải request. `TransferMoney` là command; `MoneyTransferred` là event sau success.

## Ba Level

### Level 1: Trực Giác

Command nói “hãy làm”. Event nói “đã xảy ra”. Event cho phần khác phản ứng mà object phát event không cần biết từng listener.

### Level 2: Backend Engineer

Domain Entity/Aggregate có thể record event sau valid transition. Application lấy events, lưu/publish theo transaction policy. Kafka adapter map event sang versioned integration message.

### Level 3: Architecture

Phải tách ba khái niệm:

| Khái niệm | Ownership | Guarantee |
|---|---|---|
| Domain Event | Domain fact/model | Fact được tạo đúng theo transition |
| Application event/action | Workflow nội bộ | Handler/use case phối hợp side effect |
| Integration Event | Contract giữa process/service | Versioning, serialization, delivery semantics |

Một Domain Event không tự có durability. Một Kafka message không tự là Domain Event.

## Domain Event Vs Command

```go
type FreezeAccountCommand struct {
	AccountID string
	Reason    string
}

type AccountFrozen struct {
	AccountID  AccountID
	Reason     FreezeReason
	OccurredAt time.Time
}
```

Command có thể bị reject. Event chỉ được tạo sau khi transition thành công.

Không đặt tên event `FreezeAccountEvent`; nó nghe như command và làm semantics mơ hồ.

## Record Event Trong Aggregate

Conceptual code, chưa được thêm vào mini-banking V1-V3:

```go
type Account struct {
	id     AccountID
	status AccountStatus
	events []Event
}

func (a *Account) Freeze(reason FreezeReason, at time.Time) error {
	if a.status == AccountStatusClosed {
		return ErrInvalidAccountTransition
	}
	if a.status == AccountStatusFrozen {
		return nil
	}

	a.status = AccountStatusFrozen
	a.events = append(a.events, AccountFrozen{
		AccountID:  a.id,
		Reason:     reason,
		OccurredAt: at,
	})
	return nil
}
```

Các quyết định cần explicit:

- Idempotent call có tạo event lần nữa không? Ví dụ trên không.
- `time.Time` đến từ caller/Clock để test deterministic hay gọi `time.Now()` trong domain?
- Events được clear khi nào?
- Clone/rehydration có mang pending events không? Thường rehydration không tái-phát historical events.

## Pull Events Và Rủi Ro Mất Event

```go
func (a *Account) PullEvents() []Event {
	events := append([]Event(nil), a.events...)
	a.events = nil
	return events
}
```

Nếu application pull/clear rồi publish Kafka fail, event có thể mất trong memory. Nếu publish trước DB commit, consumer có thể thấy event cho state bị rollback.

Domain API chỉ quản lý collection; reliability là application/infrastructure problem.

## Domain Event Không Phải Kafka Message

Domain event dùng rich types/private model:

```go
type MoneyTransferred struct {
	TransferID TransferID
	Amount     Money
	OccurredAt time.Time
}
```

Integration event cần stable schema:

```go
type moneyTransferredV1 struct {
	EventID      string `json:"event_id"`
	TransferID   string `json:"transfer_id"`
	AmountMinor  int64  `json:"amount_minor"`
	Currency     string `json:"currency"`
	OccurredAt   string `json:"occurred_at"`
	SchemaVersion int    `json:"schema_version"`
}
```

Adapter/application mapper quyết định:

- Event ID/idempotency key.
- JSON/Avro/Protobuf schema.
- Topic, key, headers và partitioning.
- Backward/forward compatibility.
- PII redaction.

Domain không import Kafka client hoặc biết topic.

## Outbox: Nối DB State Với Publish Intent

Yêu cầu:

```text
Account update thành công thì MoneyTransferred cuối cùng phải được publish.
```

Publish trực tiếp trong DB transaction nguy hiểm vì Kafka không tham gia local PostgreSQL atomicity.

Outbox flow:

```text
BEGIN
UPDATE accounts
INSERT transfers
INSERT outbox(event_id, type, payload)
COMMIT

worker đọc outbox
publish Kafka
mark published
```

DB state và publish intent commit atomically. Worker vẫn có thể publish rồi crash trước mark, nên consumer phải chịu duplicate/idempotent. Outbox thường cho at-least-once, không tạo exactly-once end-to-end.

## Event Handler Nằm Ở Đâu?

Nếu reaction là local application workflow:

```text
AccountFrozen -> invalidate cards
```

Handler có thể là application component gọi ports. Nếu reaction cần reliability xuyên process, integration adapter/outbox tham gia.

Đừng để Entity gọi email/Kafka trực tiếp:

```go
func (a *Account) Freeze() error {
	a.kafka.Publish(...)
	return nil
}
```

Domain transition bị network failure, retry và SDK dependency chi phối.

## Event Granularity

Quá ít event: downstream phải query/đoán state changes.

Quá nhiều technical event:

```text
AccountBalanceFieldUpdated
DatabaseRowSaved
HTTPResponseCreated
```

chỉ expose implementation. Event nên là domain fact có consumer/use case rõ hoặc audit meaning.

Không tạo event cho mọi setter. Event contract là maintenance cost.

## Failure Matrix

| Failure | Direct publish | Transactional outbox |
|---|---|---|
| Publish fail trước DB commit | Có thể rollback nhưng giữ lock lâu | DB chưa commit outbox, không publish |
| DB rollback sau publish | Phantom event | Outbox rollback cùng state |
| Commit success, process crash | Event có thể mất | Outbox row còn để worker retry |
| Worker publish rồi crash trước mark | Không áp dụng | Có duplicate, cần idempotency |
| Kafka down lâu | Request có thể chậm/fail | Outbox backlog tăng, cần alert/capacity |

## Testing Strategy

Domain test:

- Event chỉ record sau valid transition.
- Invalid/idempotent transition không tạo duplicate event theo policy.
- Event chứa domain facts đúng.

Application test:

- Events được chuyển thành outbox/publisher intent đúng.
- Transaction rollback không để state/outbox partial.

Integration test:

- Outbox row commit cùng aggregate.
- Worker retry/duplicate behavior.
- Integration schema serialization/version.

## Khi Nào Không Nên Dùng Domain Event?

- Không có reaction/audit requirement.
- Flow synchronous đơn giản, direct call rõ hơn.
- Event chỉ để tránh gọi một function trong cùng use case.
- Team chưa có ownership/reliability plan cho event handlers.

Event-driven indirection làm runtime/debug khó hơn. Dùng event khi decoupling theo fact có lợi ích thật.

## Mastery Questions

1. Command khác Domain Event ở semantics nào?
2. Tại sao Domain Event không nên là Kafka DTO?
3. Record event trong aggregate có bảo đảm event được publish không?
4. Outbox giải quyết failure nào và không giải quyết duplicate nào?
5. Event nên được record trước hay sau state validation?
6. Idempotent transition có phát event lại không? Vì sao cần policy explicit?
7. Khi direct application call rõ hơn Domain Event?
