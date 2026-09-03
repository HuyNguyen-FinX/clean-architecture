# Interview Review: trả lời bằng reasoning, code và trade-off

Senior answer không dừng ở tên pattern. Cấu trúc tốt: clarify context → identify forces → decision → consequences → failure/test → alternative.

## Rubric tự chấm

| Mức | Dấu hiệu |
|---|---|
| 1 | đọc định nghĩa/layer |
| 2 | có ví dụ nhưng tuyệt đối hóa |
| 3 | giải thích dependency/trade-off |
| 4 | production failure/testing |
| 5 | context-specific alternatives/evolution |

## 1. Clean Architecture giải vấn đề gì?

Model answer: quản source dependency và ownership để business policy không bị detail khó đổi/khó test chi phối. Nó không tự giải distributed transaction, scale hoặc reliability. Boundary chỉ đáng khi giảm change/risk cost.

Follow-up: Todo CRUD có cần đủ layers? Không; proportional architecture.

## 2. Compile-time vs runtime dependency

Runtime use case gọi Postgres object. Compile time:

~~~text
application → repository interface
postgres adapter → application interface
~~~

DI nối object. Dependency Inversion nói source dependency, không cấm runtime call ra I/O.

## 3. Interface đặt đâu?

Thường cạnh consumer vì consumer định nghĩa capability tối thiểu. Có thể ở domain nếu abstraction là stable domain concept, hoặc shared published package có governance. Không “mọi interface luôn ở application”.

Smell: producer interface mirror all concrete methods.

## 4. Use case import sql.DB có sai?

Trong strict Clean Architecture, application phụ thuộc detail. Hậu quả test/schema/transaction coupling. Nhưng tiny DB-centric service có thể chấp nhận trade-off. Hỏi volatility/domain/risk, không trả lời chỉ “vi phạm”.

## 5. Repository vs DAO

DAO nói rows/CRUD; DDD Repository nói Aggregate/domain collection. Read projection có thể dùng query port/DAO. Repository không phải wrapper bắt buộc.

## 6. Entity vs DTO vs DB model

Entity có identity/behavior/invariant; DTO transport; row persistence. Không bắt ba structs khi shapes/lifecycles giống và domain trivial. Tách khi evolution/security/nullability/behavior khác.

## 7. Application Service vs Domain Service

Application: load, authorization, transaction, call ports, save/publish. Domain Service: business rule không thuộc tự nhiên Entity/VO, không orchestration I/O.

## 8. Transaction boundary

Gần atomic use case vì Repository riêng không biết multi-write. Transfer gồm Accounts + Transfer + idempotency + outbox. Implementation via Transactor/UoW. Domain không biết sql.Tx.

Follow-up: one SQL update có thể repository-managed.

## 9. Context-based transaction trade-off

Ưu: port không leak driver, multi-repo thuận. Nhược: hidden dependency, background context thoát tx, runtime key/nested semantics. Alternatives explicit tx repositories/UoW callback.

## 10. Lost update

Read Committed + read-compute-write không lock/version. Fix SELECT FOR UPDATE, optimistic expected version hoặc atomic SQL. Mutex chỉ một process.

## 11. Deadlock

Opposite transfers lock A/B reverse order. Lock sorted IDs, keep tx short, index query, bounded retry deadlock victim. Không tuyên bố xóa mọi deadlock.

## 12. Transaction + network

DB transaction không rollback remote. Giữ lock lâu, timeout ambiguous. Persist Pending/outbox, call remote outside, idempotency + reconciliation/Saga.

## 13. HTTP adapter responsibilities

Method/content/body limit/decode, DTO→command, auth extraction, context, result/error→stable response. Không domain invariant/SQL. “Thin” vẫn có protocol policy.

## 14. Error handling

Domain stable errors; adapter maps pgx/provider; application workflow errors; delivery HTTP/gRPC. errors.Is/As/%w. Unknown safe public response, log once outer boundary.

## 15. Validation

Syntax delivery; workflow/actor application; invariant domain; durable race constraint DB. Duplicate simple check can improve UX, but domain is source of truth.

## 16. Kafka consumer layer?

Driving adapter, dù folder có thể gọi infrastructure. It maps message to command, handles offset/retry/DLQ protocol. Business logic use case.

## 17. At-least-once/idempotency

Crash after DB commit before offset commit gives duplicate. Inbox marker + effect same transaction. Event ID stable. Offset after safe outcome.

## 18. Outbox

State + intent-to-publish same DB transaction. Worker publish then mark. Crash after publish before mark duplicates, so consumer idempotent. Không exactly-once end-to-end.

## 19. Redis placement

