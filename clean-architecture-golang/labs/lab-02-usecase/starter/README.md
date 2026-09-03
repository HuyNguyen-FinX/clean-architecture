# Starter

Bắt đầu từ domain của lab 01. Thêm một file use case nhận command thuần Go:

```go
type TransferMoneyCommand struct {
	FromAccountID string
	ToAccountID   string
	Amount        int64
	Currency      string
}
```

Starter nên dùng fake repository trong test để bạn tập consumer-defined interface.
