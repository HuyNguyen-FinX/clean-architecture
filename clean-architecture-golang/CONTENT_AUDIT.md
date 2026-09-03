# Content Audit

Ngày audit: 2026-09-03

Audit này ghi lại **baseline trước đợt rewrite chuyên sâu**. Điểm số không đo số dòng hay số heading; nó đo khả năng người học hiểu bản chất, đọc code chạy được, phân tích dependency, failure mode và tự ra quyết định kiến trúc.

## Thang Điểm

| Điểm | Ý nghĩa |
|---:|---|
| 0 | Chưa có material có ý nghĩa để học |
| 1 | Outline: mới liệt kê chủ đề, câu hỏi hoặc hướng làm |
| 2 | Giải thích sơ lược: có thông điệp đúng nhưng chưa đủ để tự áp dụng |
| 3 | Đủ kiến thức cơ bản: có thể hiểu và làm ví dụ nhỏ, còn thiếu chiều sâu production |
| 4 | Chuyên sâu: có reasoning, code, trade-off và scenario đáng kể, còn một vài quality gate chưa đạt |
| 5 | Production-grade learning material: đủ lý thuyết, implementation, failure analysis, test, exercise và nguồn đọc tiếp |

## Phạm Vi Đã Đọc

Audit đã mở và đọc toàn bộ 118 file, tổng cộng 4.780 dòng tại thời điểm baseline:

- 5 tài liệu gốc: `README.md`, `ROADMAP.md`, `PROGRESS.md`, `GLOSSARY.md`, `CHEATSHEET.md`.
- 30 chapter từ `00` đến `29`.
- Toàn bộ Markdown và Go source trong `examples/`, `labs/`, `exercises/`, `code-review-exercises/`, `case-studies/`.
- `go.mod`, production code và test code của `examples/mini-banking` và `lab-01-simple-domain`.

## Kết Luận Điều Hành

Repository có curriculum map tốt, thuật ngữ nhìn chung chính xác và một vertical slice Go chạy được. Tuy nhiên nó chưa phải giáo trình chuyên sâu:

- 29/30 chapter chưa đạt mức 4. Phần lớn chỉ dài 27-51 dòng và dừng ở summary.
- Chapter `00` là material tốt nhất, nhưng vẫn thiếu debugging walkthrough, nguồn tham khảo và bài thực hành gắn trực tiếp với code để đạt mức 5.
- Chapter `01` và `02` có đúng mental model nhưng lặp nhiều nội dung của chapter `00`, chưa dẫn người học qua một quyết định kiến trúc từ code xấu đến dependency đúng.
- Chapter `03`, `11` và `20` là các lỗ hổng lớn: Domain, Transaction và Testing đều mới ở mức giới thiệu.
- Chỉ `lab-01` có Go starter/solution thật. Lab `02-12` có folder nhưng hầu như chỉ chứa README.
- `examples/mini-banking` compile và test được, nhưng mới minh họa memory adapter. Nó chưa có transaction thật, PostgreSQL, idempotency, history, Kafka/outbox hoặc observability như learning path đã hứa.
- Exercises và case studies tạo được câu hỏi đúng hướng, nhưng solution/case analysis quá ngắn để người học đối chiếu reasoning ở senior level.

## Audit Core Docs

| Tài liệu | Score | Điểm đang tốt | Vấn đề / thiếu gì | Priority |
|---|---:|---|---|---|
| `README.md` | 3 | Mục tiêu, learning path và triết lý thực dụng rõ | Chưa chỉ trạng thái maturity của từng phần; chưa mô tả evolution V1-V11 | Medium |
| `ROADMAP.md` | 3 | Phase, mastery criteria và link tương đối rõ | Nhiều phase trỏ tới material mới là outline; lab được mô tả như đã sẵn sàng dù thiếu code | High |
| `PROGRESS.md` | 2 | Có cảnh báo không đánh complete cho outline | `[COMPLETE]` ở core docs/example chưa dùng chuẩn mới; mọi chapter còn lại cùng một nhãn nên không thể hiện độ chênh | Critical |
| `GLOSSARY.md` | 2 | Định nghĩa ngắn và phần lớn đúng | Thiếu ví dụ, counter-example, quan hệ giữa khái niệm và link tới chapter sâu | Low |
| `CHEATSHEET.md` | 3 | Hữu ích khi review dependency và over-engineering | Một số rule of thumb cần context/link giải thích để tránh biến thành luật tuyệt đối | Medium |

