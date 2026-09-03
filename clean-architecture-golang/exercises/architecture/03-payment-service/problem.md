# Bài 3: Thiết Kế Payment Service

## Requirement

Payment Service:

- Authorize payment.
- Capture payment.
- Refund payment.
- Nhận webhook từ provider.
- Đảm bảo idempotency.

## Nhiệm vụ

1. Thiết kế `Payment` aggregate.
2. Thiết kế `PaymentGateway` port.
3. Xác định webhook adapter thuộc layer nào.
4. Thiết kế idempotency key.
5. Phân tích error ambiguous từ provider.

## Câu hỏi

- Provider SDK có được import trong use case không?
- Webhook payload có phải domain event không?
- Capture retry thế nào để không charge hai lần?
