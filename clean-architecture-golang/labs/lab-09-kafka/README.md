# Lab 09: Idempotent Kafka Consumer và Outbox Worker

Thời lượng: 90-150 phút. Lab mô hình hóa Kafka record/client qua ports để chạy offline; challenge thêm broker integration.

## Mục tiêu

- map integration event sang use case;
- classify permanent/retryable error;
- ngăn duplicate effect bằng inbox;
- chỉ mark outbox sau publish;
- hiểu vì sao vẫn at-least-once.

## Kiến thức cần

- [Kafka Event Driven](../../14-kafka-event-driven/README.md)
- transaction/idempotency và test doubles.

## Diagram

~~~mermaid
flowchart LR
    K["Kafka record"] --> C["Consumer Adapter"]
    C --> I["Inbox ProcessOnce"]
    I --> U["ApplyTransfer Use Case"]
    DB[("Outbox")] --> W["Worker"]
    W --> P["Publisher Port"]
    P --> K
~~~

## Problem

Starter decode rồi gọi use case mỗi delivery. Cùng message ID được nhận hai lần sẽ apply hai lần; malformed payload cũng không được phân loại.

## Yêu cầu

1. Envelope có event_id/type/version/payload.
2. Strict decode và validate.
3. Malformed/unsupported version là PermanentError.
4. Inbox ProcessOnce chỉ đánh completed khi callback success.
5. Duplicate completed trả success mà không apply lại.
6. Retryable failure không mark, delivery sau được chạy lại.
7. Outbox Worker mark published chỉ sau Publisher success.
8. Tests cover duplicate/retry/permanent/publish failure.

## Các bước

1. Tái hiện duplicate ở starter.
2. Tách decoder khỏi use case.
3. Tạo Inbox port với atomic ProcessOnce semantics.
4. Implement memory inbox cho test.
5. Tạo PermanentError + classifier.
6. Tạo outbox Repository/Publisher ports.
7. Test crash window publish-success/mark-fail bằng reasoning.

## Expected behavior

Cùng event ID deliver hai lần chỉ tăng apply count một. Callback transient fail lần đầu được gọi lại lần hai. JSON lỗi không retry vô hạn. Publish lỗi giữ outbox pending.

## Test

~~~bash
cd starter && go test ./...
cd ../solution && go test -race ./... && go vet ./...
~~~

## Questions

1. Inbox marker phải cùng transaction với effect vì sao?
2. Outbox vẫn tạo duplicate ở crash window nào?
3. Offset commit diễn ra trước/ sau Handle ảnh hưởng gì?
4. Retry topic đổi ordering ra sao?
5. Domain Event khác envelope này thế nào?

## Challenge

- Implement inbox/outbox bằng PostgreSQL transaction.
- Dùng Kafka/Testcontainers test partition/offset.
- Thêm lease/attempt/DLQ metadata.
- Thêm schema v2 và compatibility fixtures.

## Solution explanation

Solution tập trung vào semantics có thể unit test độc lập client. Memory Inbox serialize ProcessOnce nhưng chỉ là teaching adapter; production cần marker và business effect trong cùng database transaction. Outbox Worker cố ý giữ publisher/repository ports nhỏ và không tuyên bố exactly-once.