## Audit Chapters

| Module | Score | Vấn đề thực tế | Thiếu gì để đạt mức 5 | Priority |
|---|---:|---|---|---|
| 00 Software Architecture Foundations | 4 | Material tốt nhưng một số phần vẫn là survey | Debug/investigation walkthrough, executable dependency check, failure exercise, further reading có chú giải | Medium |
| 01 Clean Architecture Foundations | 2 | Lặp mental model của 00; chưa đủ một buổi học | Problem progression, boundary anatomy, policies/mechanisms, full Go flow, wrong/correct comparison, runtime sequence, production/failure scenario, testing, exercises, references | Critical |
| 02 Dependency Rule | 3 | Có import/runtime distinction và interface placement cơ bản nhưng còn ngắn | Import graph thực, semantic/data/ownership dependency, output boundary, compile-time failure, alternative designs, package tests, investigation workflow | Critical |
| 03 Domain Layer | 2 | Mới liệt kê Entity/VO/Aggregate; gần như chưa dạy modeling | Identity, equality, invariant, aggregate boundary, factory, state transition, domain service/event/error, rich vs anemic, full code/test và production concurrency reasoning | Critical |
| 04 Use Case / Application Layer | 2 | Chỉ mô tả orchestration ở mức khái niệm | Full use case, command/result, domain vs application service, transaction/idempotency policy, error propagation, test doubles, failure matrix | Critical |
| 05 Repository Pattern | 2 | Nêu đúng Repository != DAO nhưng chưa chứng minh | Repository semantics, aggregate boundary, method design, interface ownership, mapping, concurrency/locking, fake và PostgreSQL implementation/test | Critical |
| 06 Delivery Layer | 2 | Có flow và responsibility cơ bản | Authentication context, DTO mapping, transport validation, deadlines, error contract, HTTP/gRPC/Kafka comparison và code/test | Medium |
| 07 Infrastructure Layer | 2 | Danh sách concern đúng | Adapter design cụ thể, error translation, resilience ownership, lifecycle, integration test và production diagnostics | Medium |
| 08 Dependency Injection | 2 | Manual DI và composition root được nhắc đúng | Object graph hoàn chỉnh, lifecycle/cleanup, constructor invariant, optional dependency smell, Wire/Fx trade-off và wiring test | Critical |
| 09 Project Structure | 2 | Nêu package by layer/feature và `internal` | So sánh theo quy mô code/team, import graph, bounded context ownership, migration path, cyclic dependency case | Critical |
| 10 Database | 2 | Nêu mapping và pool | Domain/DB/DTO model cụ thể, pgx repository chạy được, migration, constraints, nullable data, integration test, query/locking failure | Critical |
| 11 Transaction Management | 2 | Nhận ra transaction boundary thường ở use case | ACID/isolation, sáu transaction pattern, production `Transactor`, rollback/panic/commit failure, row lock, optimistic lock, retry, idempotency, network-call analysis | Critical |
| 12 HTTP REST API | 2 | Flow và error mapping đúng | Full handler, strict decode, request limits, auth, timeout, response contract, middleware boundary, test matrix và failure diagnostics | Critical |
| 13 gRPC | 2 | Phân biệt protobuf DTO với domain | Generated contract example, interceptor, status mapping, deadlines, streaming, compatibility và tests | Low |
| 14 Kafka Event Driven | 2 | Hai chiều consumer/producer và concern chính đã có | At-least-once flow, offset/rebalance, idempotent consumer, retry/DLQ, ordering, schema evolution, outbox code và integration tests | Critical |
| 15 Redis Cache | 2 | Nêu nhiều vị trí cache hợp lệ | Cache-aside code, decorator/application/middleware comparison, invalidation race, stampede, lock caveat, metrics và tests | Medium |
| 16 External Services | 2 | Gateway và ambiguous failure được nhận diện | Production HTTP client, timeout budget, retry classification, idempotency/reconciliation, circuit breaker placement và contract tests | Medium |
| 17 Error Handling | 2 | Phân loại layer và `errors.Is/As` đúng hướng | Typed/sentinel trade-off, wrapping policy, stable application taxonomy, privacy, retryability, full mapping code/test | Critical |
| 18 Validation | 2 | Phân biệt transport/application/domain validation | Cross-field/cross-aggregate validation, normalization ownership, race giữa validate và write, examples và tests | Medium |
| 19 Logging / Observability | 2 | Không đưa framework vào domain là đúng | Instrumentation boundary code, context propagation, span/error semantics, metric cardinality, redaction, trace walkthrough và incident query | Critical |
| 20 Testing | 1 | Mới là test pyramid và danh sách smell | Code thật cho domain/use case/HTTP/repository/E2E; mock/fake/stub/spy; contract test; flaky test; testcontainers; coverage strategy | Critical |
| 21 Concurrency Golang | 2 | Nêu đúng orchestration thường ngoài domain | Race example, ownership, errgroup/worker pool implementation, backpressure, cancellation leak, database concurrency relation và tests | Medium |
| 22 Domain Driven Design | 1 | Chỉ liệt kê tactical patterns | Strategic vs tactical DDD, ubiquitous language, bounded context discovery, context mapping, aggregate design, integration events và modeling workshop | Critical |
| 23 CQRS Event Driven | 1 | Chỉ giới thiệu tên pattern/trade-off | Read/write model example, consistency timeline, event sourcing model, outbox/saga failure matrix, operations/debugging và code | Medium |
| 24 Production Architecture | 2 | Concern production được nhận diện | Runnable lifecycle, graceful shutdown, readiness semantics, config validation, timeout budget, retry policy, deployment/migration scenario | Critical |
| 25 Refactoring | 1 | Chỉ có sáu bước tổng quát | Code Step 0-8, characterization tests, incremental commits, dependency changes, rollback strategy và measured outcomes | Critical |
| 26 Anti-patterns | 1 | Danh sách đúng nhưng chưa phải bài học | Subtle realistic code, consequences, diagnosis, refactor options, justified exceptions và review exercise | Critical |
| 27 Case Studies | 1 | Chỉ là index/cách đọc | Không có một case study hoàn chỉnh tại chapter này; cần synthesis và link tới case đã phát triển sâu | Low |
| 28 System Design | 2 | Phân biệt Clean Architecture và microservices đúng | Service boundary method, data ownership, consistency choices, capacity/failure reasoning, deployment topology và case study | Low |
| 29 Interview Review | 1 | Có một số câu hỏi tốt | Model answers theo reasoning, follow-up traps, code review task, system scenario và rubric tự chấm | Low |

