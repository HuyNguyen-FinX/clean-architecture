# Roadmap

Roadmap này đi từ nền tảng architecture đến thiết kế production. Mỗi phase có kiến thức, file cần đọc, lab, bài tập và mastery criteria.

## Phase 1: Software Architecture Fundamentals

Kiến thức cần học:

- Software Architecture vs Design.
- High-level policy vs low-level details.
- Coupling, cohesion, abstraction, boundary.
- Source-code dependency khác runtime flow.

File cần đọc:

- `00-software-architecture-foundations/README.md`
- `GLOSSARY.md`
- `CHEATSHEET.md`

Lab:

- `labs/lab-01-simple-domain`

Bài tập:

- Vẽ dependency của một service bạn từng làm.
- Chỉ ra business rule nào đang bị trộn với database hoặc HTTP.

Mastery criteria:

- Giải thích được vì sao architecture không phải là folder tree.
- Phân biệt được policy và detail trong một use case thật.
- Nhận ra source dependency bằng cách đọc import, không chỉ đọc call flow.

## Phase 2: Clean Architecture Fundamentals

Kiến thức cần học:

- Dependency Rule.
- Frameworks & Drivers, Interface Adapters, Application, Domain.
- Ports and Adapters, Hexagonal Architecture, Onion Architecture.
- Vì sao dependency đi vào trong nhưng runtime call có thể đi ra ngoài.

File cần đọc:

- `01-clean-architecture-foundations/README.md`
- `02-dependency-rule/README.md`

Lab:

- `labs/lab-02-usecase`

Mastery criteria:

- Giải thích được `handler -> usecase -> repository interface` và `postgres repository implements interface`.
- Biết vì sao domain không import `pgx`, `net/http`, Kafka client hoặc Redis client.

## Phase 3: Domain Modeling

Kiến thức cần học:

- Entity, Value Object, Aggregate, Aggregate Root.
- Domain invariant, Domain Service, Domain Event.
- Entity khác database model và DTO.

File cần đọc:

- `03-domain-layer/README.md`
- `22-domain-driven-design/README.md`

Lab:

- `labs/lab-01-simple-domain`

Mastery criteria:

- Biết đặt logic như `Withdraw` vào entity thay vì sửa field trực tiếp.
- Biết khi nào tách Value Object là đáng giá, khi nào là quá sớm.

## Phase 4: Use Cases

Kiến thức cần học:

- Application Service.
- Orchestration vs domain rule.
- Input/output model của use case.
- Transaction boundary ở application layer.

File cần đọc:

- `04-usecase-application-layer/README.md`

Lab:

- `labs/lab-02-usecase`

Mastery criteria:

- Thiết kế được `TransferMoneyUseCase` không biết HTTP và không biết PostgreSQL.
- Phân biệt được Domain Service và Application Service.

## Phase 5: Dependency Inversion

Kiến thức cần học:

- Interface nhỏ theo consumer.
- `Accept interfaces, return structs`.
- Repository interface nên nằm ở đâu.
- Khi nào không cần interface.

File cần đọc:

- `02-dependency-rule/README.md`
- `05-repository-pattern/README.md`
- `08-dependency-injection/README.md`

Lab:

- `labs/lab-03-repository`
- `labs/lab-06-dependency-injection`

Mastery criteria:

- Biết tạo port ở nơi use case cần, không tạo `IUserRepository` theo thói quen.
- Biết dùng constructor injection và composition root trong `main.go`.

## Phase 6: Repository + PostgreSQL

Kiến thức cần học:

- Repository vs DAO vs Gateway vs Data Mapper.
- `database/sql`, `pgx`, connection pool.
- Mapping giữa database row và domain entity.

File cần đọc:

- `05-repository-pattern/README.md`
- `10-database/README.md`

Lab:

- `labs/lab-04-postgresql`

Mastery criteria:

- Viết được PostgreSQL adapter implement repository port mà application/domain không import PostgreSQL package.

