# 10 Database

## Tại sao cần học?

Database là detail kỹ thuật nhưng dữ liệu là tài sản cốt lõi. Clean Architecture không xem nhẹ database; nó chỉ không để database schema điều khiển domain model.

## Nội dung trọng tâm

- `database/sql` và `pgx`.
- Connection pooling.
- Query mapping.
- Migration.
- Repository implementation.
- Integration test với PostgreSQL thật.

## Flow

```text
Use Case
↓
Repository Interface
↑
PostgresRepository
↓
pgx
↓
PostgreSQL
```

## Mapping

DB model có thể có field phục vụ persistence như `created_at`, `updated_at`, nullable column, optimistic lock version. Domain entity chỉ nên giữ thông tin có ý nghĩa nghiệp vụ.

## Anti-pattern

- Domain entity chứa `db` hoặc `gorm` tag khi không cần.
- Repository trả raw row.
- SQL transaction bị mở rải rác trong từng repository method.

## Mastery Check

- [ ] Tôi biết domain không import driver database.
- [ ] Tôi biết mapping row sang aggregate ở adapter.
- [ ] Tôi biết connection pool thuộc infrastructure/composition root.
