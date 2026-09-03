# Progress

Trạng thái trong file này dùng quality gate tại [CONTENT_AUDIT.md](./CONTENT_AUDIT.md), không dựa vào việc folder/file đã tồn tại.

## Ý Nghĩa Trạng Thái

- `[DEEP COMPLETE]`: đạt quality gate production-grade learning material; code và link liên quan đã được kiểm chứng.
- `[FOUNDATIONAL COMPLETE]`: đủ làm nền và đã có chiều sâu đáng kể, nhưng còn thiếu ít nhất một quality gate nâng cao.
- `[IN PROGRESS]`: đang được rewrite theo chiều sâu; chưa nên học như chapter hoàn chỉnh.
- `[OUTLINE ONLY]`: mới là summary, danh sách chủ đề hoặc hướng làm.
- `[NOT STARTED]`: chưa có material có ý nghĩa.

## Core Docs

- [DEEP COMPLETE] `README.md`
- [DEEP COMPLETE] `ROADMAP.md`
- [DEEP COMPLETE] `GLOSSARY.md`
- [DEEP COMPLETE] `CHEATSHEET.md`
- [DEEP COMPLETE] `CONTENT_AUDIT.md` cho baseline 2026-09-03
- [DEEP COMPLETE] Runnable example `examples/mini-banking`: V1-V11 có Domain, PostgreSQL transaction/locking, Transfer history, concurrent idempotency claim, Kafka adapters, outbox relay, strict HTTP, metrics/trace correlation và graceful lifecycle.

## Chapters

- [DEEP COMPLETE] 00 Software Architecture Foundations
- [DEEP COMPLETE] 01 Clean Architecture Foundations
- [DEEP COMPLETE] 02 Dependency Rule
- [DEEP COMPLETE] 03 Domain Layer
- [DEEP COMPLETE] 04 Use Case / Application Layer
- [DEEP COMPLETE] 05 Repository Pattern
- [DEEP COMPLETE] 06 Delivery Layer
- [DEEP COMPLETE] 07 Infrastructure Layer
- [DEEP COMPLETE] 08 Dependency Injection
- [DEEP COMPLETE] 09 Project Structure
- [DEEP COMPLETE] 10 Database
- [DEEP COMPLETE] 11 Transaction Management
- [DEEP COMPLETE] 12 HTTP REST API
- [DEEP COMPLETE] 13 gRPC
- [DEEP COMPLETE] 14 Kafka Event Driven
- [DEEP COMPLETE] 15 Redis Cache
- [DEEP COMPLETE] 16 External Services
- [DEEP COMPLETE] 17 Error Handling
- [DEEP COMPLETE] 18 Validation
- [DEEP COMPLETE] 19 Logging / Observability
- [DEEP COMPLETE] 20 Testing
- [DEEP COMPLETE] 21 Concurrency Golang
- [DEEP COMPLETE] 22 Domain Driven Design
- [DEEP COMPLETE] 23 CQRS / Event Driven
- [DEEP COMPLETE] 24 Production Architecture
- [DEEP COMPLETE] 25 Refactoring
- [DEEP COMPLETE] 26 Anti-patterns
- [DEEP COMPLETE] 27 Case Studies
- [DEEP COMPLETE] 28 System Design
- [DEEP COMPLETE] 29 Interview Review

## Labs

- [DEEP COMPLETE] `lab-01-simple-domain`
- [DEEP COMPLETE] `lab-02-usecase`
- [DEEP COMPLETE] `lab-03-repository`
- [DEEP COMPLETE] `lab-04-postgresql`
- [DEEP COMPLETE] `lab-05-http`
- [DEEP COMPLETE] `lab-06-dependency-injection`
- [DEEP COMPLETE] `lab-07-testing`
- [DEEP COMPLETE] `lab-08-transaction`
- [DEEP COMPLETE] `lab-09-kafka`
- [DEEP COMPLETE] `lab-10-redis`
- [DEEP COMPLETE] `lab-11-refactoring`
- [DEEP COMPLETE] `lab-12-full-application`

## Exercises Và Case Studies

- [DEEP COMPLETE] Architecture exercises 01-07: problem/solution tách riêng, đề có assumptions/failure injection/deliverables/self-review và đáp án có reasoning/trade-off.
- [DEEP COMPLETE] Code review exercise 01: realistic Go source compile được, subtle findings, safe sequence và verification matrix.
- [DEEP COMPLETE] Case studies 01-08: 1.200+ dòng phân tích context-specific về model, boundaries, transaction, failure, test, operations và alternative.

## Trạng Thái Curriculum

Curriculum rewrite theo quality gate hiện đã khép đủ 30 chapter, 12 labs, mini-banking V1-V11, 7 architecture exercises, code review exercise và 8 case studies. Các giới hạn của learning implementation như broker integration, OpenTelemetry exporter chuẩn và double-entry ledger được công bố rõ, không bị trình bày như guarantee đã có.

## Quy Tắc Cập Nhật

