# Case Study 02: E-commerce Order

Order Service là ví dụ tốt để học Aggregate, state transition và external gateway.

Trọng tâm:

- `Order` làm Aggregate Root.
- Coupon, inventory reservation và total calculation.
- Gateway sang Inventory/Payment.
- Event `OrderCreated`, `OrderConfirmed`, `OrderCancelled`.
- Saga hoặc eventual consistency khi gọi service ngoài.

Kết luận chính: không giữ local DB transaction mở qua network call; dùng state machine và event để điều phối an toàn hơn.