Repository decorator, application cache port, HTTP cache tùy semantics. Balance Aggregate stale nguy hiểm; history projection phù hợp hơn. Cache down fail-open only performance cache.

## 20. Testing strategy

Domain no mocks; use case fake/spy; HTTP httptest; Repository/transaction real Postgres; broker integration; few E2E; race/fuzz/architecture. Mock SQL không chứng minh query/schema/lock.

## 21. Fake vs mock

Fake working simplified state; stub canned input; spy records outputs; mock predeclared expectations. Use based question, avoid sequence-heavy mock.

## 22. DDD vs Clean Architecture

DDD models domain/language/context; Clean Architecture dependency/ownership. Bounded Context not microservice. Tactical DDD not needed for trivial CRUD.

## 23. Aggregate boundary

Consistency/invariant boundary, not graph of structs. Smaller avoids contention/load; cross-Aggregate workflow coordinated by application transaction/events.

## 24. CQRS

Separate read/write responsibilities; can share process/DB. Event Sourcing/Kafka not required. Physical split adds lag/projection/rebuild.

## 25. Event Sourcing

Events source state, expected version append, fold/snapshot/upcast. Strong audit/rebuild but schema forever-ish/ops complexity. Audit requirement alone not sufficient.

## 26. Observability without domain SDK

Middleware/decorators/adapters create spans/metrics/log. Domain error/event provides outcome. context propagates through I/O, not Account.Withdraw.

## 27. Readiness vs liveness

Liveness asks restart helps; readiness asks accept traffic guarantee. DB down likely unready; cache down can degraded; Kafka down + outbox may remain ready with backlog alert.

## 28. Graceful shutdown

Readiness false, stop fetch, shutdown HTTP/drain, flush workers/producers/telemetry, close pools, timeout. log.Fatal/os.Exit bypass defers.

## 29. Over-engineering

Boundary cost: mapper/interfaces/files/testing/cognitive load. Add when domain/volatility/team/risk pressure. Remove pass-through layers. Simple is not same as coupled.

## 30. Code review prompt

Given:

~~~go
func (s *TransferService) Transfer(ctx context.Context, req *pb.TransferRequest) error {
	tx, _ := s.pool.Begin(ctx)
	account, _ := s.repo.Find(ctx, req.From)
	account.Balance -= req.Amount
	go s.kafka.Publish(context.Background(), req)
	return tx.Commit(ctx)
}
~~~

Expected findings:

- pb leaks transport;
- errors ignored;
- repository may not use tx;
- public mutable balance/invariant bypass;
- fire-and-forget/lost context;
- publish before/independent commit;
- no idempotency;
- no rollback;
- no receiver;
- unknown concurrency.

Prioritize correctness before naming/folders.

## 31. System design prompt

Design 5k TPS transfer:

1. clarify guarantees;
2. capacity;
3. sync transaction/ledger;
4. idempotency;
5. outbox/async;
6. partition/hot account;
7. API/status;
8. failures/reconcile;
9. observability/security;
10. evolution.

## 32. Behavioral prompts

- Kể một refactor architecture: context, risk, incremental steps, metrics, outcome, lesson.
- Kể incident do retry/timeout: detection, mitigation, root cause, prevention.
- Khi phản đối pattern: dữ liệu/trade-off/alternative, không preference cá nhân.

## 33. Trap questions

### Bao nhiêu layers?

Không có con số magic. Boundary theo policy/detail/change.

### Interface luôn cần?

Không. Go interface nhỏ consumer-side khi có abstraction need.

### Clean Architecture chậm?

Runtime overhead interface thường không đáng kể; development/cognitive/mapping cost mới là trade-off. Measure performance.

### Microservices sạch hơn?

Không. Thêm network/distributed data. Modular monolith có thể rõ hơn.

## 34. Practice drills

1. Trả lời mỗi câu trong 2 phút rồi 10 phút.
2. Vẽ compile/runtime graph.
3. Code Account.Withdraw + tests.
4. Code Transactor/UoW signatures.
5. Review God Handler.
6. Design failure timeline.
7. Nêu khi không dùng pattern.

## 35. Final mastery

Bạn sẵn sàng khi có thể:

- giải thích WHY bằng causal chain;
- code idiomatic Go nhỏ;
- nêu guarantee/failure;
- chọn test;
- đưa alternative;
- scale architecture theo context;
- nói “không cần pattern này” có căn cứ.

## Further reading

Quay lại Chapters 01, 02, 03, 11, 14, 20, 24, 25 và làm exercises/case studies không nhìn solution trước.

## Quality gate

- [x] Model answers 30 topics
- [x] Follow-up/traps
- [x] Code review/system design
- [x] Production/testing/trade-off
- [x] Rubric/practice/mastery