## Audit Runnable Example

### Điểm: 3/5

`examples/mini-banking` là phần có giá trị: package direction rõ, interface nhỏ, HTTP test dùng fake, domain test không mock và `go test ./...` chạy được. Tuy nhiên đây mới là V1-V3 của curriculum.

Các khoảng trống và rủi ro cần xử lý theo evolution:

| Khu vực | Nhận xét | Hệ quả học tập |
|---|---|---|
| Domain constructor | `NewAccount` kiểm tra currency và overdraft dương nhưng chưa reject initial balance thấp hơn `-overdraftLimit` | Có đường tạo aggregate đã vi phạm invariant dù `Withdraw` bảo vệ đúng |
| Money arithmetic | `int64` addition/subtraction chưa kiểm tra overflow | Không nên gọi là money implementation production-grade nếu chưa nêu giới hạn |
| Transaction | `NoopTransactor` chỉ chạy closure; memory repository save từng account riêng | Không chứng minh atomicity; save thứ hai lỗi có thể để account thứ nhất đã đổi |
| Concurrency | Load hai clone rồi save không có version/lock | Concurrent transfer có thể lost update |
| HTTP error | Handler trả `err.Error()` cả lỗi unknown | Có thể lộ infrastructure detail cho client |
| HTTP decode | Chưa reject JSON value thứ hai sau object đầu | Contract parsing chưa strict hoàn toàn |
| Server lifecycle | Có timeout header nhưng chưa graceful shutdown | Chưa đạt production lifecycle |
| Feature scope | Chưa có Deposit/Withdraw endpoint, Transfer record/history, idempotency | Chưa phải project xuyên suốt V1-V11 |
| Infrastructure | Chưa có PostgreSQL/Kafka/outbox/observability | Chưa kiểm chứng adapter và production failure mode |

