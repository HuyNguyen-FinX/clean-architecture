# 04 Use Case / Application Layer

## Tại sao cần học?

Application Layer là nơi orchestration một hành động có ý nghĩa với hệ thống: transfer money, approve loan, capture payment, create order. Nó nối domain với các port cần thiết nhưng không nên biết chi tiết HTTP hoặc SQL.

## Vấn đề

Nếu handler vừa parse request, vừa validate nghiệp vụ, vừa query database, vừa publish Kafka, use case thật sự bị mất. Khi đó rất khó thêm gRPC, worker hoặc test workflow độc lập.

## Nội dung trọng tâm

- Use Case vs Application Service.
- Orchestration vs business rule.
- Input command và output model.
- Transaction boundary.
- Idempotency ở application workflow.

## Dependency

```text
application -> domain
application -> port interface do application định nghĩa
```

Application không import delivery adapter hoặc infrastructure adapter.

## Go Implementation

```go
type TransferMoneyUseCase struct {
	accounts AccountRepository
	tx       Transactor
}
```

Constructor injection giúp `cmd/api/main.go` lắp implementation thật vào use case.

## Mastery Check

- [ ] Tôi biết use case orchestration flow nhưng không chứa SQL.
- [ ] Tôi biết domain rule khác application workflow.
- [ ] Tôi biết transaction nhiều repository thường thuộc application boundary.
