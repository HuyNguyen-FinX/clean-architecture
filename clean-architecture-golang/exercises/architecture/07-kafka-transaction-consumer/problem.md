# Bài 7: Kafka Consumer Xử Lý Transaction

## Requirement

Consumer đọc Kafka event `TransferRequested` và tạo banking transfer.

## Nhiệm vụ

1. Xác định consumer thuộc layer nào.
2. Map event payload sang command.
3. Thiết kế idempotency.
4. Thiết kế retry và DLQ.
5. Đảm bảo offset commit đúng thời điểm.

## Câu hỏi

- Kafka message có phải use case command không?
- Nếu message bị xử lý hai lần thì sao?
- Commit offset trước hay sau khi transaction DB commit?

## Constraints

Topic 24 partitions, key theo source Account, nhiều consumer replicas. Event có stable `event_id`; Kafka chỉ cung cấp at-least-once cho flow này. Transfer use case hiện đã có transaction nên bạn phải tránh transaction nesting không chủ đích.

## Failure Injection

- Process chết sau DB commit nhưng trước offset commit.
- Offset 12 xử lý xong trước offset 11 trong cùng partition.
- Payload schema v2 tới consumer chỉ hiểu v1.
- Deadlock PostgreSQL, broker rebalance và SIGTERM cùng lúc.
- DLQ publish lỗi cho một poison message.

## Deliverables

1. Consumer adapter mapping và error classification.
2. Inbox/idempotency schema cùng exact local transaction boundary.
3. Ack/commit state table cho success, duplicate, permanent, transient.
4. Lựa chọn sequential/parallel partition và contiguous offset rule.
5. Retry/backoff/DLQ/replay policy có owner.
6. Graceful shutdown sequence.
7. Tests dùng broker thật cho crash/rebalance và DB thật cho atomicity.

## Self-review

- Có commit offset trước durable effect không?
- Redis dedup có bị tách transaction khỏi PostgreSQL effect không?
- Goroutine có return callback trước khi work xong không?
- DLQ có trở thành nơi chôn lỗi không ai sở hữu không?
