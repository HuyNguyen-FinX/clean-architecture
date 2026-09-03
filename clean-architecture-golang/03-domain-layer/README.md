# 03 Domain Layer

## Tại sao cần học?

Domain Layer là nơi business rules sống lâu nhất. Nếu domain chỉ là struct chứa field, rule sẽ bị đẩy ra service, handler hoặc repository và rất dễ bị phá bởi caller khác.

## Vấn đề

Banking account không nên cho phép code bên ngoài tùy tiện làm:

```go
account.Balance -= amount
```

Thay vào đó, thay đổi state phải đi qua behavior:

```go
err := account.Withdraw(amount)
```

## Nội dung trọng tâm

- Entity và identity.
- Value Object và equality theo giá trị.
- Aggregate và Aggregate Root.
- Domain invariant.
- Domain Service khi rule không thuộc tự nhiên về một entity.
- Domain Event khi domain muốn ghi nhận điều đã xảy ra.

## Dependency

Domain không phụ thuộc HTTP, PostgreSQL, Kafka, Redis, logger framework hoặc config. Domain có thể dùng standard library khi phù hợp, ví dụ `errors`, `time`, `strings`, nhưng cần tránh kéo detail kỹ thuật vào model.

## Anti-pattern

- Anemic Domain Model.
- Public field làm caller phá invariant.
- `json`, `db`, `gorm` tag trong entity khi domain không cần biết transport/persistence.
- Domain method nhận `context.Context`.

## Mastery Check

- [ ] Tôi biết đặt business rule vào entity hoặc domain service.
- [ ] Tôi phân biệt Entity và Value Object.
- [ ] Tôi biết khi nào tách aggregate là cần thiết và khi nào là quá mức.
