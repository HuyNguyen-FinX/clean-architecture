# 28 System Design

## Tại sao cần học?

Clean Architecture nằm trong một service hoặc một module. System design nhìn rộng hơn: service boundaries, data ownership, messaging, consistency, scaling và failure modes.

## Clean Architecture != Microservices

Clean Architecture có thể dùng cho:

- Monolith.
- Modular Monolith.
- Microservice.
- CLI.
- Worker.
- Batch job.

Microservices thêm distributed systems complexity. Đừng dùng microservices chỉ để chứng minh architecture.

## Distributed Systems Limitation

Clean Architecture không tự giải quyết:

- Distributed transaction.
- Eventual consistency.
- Network failure.
- Message duplication.
- Kafka ordering.
- Service discovery.
- Split brain.

## Mastery Check

- [ ] Tôi biết bắt đầu từ modular monolith trước khi tách service khi phù hợp.
- [ ] Tôi biết Clean Architecture giải quyết source dependency, không giải quyết mọi vấn đề phân tán.
- [ ] Tôi biết boundary service phải dựa trên domain ownership.
