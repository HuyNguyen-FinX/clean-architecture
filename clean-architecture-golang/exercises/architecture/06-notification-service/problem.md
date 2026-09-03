# Bài 6: Thiết Kế Notification Service

## Requirement

Notification Service gửi email, SMS và push notification từ event.

## Nhiệm vụ

1. Xác định domain có giàu rule không.
2. Thiết kế channel gateway.
3. Thiết kế template rendering.
4. Thiết kế retry và DLQ.
5. Phân tích idempotency.

## Câu hỏi

- Notification có cần Clean Architecture đầy đủ không?
- Email provider SDK nằm ở đâu?
- Template có phải domain không?

## Bối Cảnh Bổ Sung

Kafka delivery at-least-once; mỗi provider có quota/error semantics khác. Sản phẩm thêm user preference, quiet hours, locale fallback, regulatory template và unsubscribe. Một event có thể fan-out nhiều channel.

Hãy thiết kế proportional: chỉ nâng domain richness khi rule thực sự tồn tại.

## Failure Injection

- Email provider timeout nhưng có thể đã nhận message.
- Event duplicate sau process restart.
- SMS trả `429 Retry-After`; push token invalid vĩnh viễn.
- Template version bị xóa giữa retry.
- DLQ chứa recipient/body nhạy cảm.

## Deliverables

1. Boundary diagram consumer -> use case -> channel adapters.
2. Idempotency identity cho event/channel/recipient/template version.
3. Channel-specific result/error taxonomy và retry budget.
4. Template/preference ownership decision.
5. Delivery state, DLQ/replay và PII policy.
6. Tests với fake clock, adapter fixtures, duplicate/rate limit.

## Self-review

- Một interface `Sender` có làm mất capability khác biệt giữa SMS/email không?
- Permanent invalid address có retry vô hạn không?
- Log/DLQ có lộ body/token/recipient không?
- Managed provider có thể thay phần lớn service tự xây không?
