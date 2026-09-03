# Lab 04: PostgreSQL

## Mục tiêu

Thay in-memory repository bằng PostgreSQL adapter mà không sửa domain và use case.

## Yêu cầu

- Tạo schema account tối thiểu.
- Implement repository bằng `pgx` hoặc `database/sql`.
- Mapping DB row sang domain entity.
- Viết integration test với database thật hoặc container.

## Câu hỏi

- Domain có cần biết table name không?
- SQL error nên được map ở đâu?
- Transaction sẽ ảnh hưởng repository API như thế nào?

## Mastery Check

- [ ] Tôi thay adapter mà use case không đổi.
- [ ] Tôi biết connection pool thuộc composition root/infrastructure.
- [ ] Tôi biết mapping không nằm trong domain.
