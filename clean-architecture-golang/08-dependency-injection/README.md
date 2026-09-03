# 08 Dependency Injection

## Tại sao cần học?

Dependency Injection trong Go thường đơn giản hơn nhiều so với các framework DI. Constructor injection và Composition Root đủ cho phần lớn service.

## Mental Model

Object không tự đi tìm dependency. Dependency được truyền từ bên ngoài vào:

```go
func NewTransferMoneyUseCase(repo AccountRepository, tx Transactor) *TransferMoneyUseCase
```

## Composition Root

`cmd/api/main.go` thường là nơi:

- Load config.
- Mở database pool.
- Tạo repository adapter.
- Tạo use case.
- Tạo handler.
- Start server.

## Wire và Fx

Wire hoặc Fx có thể hữu ích khi object graph rất lớn. Không nên dùng DI framework chỉ để né vài dòng constructor trong service nhỏ.

## Anti-pattern

- Global singleton.
- Service tự đọc env.
- Use case tự mở database connection.
- Interface được tạo chỉ để phục vụ container.

## Mastery Check

- [ ] Tôi biết lắp dependency bằng constructor injection.
- [ ] Tôi biết `main.go` có thể biết detail vì nó là composition root.
- [ ] Tôi biết khi nào DI framework đáng dùng.
