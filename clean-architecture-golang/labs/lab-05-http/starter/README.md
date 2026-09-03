# Starter

Bắt đầu với use case từ lab 02/03. Tạo `delivery/http` và expose route `POST /transfers`.

Handler nên nhận interface nhỏ:

```go
type TransferUseCase interface {
	Execute(ctx context.Context, cmd application.TransferMoneyCommand) error
}
```
