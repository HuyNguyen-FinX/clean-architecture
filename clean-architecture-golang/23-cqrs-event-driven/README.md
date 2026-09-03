# 23 CQRS Event Driven

## Tại sao cần học?

CQRS và Event Driven Architecture hữu ích khi read/write model khác nhau, workflow asynchronous hoặc scale requirements buộc tách processing. Chúng không phải bước đầu tiên của Clean Architecture.

## Patterns

- CQRS.
- Domain Events.
- Event Sourcing.
- Outbox Pattern.
- Saga.
- Specification Pattern.

## Trade-off

CQRS có thể làm read side đơn giản hơn nhưng làm consistency và mental model phức tạp hơn. Event Sourcing cho audit trail mạnh nhưng tăng chi phí modeling, migration và debugging.

## Anti-pattern

- Dùng event sourcing chỉ vì nghe hiện đại.
- Tách command/query cho CRUD đơn giản.
- Event schema thay domain language.
- Không có idempotency cho consumer.

## Mastery Check

- [ ] Tôi biết CQRS không bắt buộc trong Clean Architecture.
- [ ] Tôi biết Outbox giải quyết atomicity giữa DB và message publish.
- [ ] Tôi biết eventual consistency cần product-level expectation.
