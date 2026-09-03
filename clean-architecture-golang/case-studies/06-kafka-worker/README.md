# Case Study 06: Kafka Worker - At-Least-Once, Inbox Và Offset

Kafka consumer là driving adapter. Nó quản lý protocol Kafka rồi gọi application use case; nó không nên trở thành nơi chứa SQL, domain rules, retry business và publish logic trong một callback lớn.

## Bối Cảnh

Consumer nhận `TransferRequested.v1`, tạo internal banking transfer và phát `MoneyTransferred.v1`. Topic có 24 partitions, key là source account ID; deployment có nhiều replicas trong consumer group. Delivery thực tế là at-least-once khi process crash hoặc rebalance.

Yêu cầu:

- Duplicate input không tạo duplicate transfer.
- Không commit offset trước durable business result.
- Poison message không block partition vô hạn.
- Shutdown dừng fetch, chờ in-flight có giới hạn, rồi commit/close đúng thứ tự.
- Schema evolution không làm panic toàn worker.

## Boundary Và Dependency

~~~mermaid
flowchart LR
    KAFKA["Kafka broker"] --> CON["Consumer adapter"]
    CON --> DEC["Schema decoder + mapper"]
    DEC --> UC["HandleTransferRequested"]
    UC --> DOMAIN["Transfer domain"]
    UC --> UOW["UnitOfWork port"]
    PG["PostgreSQL adapter"] -.implements.-> UOW
    RELAY["Outbox relay"] --> KAFKA
~~~

Consumer package import Kafka client + application. Application import domain + owned ports. Domain không biết topic, partition, offset, headers hoặc retry count. Kafka event DTO không đi thẳng vào Entity.

## Handler Contract

~~~go
type Handler interface {
	Handle(ctx context.Context, cmd TransferCommand) error
}

type TransferCommand struct {
	MessageID string
	FromID   string
	ToID     string
	Amount   int64
	Currency string
}
~~~

Adapter validate envelope/schema, map DTO -> command, set deadline và classify result. `MessageID` là stable producer identity, không dùng partition-offset làm business id nếu producer có thể republish sang topic khác.

## Atomic Processing Với Inbox

Transaction application:

~~~text
BEGIN
  INSERT inbox(consumer, message_id, payload_hash)
    ON CONFLICT ...
  nếu duplicate cùng hash: return success
  nếu duplicate khác hash: permanent conflict
  apply business use case
  INSERT outbox(MoneyTransferred)
  mark inbox processed
COMMIT
~~~

Sau commit thành công, adapter mới cho phép offset tiến. Nếu process chết sau DB commit nhưng trước offset commit, message được giao lại và inbox biến nó thành no-op thành công.

DB transaction và Kafka offset không cần atomic cùng nhau để đạt effectively-once business effect, miễn inbox bền và business side effect nằm cùng transaction. External network side effect lại cần operation idempotency/reconciliation riêng.

## Offset Và Parallelism

Trong một partition, offset commit phải là prefix liên tục. Nếu xử lý offset 12 song song xong trước 11 rồi commit 12, crash có thể bỏ mất 11. Có ba chiến lược:

- Một worker tuần tự mỗi partition: đơn giản, giữ ordering, throughput theo partition.
- Parallel nhưng track completed gap và chỉ commit contiguous high-water mark.
- Tăng partitions/key distribution thay vì concurrency vô hạn trong cùng partition.

Không spawn goroutine rồi return callback ngay; consumer library có thể coi message đã xử lý trong khi goroutine còn chạy, mất cancellation và tạo unbounded work.

## Retry Classification

| Loại | Ví dụ | Hành vi |
|---|---|---|
| Permanent input | schema sai, currency không hỗ trợ | DLQ/quarantine có reason |
| Business rejection | insufficient balance | record outcome; thường ack, không retry vô hạn |
| Transient dependency | DB connection unavailable | retry/backoff, không commit offset |
| Concurrency transient | deadlock/serialization | retry bounded toàn transaction |
| Unknown bug | panic/invariant impossible | recover tại worker boundary, alert, retry budget rồi quarantine |

