# Solution: Review Transfer Service

Không có một refactor duy nhất. Findings sau ưu tiên correctness trước structure.

## P0: Transaction không bao Repository writes

Service BeginTx nhưng Repository methods chỉ nhận context và concrete implementation có thể dùng pool. tx commit/rollback không quản Save. Receiver save fail có thể để source đã debit.

Fix: Transactor context contract có integration test, hoặc UoW callback cấp tx-bound repositories. Test real PostgreSQL inject second-save failure và reload cả hai.

## P0: Public Balance bypass invariant

Service mutate int64 trực tiếp; negative/overflow/frozen/currency không được bảo vệ. Concurrent stale object càng nguy hiểm.

Fix incremental: tạo Account.Withdraw/Deposit private state, table tests, mapper Repository.

## P0: Fire-and-forget Kafka

Goroutine dùng Background, request trả trước outcome; process shutdown mất event; publish có thể xảy ra trước DB durable; error ignored.

Fix: Transfer + outbox same transaction, owned worker publish/mark, idempotent consumer. EventPublisher synchronous sau commit vẫn còn dual-write.

## P1: Không idempotency

Commit success/response lost khiến client retry double transfer. Thêm key+request hash/state cùng transaction, replay stable result.

## P1: Context bị mất

Background trong goroutine bỏ cancellation/trace; goroutine không ownership. Nếu best-effort vẫn cần owned worker/lifecycle.

## P1: Domain error chứa HTTP

DomainError.Status tạo dependency transport. Domain stable error; handler map 409. Kafka/gRPC có mapping riêng.

## P1: Raw error public

http.Error(err.Error()) leak SQL/provider. Return stable public code; log wrapped cause outer boundary.

## P1: Errors/commit ignored

Begin/Find/Save/Commit outcomes phải xử lý. Commit failure là operation failure/possibly ambiguous. Deferred rollback.

## P1: Lost update

Load/mutate/save không SELECT FOR UPDATE/version. Add stable lock order for two accounts hoặc optimistic expected version.

## P2: Interface không nói guarantee

Repository port chỉ Find/Save nhưng transaction semantics không công bố. Interface shape đúng không đủ. Contract/tests/documentation cần nói tx-bound lock behavior.

## Safe sequence

1. Characterization + PostgreSQL rollback/concurrency test.
2. Handle errors/rollback/public sanitization.
3. Extract Account behavior/tests.
4. Introduce Transactor/UoW and tx-bound Repository.
5. Add idempotency + Transfer record.
6. Add outbox, remove goroutine publish.
7. Move HTTP DTO/mapping.
8. Add fitness tests/observability.

## Verification

- domain boundaries;
- receiver-save rollback;
- opposite transfer concurrency;
- duplicate HTTP key;
- Kafka down outbox pending;
- worker duplicate;
- unknown error not leaked;
- graceful shutdown.
