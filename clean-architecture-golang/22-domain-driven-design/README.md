# 22 Domain Driven Design

## Tại sao cần học?

DDD giúp mô hình hóa domain phức tạp. Clean Architecture giúp bảo vệ dependency boundary. Hai thứ bổ sung cho nhau nhưng không đồng nghĩa.

## Nội dung trọng tâm

- Entity.
- Value Object.
- Aggregate.
- Bounded Context.
- Domain Service.
- Domain Event.
- Repository.
- Ubiquitous Language.

## Không Giáo Điều

Không biến toàn bộ project thành DDD nếu domain đơn giản. Một CRUD admin nhỏ không cần aggregate phức tạp. Banking, payment, loan, order lifecycle thường đáng đầu tư domain model hơn.

## Dependency

Bounded Context nên có ownership rõ. Shared kernel phải rất cẩn thận vì dễ biến thành coupling toàn cục.

## Mastery Check

- [ ] Tôi biết DDD giải quyết modeling, Clean Architecture giải quyết dependency.
- [ ] Tôi biết xác định Aggregate Root.
- [ ] Tôi biết khi nào DDD là over-engineering.
