# Lab 11: Refactor God Handler Theo Vertical Slice

Thời lượng: 120-180 phút.

## Mục tiêu

- khóa behavior legacy;
- extract domain invariant;
- extract application use case/ports;
- move storage/outbox ra adapter;
- giữ HTTP contract;
- refactor qua bước nhỏ có test.

## Kiến thức cần

- [Refactoring](../../25-refactoring/README.md)
- [Anti-patterns](../../26-anti-patterns/README.md)

## Before/after

~~~mermaid
flowchart LR
    GOD["HTTP + state + rule + event"] --> MAP["global maps"]
    HTTP["HTTP Adapter"] --> UC["TransferMoney"]
    UC --> DOMAIN["Account"]
    UC --> UOW["UnitOfWork"]
    MEMORY["Memory Adapter"] -.implements.-> UOW
~~~

## Problem

Starter Handler giữ concrete Store/Producer, mutate public balances và publish sau writes không atomic. Characterization tests khóa current 201/409/event behavior.

## Yêu cầu

1. Không đổi external JSON/status trong refactor đầu.
2. Account fields private, Withdraw/Deposit.
3. TransferMoney không import net/http/concrete adapter.
4. Transaction/outbox intent cùng UnitOfWork.
5. HTTP map errors.
6. Tests ở domain/application/HTTP.
7. Không tạo pass-through layers.

## Steps

0. Đọc current code và list risks.
1. Characterization tests.
2. Extract Account behavior.
3. Extract command/use case.
4. Consumer-owned UoW port.
5. Move memory detail.
6. Move event thành outbox intent atomic.
7. Keep handler DTO/error.
8. Add architecture/test gaps.

Mỗi step nên là một commit xanh. Mechanical moves tách khỏi behavior change.

## Expected behavior

Valid transfer 201 + one outbox; insufficient 409 + no write/event. Application tests không cần HTTP. Domain tests không cần fake.

## Test

~~~bash
cd starter && go test -race ./...
cd ../solution && go test -race ./... && go vet ./...
~~~

## Questions

1. Step nào đổi dependency thật?
2. EventPublisher port sau DB commit còn lỗi gì?
3. Characterization test có nên giữ bug?
4. Vì sao solution không tạo Repository + DAO cùng lúc?
5. Production adapter cần thêm test gì?

## Challenge

- PostgreSQL transaction + outbox migration.
- Idempotency.
- feature-flag old/new comparison.
- architecture import test.

## Solution explanation

Solution dùng một UoW port vì atomic operation cần Accounts và outbox cùng boundary. Memory adapter copy-on-write chạy được nhưng không claim durability. Đây là một lựa chọn; Repository+Transactor riêng như mini-banking cũng hợp lý.
