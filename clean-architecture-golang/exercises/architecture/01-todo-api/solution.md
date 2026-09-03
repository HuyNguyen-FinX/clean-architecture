# Solution: Todo API

Todo API đơn giản thường không cần chia quá nhiều layer. Một cấu trúc đủ tốt:

```text
cmd/api
internal/todo
  handler.go
  service.go
  repository.go
  model.go
```

Nếu domain chỉ có CRUD, tách `domain/application/infrastructure/delivery` đầy đủ có thể là over-engineering.

Tuy nhiên `MarkCompleted` có thể là behavior nếu status transition có rule:

```text
Archived Todo không được completed.
Completed Todo không được đổi due date.
```

Khi rule tăng, bắt đầu tách domain object và use case. Architecture nên tăng theo complexity, không đi trước quá xa.