Không cần sửa tất cả trong một bước. Mỗi version phải ghi rõ guarantee nào đã có và guarantee nào chưa có để người học không nhầm demo với production.

## Audit Labs

| Lab | Score | Hiện trạng | Thiếu gì | Priority |
|---|---:|---|---|---|
| 01 Simple Domain | 3 | Có code compile, starter, solution và test | Starter chưa tạo red tests dẫn tới đầy đủ behavior; constructor/test matrix còn thiếu; explanation cần sâu hơn | Critical |
| 02 Use Case | 1 | Chỉ có README trong starter/solution | Go module, starter code có failure có chủ đích, fake repo, solution compile/test, diagram và solution walkthrough | Critical |
| 03 Repository | 1 | Chỉ có README | Runnable port + adapters, contract test và design alternatives | Critical |
| 04 PostgreSQL | 1 | Chỉ có README | Schema, migration, pgx adapter, container/local setup, integration tests | Critical |
| 05 HTTP | 1 | Chỉ có README | Runnable handler starter/solution, strict validation/error mapping tests | High |
| 06 Dependency Injection | 1 | Chỉ có README | Object graph code, lifecycle cleanup, alternative wiring và tests | High |
| 07 Testing | 1 | Chỉ có README | Code cho từng test level và các test double | Critical |
| 08 Transaction | 1 | Chỉ có README | PostgreSQL transaction implementation, rollback/locking/concurrency tests | Critical |
| 09 Kafka | 1 | Chỉ có README | Consumer/producer code, duplicate/retry tests, outbox evolution | High |
| 10 Redis | 1 | Chỉ có README | Cache decorator/rate-limit code, invalidation/stampede scenario và tests | Medium |
| 11 Refactoring | 1 | Chỉ có README | Realistic God Handler code, characterization test và từng refactor step | Critical |
| 12 Full Application | 1 | Chỉ có README | Runnable integrated application, deployment dependencies và end-to-end tests | High |

## Audit Exercises Và Case Studies

| Nhóm | Score | Nhận xét | Priority |
|---|---:|---|---|
| Architecture exercises 01-07 | 2 | Problem tách khỏi solution và câu hỏi tốt; solution quá ngắn, thiếu alternative/rubric/failure analysis | Medium |
| Code review exercise 01 | 1 | Code xấu chỉ là comment minh họa, chưa đủ tinh tế để review như production code | High |
| Case study briefs 01-08 | 1 | Mỗi case 12-14 dòng, chỉ là danh sách chủ đề và kết luận | Medium |

## Ưu Tiên Rewrite

Thứ tự được chọn theo ba yếu tố: đây có phải concept nền cho chapter sau không, khoảng cách với quality gate lớn đến đâu, và mini-banking có thể dùng để chứng minh ngay không.

1. `01-clean-architecture-foundations`: thiết lập vocabulary và phương pháp reasoning thống nhất.
2. `02-dependency-rule`: làm rõ source/runtime/data/semantic/ownership dependency bằng Go import graph.
3. `03-domain-layer`: xây model Account/Money đúng trước khi orchestration và persistence lớn lên.
4. `04-usecase-application-layer` và `05-repository-pattern`: hoàn thiện V2-V3 của mini-banking.
5. `08-dependency-injection` và `09-project-structure`: nối package graph với object graph/team ownership.
6. `10-database` và `11-transaction-management`: thêm PostgreSQL, locking và atomicity thật.
7. `12-http-rest-api`, `17-error-handling`, `20-testing`: củng cố boundary bằng transport contract và test strategy.
8. `14-kafka-event-driven`, `19-logging-observability`, `24-production-architecture`: tiến tới outbox và vận hành production.
9. `22-domain-driven-design`, `25-refactoring`, `26-anti-patterns`: tổng hợp modeling và migration skill.

