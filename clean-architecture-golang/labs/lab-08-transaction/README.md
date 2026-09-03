# Lab 08: Transaction Boundary và Rollback

Thời lượng: 90-150 phút. Lab dùng transactional memory adapter để mọi máy chạy được; phần cuối chạy cùng scenario trên PostgreSQL mini-banking.

## Mục tiêu

- tái hiện partial write;
- đặt transaction quanh toàn use case;
- implement closure Transactor;
- chứng minh rollback khi Save thứ hai fail;
- truyền đúng transaction context;
- liên hệ local atomicity với row locking/idempotency.

## Kiến thức cần

- [Transaction Management](../../11-transaction-management/README.md)
- [Use Case](../../04-usecase-application-layer/README.md)
- context, mutex và copy ownership.

## Diagram

~~~mermaid
flowchart TB
    UC["TransferMoney"]
    subgraph TX["WithinTransaction"]
      LOAD["load A + B"] --> DOMAIN["withdraw + deposit"]
      DOMAIN --> SAVE["save A + B"]
    end
    UC --> TX
    MEM["Transactional Memory Adapter"] -.implements.-> TX
~~~

## Problem

Starter save A thành công rồi giả lập save B thất bại. Snapshot cuối bị lệch. Transaction solution tạo working copy, chỉ publish khi closure trả nil.

## Yêu cầu

1. Transactor port thuộc application.
2. Validation command chạy trước transaction.
3. Mọi Find/Save dùng transaction context.
4. Error ở bất kỳ bước nào bỏ toàn bộ staged state.
5. Success commit cả hai balance.
6. Nested/wrong-store transaction context bị reject hoặc xử lý rõ.
7. Race test pass.
8. Giải thích memory mutex khác PostgreSQL row lock thế nào.

## Các bước

1. Chạy starter và xác minh partial state.
2. Viết use case closure.
3. Adapter lock, clone store thành staged state.
4. Đưa private transaction handle vào context.
5. Repository chọn staged state khi có transaction.
6. Chỉ replace durable map khi callback thành công.
7. Inject failure ở Save thứ hai và assert rollback.
8. Chạy PostgreSQL integration test của mini-banking.

## Expected behavior

- Transfer 300: A 1.000 → 700, B 100 → 400.
- Save B lỗi: cả hai giữ nguyên.
- Invalid amount không mở transaction.
- Concurrent transfer không data race.

## Test

~~~bash
cd starter && go test ./...
cd ../solution && go test -race ./... && go vet ./...

cd ../../../examples/mini-banking
TEST_DATABASE_URL=postgres://... go test -race ./internal/account/infrastructure/postgres
~~~

## Questions

1. Vì sao callback transaction thuộc use case?
2. Memory snapshot rollback không mô phỏng điều gì của PostgreSQL?
3. Vì sao dùng context ban đầu trong Repository call làm atomicity vỡ?
4. Commit thành công nhưng response mất cần cơ chế nào?
5. Network call trong closure nguy hiểm ra sao?

## Challenge

- Reject nested transaction.
- Thêm optimistic version/conflict.
- Thêm retry bounded, bảo đảm reload object mỗi attempt.
- Persist idempotency record cùng balance.
- Thiết kế outbox write trong transaction.

## Solution explanation

Solution dùng copy-on-write transaction để quan sát atomic commit/rollback mà không cần database. Đây là teaching adapter, serialize toàn store và không đại diện isolation/lock behavior PostgreSQL. Mini-banking cung cấp pgx transaction + SELECT FOR UPDATE + integration test thật.
