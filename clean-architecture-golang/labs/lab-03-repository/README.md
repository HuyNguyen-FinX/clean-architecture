# Lab 03: Repository contract và Memory Adapter

Thời lượng gợi ý: 75-120 phút. Lab này không yêu cầu PostgreSQL; mục tiêu là thiết kế semantics trước khi viết SQL.

## Mục tiêu

- phân biệt persistence port với concrete store;
- để application consumer sở hữu interface nhỏ;
- bảo vệ Aggregate khỏi pointer alias;
- chuẩn hóa not-found và context semantics;
- viết reusable Repository contract test;
- hiểu rõ điều Memory Adapter chưa thể chứng minh.

## Kiến thức cần

- [Repository Pattern](../../05-repository-pattern/README.md)
- [Dependency Rule](../../02-dependency-rule/README.md)
- mutex, context cancellation và Go interface satisfaction.

## Architecture diagram

~~~mermaid
flowchart LR
    USECASE["DepositMoney"] --> PORT["AccountRepository"]
    USECASE --> DOMAIN["Account"]
    MEMORY["Memory Repository"] -.implements.-> PORT
    MEMORY --> DOMAIN
    CONTRACT["Repository contract tests"] --> PORT
~~~

Application không import Memory Adapter. Composition root mới chọn adapter; contract suite nhìn adapter qua port.

## Problem

[Starter](./starter/) có một MemoryStore concrete được inject thẳng vào Service. Find trả chính pointer trong map. Vì vậy:

~~~text
load account
↓
mutate returned pointer
↓
stored state đã đổi dù Save chưa chạy
~~~

Trong PostgreSQL, mutation của Go object không tự update row. Fake và production đang có guarantee khác nhau, làm use-case test có thể pass sai.

## Yêu cầu

1. Interface AccountRepository nằm phía application và chỉ có FindByID/Save.
2. Memory Adapter implement interface bằng structural typing.
3. FindByID và Save tôn trọng context đã hủy.
4. Không tìm thấy account trả stable application.ErrAccountNotFound.
5. Adapter clone khi seed, load và save.
6. Repository an toàn cho concurrent access.
7. Contract suite kiểm tra round trip, not found, detached read và detached save.
8. Một compile-time assertion xác minh adapter implement port.

## Các bước

### Bước 1: chạy và quan sát starter

~~~bash
cd starter
go test ./...
~~~

Test PointerAlias mô tả bug hiện tại và vẫn pass vì đó là behavior baseline. Sau khi refactor, behavior đúng phải đảo lại: mutate bản đã load không làm store đổi trước Save.

### Bước 2: định nghĩa port từ consumer

Đọc DepositMoney và chỉ thêm operation nó cần. Không thêm Delete/List hoặc generic filter.

### Bước 3: xác định ownership

Viết ra ai sở hữu pointer ở ba điểm: seed, FindByID và Save. Clone qua RehydrateAccount để constructor vẫn kiểm tra invariant.

### Bước 4: stable error semantics

Memory Adapter không trả lỗi map-specific. Sau này PostgreSQL Adapter sẽ map pgx.ErrNoRows về cùng ErrAccountNotFound.

### Bước 5: contract test

Viết một hàm nhận RepositoryFactory. Mỗi adapter gọi lại suite đó. Test trên observable behavior, không nhìn vào internal map.

### Bước 6: test cancellation và race

Context đã cancel phải được trả trước khi đọc/ghi. Chạy race detector để kiểm tra mutex path.

## Expected behavior

- Seed A=100, load A, deposit 50 nhưng chưa Save: lần load mới vẫn thấy 100.
- Sau Save: lần load mới thấy 150.
- Mutate object sau Save: store vẫn giữ snapshot 150.
- ID lạ: errors.Is nhận ra ErrAccountNotFound.
- Context cancel: errors.Is nhận ra context.Canceled.

## Test

~~~bash
cd solution
go test -race ./...
go vet ./...
~~~

## Questions

1. Vì sao copy ownership là một phần của Repository semantics?
2. Contract test khác use-case unit test và PostgreSQL integration test thế nào?
3. Mutex trong Memory Adapter có giải quyết concurrency giữa nhiều service replica không?
4. Vì sao Rehydrate vẫn phải validate dữ liệu từ store?
5. Nếu Save dùng upsert, caller đang dựa vào guarantee gì mà có thể bị che?

## Challenge

- Thêm Version và optimistic conflict.
- Viết adapter file-backed nhưng giữ cùng contract.
- Viết PostgreSQL adapter, migration và chạy contract suite bằng Testcontainers.
- Thêm query port ListAccounts với cursor mà không làm AccountRepository thành God interface.

## Solution explanation

[Solution](./solution/) tách domain, application và memory adapter. Package repositorytest chứa suite dùng được cho adapter khác. Memory Repository dùng RWMutex và clone tại mọi ownership boundary. Nó vẫn không chứng minh SQL, schema, lock hoặc rollback; đó là phạm vi của Lab 04 và Lab 08.
