# 02 Dependency Rule

Chapter này đang ở trạng thái `[IN PROGRESS]`: đây là chủ đề trọng tâm và sẽ được mở rộng thêm bằng PostgreSQL, Kafka, Redis, gRPC và transaction case study.

## Tại sao cần học?

Dependency Rule là phần dễ bị hiểu sai nhất của Clean Architecture. Nhiều người nhìn runtime flow rồi kết luận use case phụ thuộc database. Thực tế cần phân biệt rõ:

```text
Runtime call direction != Source-code dependency direction
```

## Vấn đề

Nếu application import trực tiếp `pgx`, `gin`, Kafka client hoặc Redis client, use case bắt đầu biết quá nhiều detail. Khi detail đổi, policy phải đổi theo.

Ví dụ xấu:

```go
type TransferMoneyUseCase struct {
	db *pgxpool.Pool
}
```

Use case lúc này không chỉ orchestration transfer. Nó còn biết persistence mechanism.

## Mental Model

Use case nên nói:

```text
Tôi cần load account.
Tôi cần save account.
Tôi cần chạy một boundary transaction.
```

Nó không nên nói:

```text
Tôi cần pgx pool.
Tôi cần SQL query này.
Tôi cần HTTP status code này.
Tôi cần Redis key này.
```

## Golang Implementation

Application định nghĩa port nó cần:

```go
type AccountRepository interface {
	FindByID(ctx context.Context, id domain.AccountID) (*domain.Account, error)
	Save(ctx context.Context, account *domain.Account) error
}
```

Infrastructure implement port:

```go
type AccountRepository struct {
	pool *pgxpool.Pool
}
```

Composition Root lắp ráp:

```go
repo := postgres.NewAccountRepository(pool)
uc := application.NewTransferMoneyUseCase(repo, tx)
```

Application không import package `postgres`.

## Interface Nên Nằm Ở Đâu?

Không có một câu trả lời áp dụng cho mọi project. Nhưng rule thực dụng trong Go:

- Nếu interface thể hiện nhu cầu của use case, đặt gần application/use case.
- Nếu interface là contract domain-level thật sự, đặt trong domain.
- Nếu interface chỉ để che một implementation trong cùng package, cân nhắc bỏ.
- Nếu producer định nghĩa interface lớn để mọi consumer dùng chung, cẩn thận interface sẽ phình và coupling ngược.

Repository interface cho `TransferMoneyUseCase` thường đặt ở application vì use case là consumer.

## Anti-pattern

- `application` import `infrastructure/postgres`.
- `domain` chứa `json`, `db`, `gorm`, `bson` tag khi domain không cần biết transport/persistence.
- Interface có 20 method vì gom nhu cầu của nhiều use case.
- Tạo interface cho mọi service dù không có boundary.
- Dùng `context.Context` trong entity domain.

## Production Scenario

Trong banking transfer, transaction thật cần row locking. PostgreSQL adapter có thể dùng:

```text
SELECT ... FOR UPDATE
```

Nhưng use case chỉ cần diễn đạt intent:

```text
Load account để transfer trong transaction.
```

Bạn có thể để port là `FindByIDForUpdate` nếu locking là intent application cần. Nhưng domain vẫn không biết SQL lock.

## Mastery Check

- [ ] Tôi đọc import graph để kiểm tra dependency direction.
- [ ] Tôi biết runtime gọi PostgreSQL không đồng nghĩa application import PostgreSQL.
- [ ] Tôi biết port nên được định nghĩa theo nhu cầu consumer.
- [ ] Tôi biết khi nào interface đang làm code phức tạp hơn.
