# Case Study 06: Kafka Worker

Kafka Worker giúp học asynchronous delivery adapter.

Trọng tâm:

- Consumer map event sang command.
- Idempotent processing.
- Retry, backoff, DLQ.
- Offset commit.
- Outbox cho publish event.

Kết luận chính: Kafka consumer không nên trở thành god service; nó là adapter gọi use case.