Mỗi lần đổi trạng thái phải ghi bằng chứng ở cuối file:

- Material nào được thêm hoặc sửa.
- Full example nào compile/test.
- Failure scenario nào đã được phân tích.
- Quality gate nào còn thiếu.

Không nâng trạng thái hàng loạt chỉ vì đã thêm cùng một template vào nhiều chapter.

## Lịch Sử Increment

### 2026-09-03 - Foundations, Dependency Rule Và Domain

- `CONTENT_AUDIT.md`: audit baseline toàn bộ 118 file/4.780 dòng và xếp priority.
- Chapter 01: 878 dòng, đạt quality gate foundations.
- Chapter 02: 894 dòng và architecture fitness test bằng `go/parser`.
- Chapter 03: 10 bài, tổng 2.831 dòng, dùng mini-banking làm model xuyên suốt.
- Mini-banking domain: constructor invariant, Currency/Money equality, overflow guard, frozen state và test matrix.
- `lab-01`: starter/solution có code thật và hướng dẫn 60-90 phút.
- Verification: mini-banking, lab starter và lab solution đều pass `go test -race ./...`; relative links của material mới tồn tại; `git diff --check` sạch.

### 2026-09-03 - Application, Repository, PostgreSQL Và Transaction

- Chapter 04-05: use-case orchestration, consumer-owned ports, Repository semantics, ownership, mapping, transaction/concurrency và testing strategy.
- Chapter 08-11: manual DI/lifecycle, feature-oriented package ownership, pgx database adapter, sáu transaction patterns, isolation/locking/deadlock/retry/idempotency/outbox reasoning.
- Mini-banking V5a: pgx v5.7.2 trên Go 1.22, migration, strict row rehydration, context-bound transaction, `SELECT FOR UPDATE`, deterministic Account lock order và optional PostgreSQL integration suite.
- Labs 02/03/04/06/08: starter và solution đều có Go module/code/test thật; Repository contract test, composition smoke test và rollback/concurrency tests.
- Verification: mọi module mới pass `go test -race ./...` và `go vet ./...`; PostgreSQL suites compile và skip có thông báo khi thiếu `TEST_DATABASE_URL`.

### 2026-09-03 - Delivery, HTTP, Kafka, Redis Và External Adapters

- Chapter 06-07 và 12-18: driving/driven adapters, strict HTTP, gRPC mapping/evolution, Kafka delivery semantics/outbox, Redis consistency, external ambiguous outcomes, error taxonomy và validation ownership.
- Mini-banking HTTP: strict single-object JSON, unknown-field/body limit, safe stable error response, nil dependency guard và tests chống internal-error leak.
- Labs 05/09/10: executable baselines và solutions cho HTTP contract, duplicate-safe consumer/outbox worker, TTL cache decorator/fail-open.
- Verification: tất cả module mới pass `go test -race ./...` và `go vet ./...`.

### 2026-09-03 - Testing, Observability, DDD/CQRS Và Production

- Chapter 19-24: structured telemetry/cardinality/SLO, full test portfolio, Go concurrency ownership, strategic+tactical DDD, CQRS/Event Sourcing/Saga và lifecycle/reliability.
- Mini-banking: slog request middleware, bounded request ID, server timeouts, signal-based graceful shutdown và resource cleanup không dùng log.Fatal.
- Lab 07: domain/use-case/HTTP/memory test portfolio có code và test gaps explicit.
- Lab 12: capstone chạy được với Account, explicit UnitOfWork, Transfer history, durable-semantics idempotency hash, outbox intent và HTTP replay/history tests.
- Verification: mini-banking và lab modules pass `go test -race ./...`, `go vet ./...`.

### 2026-09-03 - Refactoring, Case Studies Và Vertical Slice V1-V11

- Chapter 25-29: refactor Step 0-8, 20+ anti-patterns, cross-case synthesis, system design transfer 5.000 TPS và interview model answers/rubric.
- Lab 11: executable God Handler starter và refactored Domain/Application/Adapters solution; race/vet sạch.
- Exercises 01-07: đề không lộ đáp án nhưng có constraints, failure injection, deliverables và self-review; solution phân tích context/trade-off.
- Case studies 01-08: từ brief 12-14 dòng thành 150-170 dòng/case, bao phủ model, dependency, data/transaction, concurrency, failure matrix, tests, observability và alternatives.
- Mini-banking V1-V11: Transfer Entity/history, atomic memory/PostgreSQL artifacts, concurrency-safe durable idempotency claim, outbox relay, real Kafka producer/consumer adapters, strict history API, metrics, W3C trace correlation và lifecycle.
- Core docs: foundations có investigation/failure exercise; Glossary có semantic relationships/counter-examples; Cheatsheet có decision/failure/review checklists.
- Verification: mini-banking, code-review starter và toàn bộ 24 starter/solution lab modules pass `go test -race ./...` và `go vet ./...`. PostgreSQL integration suites compile và skip có chủ đích khi thiếu `TEST_DATABASE_URL`.
