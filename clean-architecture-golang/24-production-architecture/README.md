# 24 Production Architecture

## Tại sao cần học?

Tutorial thường dừng ở handler-usecase-repository. Production cần timeout, shutdown, config, migration, health check, observability, retry và operational safety.

## Concerns

- Graceful shutdown.
- Database pool.
- Timeout.
- Retry/backoff.
- Circuit breaker.
- Idempotency.
- Rate limiting.
- Config và secrets.
- Health, readiness, liveness.
- Migration.
- Observability.

## Layer Placement

Config được đọc ở composition root, truyền vào infrastructure. Domain không đọc env. Health check nằm ở delivery/infra boundary. Retry policy có thể nằm ở gateway adapter hoặc application tùy semantics.

## Anti-pattern

- Domain đọc env.
- Global config mutable.
- Retry mọi lỗi như nhau.
- Không timeout external calls.
- Health check chỉ trả OK mà không kiểm tra dependency cần thiết.

## Mastery Check

- [ ] Tôi biết production concern thuộc layer nào.
- [ ] Tôi biết timeout và cancellation đi qua context.
- [ ] Tôi biết Clean Architecture không thay thế reliability patterns.
