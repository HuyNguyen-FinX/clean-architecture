# Lab 09: Kafka

## Mục tiêu

Thêm Kafka consumer/producer adapter mà không sửa domain.

## Yêu cầu

- Consumer parse message và gọi use case.
- Producer implement event publisher port.
- Thiết kế idempotency key.
- Phân tích retry và DLQ.

## Câu hỏi

- Kafka consumer thuộc delivery hay infrastructure?
- Event schema có phải domain entity không?
- Publish event cùng transaction DB xử lý thế nào?

## Mastery Check

- [ ] Tôi biết Kafka adapter map message sang command.
- [ ] Tôi biết dùng outbox khi cần atomicity DB + publish.
- [ ] Tôi biết consumer phải idempotent.
