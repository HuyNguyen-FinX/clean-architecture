# Bài 5: Thiết Kế Banking Transfer

## Requirement

A chuyển 500.000 VND cho B.

Phải xử lý:

- Balance invariant.
- Transaction.
- Row locking.
- Idempotency.
- Retry.
- Double spending.
- Transaction history.

## Nhiệm vụ

1. Thiết kế `Account`, `Money`, `Transfer`.
2. Thiết kế `TransferMoneyUseCase`.
3. Thiết kế repository port.
4. Đặt transaction boundary.
5. Phân tích Kafka event nếu cần publish `MoneyTransferred`.

## Câu hỏi

- Locking thuộc domain hay repository adapter?
- Idempotency thuộc layer nào?
- Clean Architecture giúp gì và không giúp gì?
