# Lab 10: Redis

## Mục tiêu

Thêm Redis cho cache hoặc rate limiting mà vẫn giữ boundary rõ.

## Yêu cầu

- Chọn một use case cache read hoặc rate limit transfer.
- Đặt Redis adapter đúng layer.
- Thiết kế key format không rò vào domain.
- Phân tích invalidation.

## Câu hỏi

- Cache nên nằm ở repository, use case hay middleware?
- Distributed lock có giải quyết được double spending không?
- TTL ảnh hưởng consistency thế nào?

## Mastery Check

- [ ] Tôi biết Redis là infrastructure detail.
- [ ] Tôi biết cache placement theo mục tiêu.
- [ ] Tôi biết lock không thay thế transaction.
