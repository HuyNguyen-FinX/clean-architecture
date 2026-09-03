# 15 Redis Cache

## Tại sao cần học?

Redis thường được dùng cho cache, distributed lock, rate limit và ephemeral state. Nó là infrastructure detail nhưng quyết định đặt cache ở đâu có ảnh hưởng trực tiếp đến consistency và boundary.

## Cache Nên Nằm Ở Đâu?

Không có một câu trả lời duy nhất:

- Repository decorator hợp khi cache che persistence read.
- Use case hợp khi cache là một phần workflow application.
- Delivery middleware hợp với rate limit hoặc response cache đơn giản.

## Nội dung trọng tâm

- Cache aside.
- TTL.
- Cache invalidation.
- Distributed lock.
- Rate limiting.

## Anti-pattern

- Domain đọc Redis trực tiếp.
- Cache key rò vào use case khi không cần.
- Dùng distributed lock thay thế transaction local một cách mơ hồ.

## Mastery Check

- [ ] Tôi biết Redis là adapter detail.
- [ ] Tôi biết phân tích cache placement theo use case.
- [ ] Tôi biết cache có thể làm consistency phức tạp hơn.
