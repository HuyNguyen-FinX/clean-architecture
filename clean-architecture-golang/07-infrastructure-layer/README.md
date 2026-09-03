# 07 Infrastructure Layer

## Tại sao cần học?

Infrastructure Layer chứa các detail dễ thay đổi: database driver, Redis client, Kafka library, external API SDK, file storage, email provider. Nó quan trọng trong production nhưng không nên điều khiển shape của domain.

## Nội dung trọng tâm

- Adapter implement port.
- Mapping giữa persistence model và domain.
- Retry, timeout, circuit breaker ở external client.
- Connection pool, migration, health check.
- Infrastructure error wrapping.

## Dependency

```text
infrastructure/postgres -> application/domain
infrastructure/kafka -> application/domain
infrastructure/redis -> application/domain
```

Infrastructure được phép biết core contract để implement. Core không biết infrastructure.

## Anti-pattern

- Domain import `pgx`, Redis client hoặc Kafka message.
- DB row được truyền xuyên use case như domain entity.
- External API response trở thành business model.

## Mastery Check

- [ ] Tôi biết infrastructure là detail nhưng vẫn cần code production-grade.
- [ ] Tôi biết mapping nằm ở adapter.
- [ ] Tôi biết adapter failure không nên rò raw detail vào domain.
