# Solution: God Handler

## Findings

- Handler chứa quá nhiều concern: HTTP parsing, SQL, domain validation, state mutation, Kafka và response mapping.
- Business invariant nằm ở nơi không tái sử dụng được nếu thêm gRPC hoặc Kafka consumer.
- Transaction boundary không rõ.
- Kafka publish trong cùng flow có thể tạo inconsistent state nếu DB commit thành công nhưng publish fail.
- Test business rule buộc phải dựng HTTP/database nếu không tách.

## Refactor

```text
Step 1: Viết characterization test cho behavior hiện tại.
Step 2: Tách Account.Withdraw và Account.Deposit.
Step 3: Tạo TransferMoneyUseCase.
Step 4: Tạo AccountRepository port.
Step 5: Đưa SQL vào postgres adapter.
Step 6: Đưa Kafka publish sau port hoặc outbox.
Step 7: Handler chỉ map HTTP <-> command/response.
```

## Boundary Sau Refactor

```text
HTTP Handler -> TransferMoneyUseCase -> Account domain
PostgresRepository -> implements AccountRepository
KafkaPublisher -> implements EventPublisher
```