Retry tight loop giữ partition nóng và đập dependency đang lỗi. Dùng exponential backoff có jitter, max attempts/time budget và quan sát age. Retry topic giúp giải phóng partition nhưng làm ordering phức tạp; đây là trade-off business.

## DLQ Không Phải Thùng Rác

DLQ record cần original topic/partition/offset, key, headers cần thiết, schema version, failure classification, attempts, first/last failure time và sanitized payload/ref. Có runbook inspect, fix, replay với authorization. Không đưa secret/PII nguyên bản nếu policy cấm.

Một message permanent không được retry hàng giờ trước khi DLQ. Một outage diện rộng không nên đẩy hàng triệu message hợp lệ vào DLQ; circuit/pause consumer phù hợp hơn.

## Outbox Chiều Ra

Use case không gọi Kafka producer trong transaction. Nó ghi business state + outbox. Relay select batch bằng `FOR UPDATE SKIP LOCKED`, publish, rồi mark. Crash sau publish trước mark tạo duplicate; downstream vẫn phải idempotent.

Outbox payload là integration event versioned, được map từ domain/application fact. Không serialize nguyên Aggregate vì làm lộ field nội bộ và khóa schema vào implementation.

## Rebalance Và Shutdown

Khi partition bị revoke, worker ngừng nhận item mới của partition đó, cancel/chờ in-flight theo library contract và chỉ commit offsets đã hoàn tất. Shutdown sequence:

1. Cancel fetch loop/readiness false.
2. Chờ handlers trong grace period.
3. Flush/commit contiguous offsets theo contract.
4. Close consumer; để orchestration kill nếu vượt deadline.

`context.Context` từ worker có deadline vào application/repository. Domain methods không nhận context.

## Failure Walkthrough

| Crash point | Khi restart |
|---|---|
| Trước inbox insert | xử lý lại bình thường |
| Sau inbox insert nhưng tx rollback | row không tồn tại, xử lý lại |
| Sau business commit trước offset | inbox nhận duplicate và skip |
| Sau offset trước DB commit | đây là data loss; flow phải cấm |
| Outbox publish trước mark | publish lại, downstream deduplicate |
| Rebalance giữa handler | context/callback contract quyết định; không commit work dở |

## Testing Strategy

- Decoder contract tests theo fixtures v1/v2, missing/unknown fields và corrupt bytes.
- Application test duplicate cùng hash/khác hash, rollback và business rejection.
- PostgreSQL integration cho inbox unique, business + outbox atomicity.
- Consumer adapter test rằng ack chỉ xảy ra sau success/permanent classification.
- Concurrency test contiguous-offset tracker với completion đảo thứ tự.
- Broker integration test rebalance, process kill và replay.
- Load test partition skew/hot key, lag và backpressure.

## Observability

Metrics: consumer lag, processing latency, retry count, age of oldest retry, DLQ rate, rebalance count, inbox duplicate, outbox lag. Topic/consumer group là bounded labels; message ID chỉ ở trace/log. Log quyết định ack/retry/DLQ với classification, không log payload mặc định.

## Khi Nào Đơn Giản Hơn

Worker analytics idempotent có thể chỉ ghi upsert và commit offset, không cần generic inbox. Nếu business side effect tự nhiên commutative/idempotent, tận dụng nó. Với throughput nhỏ, xử lý tuần tự mỗi partition dễ vận hành hơn worker pool phức tạp.

## Câu Hỏi Mastery

1. Vì sao DB commit trước offset commit tạo duplicate nhưng không data loss?
2. Inbox key nên gồm những gì?
3. Retry topic đánh đổi ordering lấy điều gì?
4. Outbox có làm consumer downstream exactly-once không?
5. Một business rejection nên retry hay ack? Context nào thay đổi câu trả lời?

## Bài Thực Hành

Hoàn thành [Lab 09](../../labs/lab-09-kafka/README.md), rồi bổ sung offset tracker cho ba message hoàn thành theo thứ tự 2, 1, 3. Chứng minh bằng test offset 2 không được commit trước offset 1.
