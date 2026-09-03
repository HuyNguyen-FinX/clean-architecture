# Solution: Kafka Consumer Xử Lý Transaction

Kafka consumer là delivery adapter cho asynchronous input. Nó không nên chứa business logic transfer.

Flow:

```text
Kafka message
↓
Consumer adapter
↓
TransferMoneyCommand
↓
TransferMoneyUseCase
↓
Domain + Repository
```

Idempotency key nên đến từ event ID hoặc business operation ID. Commit offset sau khi use case xử lý thành công hoặc sau khi ghi trạng thái đủ để retry an toàn. Với failure không recover được, đưa message sang DLQ kèm cause và metadata.
