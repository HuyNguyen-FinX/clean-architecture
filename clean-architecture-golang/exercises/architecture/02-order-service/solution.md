# Solution: Order Service

`Order` thường là Aggregate Root. `OrderItem`, status transition và total calculation có thể nằm trong domain nếu rule ổn định.

Gateway:

```go
type InventoryGateway interface {
	Reserve(ctx context.Context, items []ReserveItem) error
}
```

Use case `CreateOrder` orchestration:

```text
validate command
create order aggregate
apply coupon policy
save order as pending
reserve inventory
confirm or mark failed
publish event
```

Không nên giữ DB transaction local mở trong lúc gọi Inventory Service qua network. Cần phân tích Saga hoặc state machine nếu consistency phức tạp.
