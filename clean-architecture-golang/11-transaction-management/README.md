# 11 Transaction Management

## Tại sao cần học?

Transaction là nơi Clean Architecture gặp thực tế production. Use case cần atomicity, repository biết database, còn domain không nên biết transaction driver.

## Câu hỏi chính

Transaction nên bắt đầu ở repository hay use case?

Nếu một use case chỉ save một aggregate, repository tự xử lý có thể đủ. Nếu một use case cần nhiều repository hoặc nhiều aggregate nhất quán, transaction boundary thường nên nằm ở application layer qua một abstraction như `Transactor`.

## Patterns

- Transaction Manager.
- Unit of Work.
- Transactional Repository.
- Context-based transaction.

## Trade-off

Context-based transaction tiện nhưng dễ giấu dependency. Unit of Work rõ boundary nhưng có thể thêm ceremony. Transaction Manager function kiểu `WithinTx` thường hợp với Go vì đơn giản và explicit.

## Anti-pattern

- Mỗi repository tự commit làm flow nhiều bước mất atomicity.
- Domain entity biết `*sql.Tx`.
- Transaction kéo dài qua network call không kiểm soát.
- Retry transaction nhưng không idempotent.

## Mastery Check

- [ ] Tôi biết đặt transaction boundary theo use case.
- [ ] Tôi biết locking và idempotency liên quan nhưng không thay thế transaction.
- [ ] Tôi biết trade-off của `WithinTx`.
