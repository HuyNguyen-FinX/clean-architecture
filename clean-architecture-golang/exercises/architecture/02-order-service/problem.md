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

## Assumptions Bắt Buộc

Inventory và Payment do hai team khác sở hữu, không dùng chung database. Broker at-least-once. Payment timeout có thể xảy ra sau khi provider đã authorize. Khách cần thấy trạng thái pending thay vì chờ vô hạn.

Bạn phải chọn synchronous orchestration, asynchronous Saga hoặc hybrid; không được chỉ ghi "dùng microservices".

## Failure Injection

- Inventory reserve thành công, Payment declined.
- Payment success nhưng response mất.
- `PaymentAuthorized` bị giao hai lần hoặc đến sau `OrderCancelled`.
- Outbox publish thành công rồi worker chết trước mark.
- Compensation release inventory thất bại quá SLA.

## Deliverables

1. Aggregate boundary và transition table của Order.
2. Ownership của coupon rule và price snapshot.
3. Ports theo business intent, không dùng generic CRUD/gateway.
4. Sequence happy path cùng ít nhất hai compensation paths.
5. Local transaction boundaries, outbox/inbox/idempotency keys.
6. API representation cho pending/failed/manual-repair.
7. Test portfolio và ba metrics/runbook signals.

## Self-review

- Có giữ DB lock qua network call không?
- Event schema có serialize nguyên Aggregate/PII không?
- Duplicate compensation có tạo side effect lần hai không?
- State cũ có thể bị event đến muộn kéo lùi không?
