# Bài 2: Thiết Kế Order Service

## Requirement

Order Service xử lý:

- Create order từ cart.
- Reserve inventory.
- Calculate total.
- Apply coupon.
- Confirm order.
- Cancel order.

## Nhiệm vụ

1. Xác định Aggregate Root.
2. Tách Domain Service nếu cần.
3. Thiết kế repository port.
4. Thiết kế gateway gọi Inventory Service.
5. Phân tích transaction và eventual consistency.

## Câu hỏi

- Coupon rule thuộc Order domain hay external service?
- Reserve inventory có nên nằm trong DB transaction của order không?
- Event nào cần publish?
