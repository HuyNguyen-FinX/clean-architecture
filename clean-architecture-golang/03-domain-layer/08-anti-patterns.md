# Domain Modeling Anti-patterns

Anti-pattern chỉ có ý nghĩa khi thấy consequence và context. Một struct public fields không luôn sai; nó sai khi model cần bảo vệ behavior nhưng caller có thể bypass.

## 1. Anemic Domain Model Trong Domain Giàu Rule

```go
type Account struct {
	Balance int64
	Status  string
}

type AccountService struct{}

func (AccountService) Withdraw(a *Account, amount int64) error {
	// toàn bộ rule ở service
}
```

Hậu quả:

- Bất kỳ caller nào cũng mutate state.
- Rule bị copy giữa services/adapters.
- Entity lifecycle không hiện trong API.
- Test thường tập trung service mocks hơn invariant.

Refactor theo behavior có rủi ro cao trước:

```go
func (a *Account) Withdraw(amount Money) error
```

Không cần chuyển mọi calculation vào Entity trong một commit.

### Khi Anemic Data Hợp Lý

DTO, read model, ETL row hoặc CRUD reference data có thể chỉ là data. Đừng gắn methods vô nghĩa để tránh nhãn “anemic”.

## 2. Public Field + Validation Ở Handler

```go
if req.Amount > account.Balance {
	return conflict
}
account.Balance -= req.Amount
```

Kafka consumer hoặc batch path có thể quên validation. Domain truth bị đồng nhất với một transport.

Correct: handler validate syntax; domain behavior enforce invariant.

## 3. Setter-driven Model

```go
account.SetBalance(account.Balance().Sub(amount))
account.SetStatus(StatusActive)
```

Setter giữ field private nhưng không giữ domain intent. Caller orchestration low-level steps và có thể bỏ một guard.

Ưu tiên `Withdraw`, `Freeze`, `Approve`, `Confirm`, `Cancel`.

Setter có thể hợp lý cho property không có transition semantics, nhưng hãy hỏi business action đang bị mất tên gì.

## 4. Primitive Obsession

```go
func Transfer(from string, to string, amount int64, currency string) error
```

Rủi ro:

- Đảo from/to compile vẫn pass.
- Amount/currency tách rời.
- Validation lặp.
- Unit/scale mơ hồ.

Value Objects/command gom meaning:

```go
type TransferMoneyCommand struct { /* boundary primitives */ }
type AccountID string
type Money struct { /* amount + currency */ }
```

Không wrapper mọi string. Chỉ tạo type khi có meaning/rule hoặc nguy cơ nhầm đáng kể.

## 5. ORM Entity Là Domain Entity

```go
type Account struct {
	ID        string         `gorm:"primaryKey" json:"id"`
	Balance   int64          `gorm:"column:balance"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