## Phase 7: HTTP / gRPC

Kiến thức cần học:

- Delivery layer.
- DTO, request validation, response mapping.
- Thêm HTTP và gRPC cùng gọi một use case.

File cần đọc:

- `06-delivery-layer/README.md`
- `12-http-rest-api/README.md`
- `13-grpc/README.md`

Lab:

- `labs/lab-05-http`

Mastery criteria:

- Handler không chứa business rule.
- Thêm delivery adapter mới mà không sửa domain.

## Phase 8: Testing

Kiến thức cần học:

- Domain test không cần mock.
- Use case test với fake repository.
- Repository integration test với database thật.
- HTTP test bằng `httptest`.

File cần đọc:

- `20-testing/README.md`

Lab:

- `labs/lab-07-testing`

Mastery criteria:

- Biết mock ít hơn, fake đúng chỗ hơn.
- Test được use case mà không cần PostgreSQL.

## Phase 9: Transactions

Kiến thức cần học:

- Transaction Manager.
- Unit of Work.
- Transactional Repository.
- Context-based transaction.
- Locking, idempotency, retry.

File cần đọc:

- `11-transaction-management/README.md`

Lab:

- `labs/lab-08-transaction`

Mastery criteria:

- Giải thích được transaction nên bắt đầu ở use case trong các flow nhiều repository.
- Biết trade-off của từng pattern transaction trong Go.

## Phase 10: Kafka / Redis

Kiến thức cần học:

- Kafka consumer/producer là adapter.
- Event schema, retry, DLQ, idempotency.
- Cache, distributed lock, rate limiting.
- Transactional outbox.

File cần đọc:

- `14-kafka-event-driven/README.md`
- `15-redis-cache/README.md`

Lab:

- `labs/lab-09-kafka`
- `labs/lab-10-redis`

Mastery criteria:

- Kafka message không rò vào domain entity.
- Biết cache nên đặt ở repository, use case hoặc decorator tùy mục tiêu.

## Phase 11: DDD

Kiến thức cần học:

- Bounded Context.
- Aggregate.
- Domain Event.
- Domain Service.
- Repository theo domain intent.

File cần đọc:

- `22-domain-driven-design/README.md`
- `23-cqrs-event-driven/README.md`

Mastery criteria:

- Biết dùng DDD vừa đủ cho domain phức tạp.
- Không ép DDD vào CRUD đơn giản.

## Phase 12: Production Architecture

Kiến thức cần học:

- Config, secrets, health check, readiness, liveness.
- Graceful shutdown, timeout, retry, backoff.
- Observability: log, metrics, tracing, OpenTelemetry.

File cần đọc:

- `19-logging-observability/README.md`
- `24-production-architecture/README.md`

Mastery criteria:

- Biết production concern thuộc layer nào.
- Instrumentation không làm domain phụ thuộc framework quan sát.

## Phase 13: Refactoring

Kiến thức cần học:

- Refactor từ layered/spaghetti sang Clean Architecture từng bước.
- Identify business rules.
- Extract domain, port, use case, adapter.

File cần đọc:

- `25-refactoring/README.md`
- `26-anti-patterns/README.md`

Lab:

- `labs/lab-11-refactoring`

Mastery criteria:

- Refactor được một handler trộn HTTP, SQL, validation, Kafka thành các boundary rõ hơn mà không rewrite toàn bộ.

## Phase 14: Advanced Architecture

Kiến thức cần học:

- CQRS, Event Sourcing, Saga, Outbox, Specification Pattern.
- Microservices vs Modular Monolith.
- Distributed systems limitation.

File cần đọc:

- `23-cqrs-event-driven/README.md`
- `28-system-design/README.md`
- `29-interview-review/README.md`

Lab:

- `labs/lab-12-full-application`

Mastery criteria:

- Biết Clean Architecture giải quyết dependency và business rule independence, không tự giải quyết distributed transaction hoặc network failure.