## Quality Gate Áp Dụng Từ Audit Này

Một chapter chỉ được đánh `[DEEP COMPLETE]` khi các mục phù hợp đều có bằng chứng cụ thể:

- [ ] Problem và chuỗi WHY rõ.
- [ ] Mental model và core theory đủ để không phải học thuộc sơ đồ.
- [ ] Go implementation idiomatic; full example compile/test được.
- [ ] Wrong/correct example có phân tích hậu quả, không chỉ dán nhãn.
- [ ] Runtime flow tách khỏi compile-time dependency.
- [ ] Data, semantic và ownership dependency được xét khi liên quan.
- [ ] Trade-off và mục “khi nào không nên dùng”.
- [ ] Ít nhất một production scenario và failure scenario.
- [ ] Có debug/investigation workflow.
- [ ] Testing strategy và test code phù hợp.
- [ ] Exercise không lộ đáp án trước khi người học suy nghĩ.
- [ ] Mastery questions kiểm tra reasoning.
- [ ] Further Reading dùng nguồn chất lượng và có chú giải.
- [ ] Mọi link nội bộ tồn tại; command được ghi đã được chạy kiểm chứng.

## Cách Giữ Audit Luôn Hữu Ích

Điểm trong bảng chapter là baseline, không tự tăng chỉ vì file dài hơn. Sau mỗi increment, `PROGRESS.md` ghi trạng thái hiện tại và phần “Lịch sử increment” ghi bằng chứng: file nào đổi, command nào pass, quality gate nào còn thiếu. Khi chapter đạt 5, audit sẽ thêm một mục re-evaluation thay vì xóa baseline; nhờ đó người học thấy repository đã tiến hóa như thế nào.

## Re-evaluation Sau Increment 01

Ngày đánh giá lại: 2026-09-03

| Material | Baseline | Hiện tại | Bằng chứng nâng chất lượng |
|---|---:|---:|---|
| 01 Clean Architecture Foundations | 2 | 5 | 878 dòng; problem progression, ba level, ports/adapters, 5 loại dependency, runtime/compile-time diagrams, production failures, investigation, tests, exercises và annotated references |
| 02 Dependency Rule | 3 | 5 | 894 dòng; Go import graph, interface ownership alternatives, context/transaction leakage, 6 loại dependency, architecture fitness test chạy trong module |
| 03 Domain Layer | 2 | 5 | 2.831 dòng qua 10 bài; Entity, VO, invariant, Aggregate, Domain Service/Event, walkthrough, anti-patterns, exercises; code/test domain được nâng tương ứng |
| lab-01 Simple Domain | 3 | 5 | README đủ problem/diagram/steps/expected behavior/questions/challenge; starter có executable bad baseline; solution có constructor/transition invariant và test matrix, `go test -race ./...` pass |
| examples/mini-banking | 3 | 3 | Domain và architecture guard mạnh hơn, nhưng toàn project vẫn ở V3; transaction/PostgreSQL/idempotency/outbox chưa có nên chưa nâng score |

Các finding source đã xử lý trong increment:

- `NewAccount` đã reject initial balance thấp hơn `-overdraftLimit` và invalid zero-value Money.
- `Money` đã có `Currency`, value equality, immutable-style arithmetic và overflow/underflow guard.
- `Account` đã có active/frozen transition; frozen withdrawal bị reject và HTTP adapter map outcome có chủ đích.
- Architecture test chặn outer/third-party imports trong production code của domain/application.
- README mini-banking công bố rõ guarantee hiện tại và evolution V1-V11.

Finding còn mở không bị che giấu:

- `NoopTransactor` và memory repository chưa bảo đảm atomic save sender/receiver.
- Chưa có protection cho concurrent lost update.
- Chưa có PostgreSQL, transfer record/history, idempotency, Kafka/outbox và production observability.
- HTTP unknown error vẫn cần stable public response/error taxonomy ở increment HTTP/Error Handling.

Priority sau increment là `04-usecase-application-layer`, `05-repository-pattern`, sau đó `08`, `09`, `10` và `11` như bảng audit.