```

Hậu quả trong domain-rich flow:

- Export fields để ORM map làm invariant dễ bypass.
- Soft delete, nullable/time representation đi vào domain dù không có meaning.
- API/persistence schema cùng shape.
- ORM hooks có thể chứa hidden behavior.

Tách `accountRow` trong adapter và rehydrate domain khi coupling cost đủ lớn.

### Khi Reuse ORM Struct Chấp Nhận Được

Tiny CRUD, một adapter, ít behavior và team chấp nhận coupling. Ghi trigger để tách khi rule/consumer tăng.

## 6. Aggregate Theo Object Graph

“Customer có Accounts nên Customer chứa toàn bộ Accounts”:

```go
type Customer struct {
	Accounts []*Account
}
```

Nếu profile update không cần balance consistency, aggregate này load/lock quá rộng. Association không đồng nghĩa consistency boundary.

Model cross-aggregate reference bằng IDs và repositories/application orchestration khi lifecycle độc lập.

## 7. Aggregate Root Marker Interface

```go
type AggregateRoot interface {
	IsAggregateRoot()
}
```

Marker không enforce invariant, repository boundary hay ownership. Nó có thể hữu ích cho generic event plumbing nội bộ, nhưng đừng coi việc implement marker là hoàn thành aggregate design.

## 8. Generic Repository Điều Khiển Domain

```go
type Repository[T any] interface {
	Create(context.Context, T) error
	Update(context.Context, T) error
	Delete(context.Context, string) error
}
```

Domain-rich use case cần semantics như load-for-update, pending transfer, save aggregate/version. Generic CRUD có thể:

- Che mất intent.
- Cho update child trực tiếp.
- Giả vờ stores có cùng consistency.
- Trả partial models.

Generic helper implementation bên trong adapter có thể tái sử dụng SQL plumbing; public port vẫn nên nói language của consumer.

## 9. Domain Service Thành God Service

```go
type DomainService struct {
	db     *sql.DB
	kafka  *kafka.Writer
	logger *zap.Logger
}
```

Service có infrastructure dependencies và workflow nên không phải pure Domain Service. Tách application orchestration, ports và domain calculation.

Không phải tên package `domain` làm code trở thành domain.

## 10. `context.Context` Trong Mọi Method

```go
func (m Money) Add(ctx context.Context, other Money) (Money, error)
func (a *Account) Withdraw(ctx context.Context, amount Money) error
```

Phép tính nhỏ không có I/O/deadline semantics. Context làm API noisy và dễ bị dùng làm service locator.

Context hợp lý ở application/repository/gateway boundary. Computation dài cần cancellation là ngoại lệ có chủ đích.

## 11. Domain Import Logger/Tracer

```go
func (a *Account) Withdraw(ctx context.Context, amount Money) error {
	span := trace.SpanFromContext(ctx)
	logger.Info("withdrawing")
}
```

Domain bị observability SDK chi phối. Instrument use case/adapter spans; record domain outcome/error attributes ở boundary. Pure domain test không cần logger mock.

Nếu domain decision cần audit fact, return/record Domain Event; operational logging khác audit business record.

## 12. Event Cho Mọi Field Change

```text
BalanceChanged
StatusFieldChanged
UpdatedAtChanged
```

Technical events expose implementation và tạo consumer coupling. Event nên có domain meaning như `MoneyWithdrawn` hoặc `AccountFrozen`, nếu có consumer/audit need.

## 13. Constructor Bypass

Adapter dùng reflection/unsafe/composite literal để map private state hoặc có `newAccountUnsafe` không validate. Invalid data vào core rồi fail ở vị trí xa nguồn.

Rehydration factory phải validate. Nếu legacy data cần nạp, viết explicit migration/legacy model path và theo dõi, không biến unsafe path thành default.

## 14. Error Có HTTP Status

```go
type InsufficientBalanceError struct {
	StatusCode int
}
```

Domain event/error bị transport semantics chi phối. HTTP adapter map domain outcome; gRPC/Kafka adapter có mapping khác.

## 15. “Rich” Bằng Cách Nhồi Mọi Thứ Vào Entity

Entity gọi repository, pricing API, email, Kafka và metrics không phải rich model; nó là God Object.

Rich nghĩa là state + behavior/invariant liên quan chặt, không phải nhiều dependencies.

## Review Checklist

- Business transition có tên hay caller set fields?
- Constructor/rehydration có thể tạo invalid state không?
- Rejected method có giữ state unchanged không?
- Value Object có alias mutable data không?
- Aggregate được chọn theo invariant hay association/table?
- Repository có cho bypass root không?
- Domain có import transport/storage/observability không?
- Service đang chứa pure rule hay I/O workflow?
- Domain event có meaning hay chỉ field change?
- Abstraction nào có thể bỏ mà không mất rule/boundary?

## Mastery Questions

1. Tại sao public field không luôn là anti-pattern?
2. Setter khác behavior method ở intent/invariant thế nào?
3. Generic repository có thể dùng ở đâu mà không làm yếu domain port?
4. ORM entity reuse hợp lý trong context nào?
5. Aggregate marker interface không chứng minh điều gì?
6. Một Entity có nhiều methods/dependencies có tự động là rich domain model không?
7. Operational log và Domain Event khác nhau thế nào?
