# Lab 12: Full Application Vertical Slice

Thời lượng: 3-6 giờ. Đây là capstone ghép Account, Transfer, history, transaction, idempotency, outbox intent, HTTP và composition.

## Mục tiêu

- giữ invariant trong Domain;
- orchestration/UoW/idempotency ở Application;
- adapter memory transactional chạy mọi máy;
- strict HTTP + history query;
- outbox intent atomic với transfer;
- composition root/lifecycle rõ;
- test replay và rollback end-to-end in-process.

## Kiến thức cần

Hoàn thành Labs 01-10 và Chapters 01-24.

## Architecture

~~~mermaid
flowchart LR
    HTTP["HTTP Adapter"] --> UC["TransferMoney"]
    UC --> UOW["UnitOfWork Port"]
    UC --> DOMAIN["Account"]
    MEM["Transactional Memory Adapter"] -.implements.-> UOW
    MEM --> HISTORY["History Read Port"]
    MEM --> OUTBOX["Outbox records"]
    CMD["cmd/api"] --> HTTP
    CMD --> MEM
~~~

## Scope

Solution có:

- Account/Withdraw/Deposit behavior;
- Transfer record/history;
- explicit Unit of Work;
- durable-semantics idempotency model trong cùng transaction;
- outbox record trong cùng transaction;
- POST /transfers và GET /accounts/{id}/transfers;
- structured startup log;
- tests.

Memory adapter chỉ là executable profile. Production profile cần thay bằng PostgreSQL, Kafka publisher và optional Redis read cache. Mini-banking chứng minh pgx adapter/lock; Lab 09 chứng minh consumer/outbox worker.

## Problem

Starter là God handler mutate global maps. Nhiệm vụ refactor từng lát, luôn giữ tests xanh.

## Yêu cầu

1. Private Account state và domain errors.
2. Command có IdempotencyKey.
3. Same key + same request replay result; same key + khác request conflict.
4. Accounts, Transfer, idempotency và outbox commit atomically.
5. History là read projection, không load Aggregate.
6. HTTP unknown error safe.
7. Composition root chọn concrete adapters.
8. Race/vet pass.

## Các bước

1. Characterization test starter.
2. Extract Account.
3. Extract TransferMoney command.
4. Define explicit UnitOfWork/Transaction ports.
5. Implement copy-on-write memory transaction.
6. Add Transfer record.
7. Add idempotency request hash.
8. Append outbox event in same transaction.
9. Add HTTP and history query.
10. Build graph and test duplicate request.
11. Thiết kế PostgreSQL schema/locking.
12. Thiết kế outbox worker/Kafka consumer.

## Expected behavior

- First transfer returns created ID.
- Retry same key/body returns same ID, không đổi balance lần hai.
- Same key/different body returns conflict.
- Domain rejection không có history/outbox.
- History liệt kê durable transfer projection.

## Test và chạy

~~~bash
cd starter && go test ./...
cd ../solution
go test -race ./...
go vet ./...
go run ./cmd/api
~~~

## Questions

1. Vì sao UnitOfWork callback nhận Transaction interfaces explicit?
2. Outbox row nhưng chưa có publisher guarantee gì?
3. Idempotency hash cần canonical hóa ra sao?
4. History read model vì sao không trả Account Aggregate?
5. Memory adapter khác PostgreSQL ở isolation/locking/durability nào?
6. Khi scale 5.000 TPS, hot account cần đổi model gì?

## Challenge production profile

- PostgreSQL implementation + migrations/contract tests.
- Outbox worker publish Kafka + inbox consumer.
- Redis cache history with stale contract.
- OpenTelemetry spans/Prometheus metrics.
- graceful shutdown/readiness.
- double-entry ledger/reconciliation.

## Solution explanation

Solution ưu tiên explicit guarantee hơn framework: UnitOfWork cung cấp transaction-scoped capabilities nên không giấu pgx.Tx trong context; memory adapter clone toàn state và publish snapshot chỉ khi callback success. Đây không phải claim production durability. Hãy so sánh với context-based pgx Transactor của mini-banking và giải thích trade-off.
