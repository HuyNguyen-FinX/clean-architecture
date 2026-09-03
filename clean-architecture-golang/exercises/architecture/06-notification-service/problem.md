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
