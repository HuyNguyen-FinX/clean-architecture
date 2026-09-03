# Mini Banking Example

Ví dụ này là một lát cắt nhỏ của hệ thống banking:

- `Account` và `Money` nằm trong Domain Layer.
- `TransferMoneyUseCase` nằm trong Application Layer.
- `AccountRepository` là port do application cần.
- `memory.Repository` là infrastructure adapter implement port.
- HTTP handler là delivery adapter.
- `cmd/api/main.go` là Composition Root.

## Cấu trúc

```text
cmd/api/main.go
internal/account/
├── domain/
├── application/
├── infrastructure/memory/
└── delivery/http/
```

## Dependency Direction

Source-code dependency:

```text
delivery/http -> application -> domain
infrastructure/memory -> application/domain
cmd/api -> delivery/http + application + infrastructure/memory
```

Không có dependency:

```text
domain -> http
domain -> memory repository
application -> delivery/http
application -> infrastructure/memory
```

Runtime flow:

```text
HTTP request
↓
Handler
↓
TransferMoneyUseCase
↓
AccountRepository interface
↓
memory.Repository object
```

Runtime đi xuống adapter thật, nhưng source code của application chỉ biết interface.

## Chạy Test

```bash
go test ./...
```

Các test minh họa:

- Domain test kiểm tra invariant withdraw mà không mock gì.
- Use case test dùng in-memory repository như fake adapter.
- HTTP test kiểm tra mapping request/error mà không cần database.

## Chạy API

```bash
go run ./cmd/api
```

Gửi transfer:

```bash
curl -X POST http://localhost:8080/transfers \
  -H 'Content-Type: application/json' \
  -d '{"from_account_id":"A-100","to_account_id":"B-200","amount":500000,"currency":"VND"}'
```

## Vì sao chưa dùng PostgreSQL?

Increment đầu tiên dùng memory adapter để làm rõ dependency. PostgreSQL sẽ được thêm ở phase database/transaction để tránh trộn hai bài học:

- Boundary và Dependency Rule.
- Driver, SQL mapping, pool, transaction, locking.

Trong production, `memory.Repository` sẽ được thay bằng `postgres.Repository`, nhưng `TransferMoneyUseCase` không cần đổi nếu port giữ nguyên.
