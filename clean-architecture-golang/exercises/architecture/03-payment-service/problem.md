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

## Assumptions Bắt Buộc

Provider hỗ trợ idempotency key với retention hữu hạn, webhook có thể duplicate/out-of-order và một Payment hỗ trợ partial capture/refund. Hệ thống không được lưu PAN/CVV.

Phân biệt rõ business decline, transient unavailable và unknown outcome. Không được dùng một lỗi `provider failed` cho cả ba.

## Failure Injection

- Timeout trước khi request rời process và timeout sau khi provider nhận request.
- Provider authorize thành công nhưng local result transaction rollback.
- Webhook success đến trước HTTP response decline.
- Hai refund đồng thời vượt captured amount.
- Webhook signature hợp lệ nhưng event ID đã xử lý.

## Deliverables

1. Payment/operation state machine và invariant amount.
2. Gateway contract có authorize/capture/refund/inquiry semantics.
3. Idempotency schema gồm scope, hash, local/provider references.
4. Webhook trust-boundary flow và replay protection.
5. Crash matrix quanh external call và local commits.
6. Reconciliation use case/runbook.
7. Tests cho domain, adapter contract, concurrency, security/log redaction.

## Self-review

- Có chỗ nào biến timeout thành declined không?
- Retry có giữ nguyên operation ID không?
- Raw provider payload có lọt vào domain hoặc logs không?
- Operator có thể giải thích timeline mà không sửa DB trực tiếp không?
