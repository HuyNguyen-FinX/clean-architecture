# Mini Banking Example

Ví dụ này là một lát cắt nhỏ của hệ thống banking:

- `Account`, `AccountStatus`, `Currency` và `Money` nằm trong Domain Layer.
- `TransferMoneyUseCase` nằm trong Application Layer.
- `AccountRepository` là port do application cần.
- `memory.Repository` là infrastructure adapter implement port.
- HTTP handler là delivery adapter.
- `cmd/api/main.go` là Composition Root.
- Architecture test kiểm import direction của core packages.

## Trạng Thái Hiện Tại: V3

Project đang ở V3 của learning evolution: domain + use case + repository port/memory adapter. Nó chạy được, nhưng chưa có production transaction/persistence guarantee.

| Guarantee | Hiện tại | Ghi chú |
|---|---|---|
| Money currency/equality/overflow | Có | Domain code và tests |
| Account overdraft/frozen invariant | Có | Constructor và transitions cùng validate |
| HTTP -> command mapping | Có | `net/http` adapter + `httptest` |
| Core import direction | Có guard | `internal/architecture/dependency_test.go` |
| Sender/receiver atomic save | Chưa | `NoopTransactor` không rollback memory writes |
| Concurrent lost-update protection | Chưa | Chưa có row lock/optimistic version |
| PostgreSQL persistence | Chưa | Memory adapter chỉ dùng cho learning slice |
| Transfer record/history | Chưa | Sẽ thêm cùng transaction model |
| Request idempotency | Chưa | Retry có thể thực hiện transfer lần nữa |
| Kafka/outbox | Chưa | Không có event delivery guarantee |
| Production observability/lifecycle | Chưa | Chưa graceful shutdown/tracing/metrics |

## Evolution V1-V11

```text
V1  Money + Account domain/invariants
V2  TransferMoney application use case
V3  Repository port + memory adapter + HTTP slice       <- hiện tại
V4  PostgreSQL row mapping + integration tests
V5  Real transaction + locking/version + Transfer record
V6  Strict HTTP contract + lifecycle/error taxonomy
V7  Full test strategy across boundaries
V8  Idempotency và safe retry
V9  Kafka producer/consumer adapters
V10 Transactional outbox
V11 Logging, metrics, tracing và production operations
```

Mỗi version chỉ được nhận guarantee sau khi có implementation và test tương ứng. Interface hoặc folder name không được tính là guarantee.

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

- Domain test kiểm tra Money, creation invariant, withdraw boundary và frozen transition mà không mock gì.
- Use case test dùng in-memory repository như fake adapter.
- HTTP test kiểm tra mapping request/error mà không cần database.
- Architecture test parse production imports để chặn outer dependency trong domain/application.

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

Trong production, `memory.Repository` sẽ được thay bằng `postgres.Repository`. `TransferMoneyUseCase` không cần đổi nếu capability/consistency contract giữ nguyên. Nếu persistence mới không cung cấp cùng transaction semantics, application workflow có thể phải đổi; boundary làm khác biệt đó explicit chứ không giả vờ mọi database tương đương.
