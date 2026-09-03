# Solution

Solution tốt đặt interface gần use case:

```go
type AccountRepository interface {
	FindByID(ctx context.Context, id domain.AccountID) (*domain.Account, error)
	Save(ctx context.Context, account *domain.Account) error
}
```

In-memory repository là adapter, không phải domain. Nó có thể nằm dưới `infrastructure/memory`.
