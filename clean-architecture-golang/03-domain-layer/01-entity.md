# Entity Và Identity

## Problem

Hai record có cùng dữ liệu có phải cùng một object không?

```text
Account A hôm nay: balance 1.000.000, status active
Account A ngày mai: balance   500.000, status frozen
```

Field đã đổi, nhưng business vẫn nói về **cùng Account A**. Ngược lại, hai account khác nhau có thể tình cờ cùng balance/status nhưng không thể thay thế cho nhau.

Entity là object có continuity được xác định bởi identity, không phải toàn bộ current attributes.

## Ba Level

### Level 1: Trực Giác

Entity giống một người: tuổi, địa chỉ và trạng thái có thể đổi, nhưng ta vẫn theo dõi cùng người đó qua thời gian.

`Account` là Entity. `Money(500.000 VND)` thường là Value Object vì hai value cùng amount/currency có thể thay thế nhau.

### Level 2: Backend Engineer

Entity cần:

- Identity type rõ.
- Constructor/rehydration bảo vệ valid state.
- Methods biểu diễn transition.
- Equality theo identity khi business cần so sánh Entity.
- Persistence giữ identity ổn định qua load/save.

### Level 3: Modeling

Identity định hình aggregate reference và lifecycle:

- Object ngoài aggregate tham chiếu Aggregate Root ID, không giữ pointer vào internal child tùy ý.
- Event dùng identity để nói fact thuộc aggregate nào.
- Idempotency/business key có thể khác database surrogate key.
- Merge dữ liệu không được đồng nhất hai Entity chỉ vì attributes giống nhau.

## Có Field `ID` Chưa Chắc Là Entity

DTO có request ID, DB row có primary key, Kafka message có event ID. Chúng không tự động là Domain Entity.

Hỏi:

1. Business có cần theo dõi object qua thay đổi state không?
2. Hai object cùng attributes nhưng khác identity có được xem là khác không?
3. Object có lifecycle/state transition riêng không?
4. Identity đến từ domain hay chỉ để storage định vị row?

Ví dụ `ExchangeRateSnapshot` có DB ID để audit nhưng business có thể coi nó là immutable value tại một thời điểm. Classification phục vụ model, không phục vụ ORM.

## Identity Type Trong Go

Mini-banking không dùng raw string khắp nơi:

```go
type AccountID string

func NewAccountID(raw string) (AccountID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidAccountID
	}
	return AccountID(raw), nil
}
```

Lợi ích:

- Function signature nói rõ ID thuộc concept nào.
- Tránh truyền nhầm `CustomerID` vào chỗ cần `AccountID` nếu mỗi ID là distinct named type.
- Validation/normalization có một entry point.

Go vẫn cho caller convert explicit `AccountID("")`. Vì vậy aggregate constructor phải validate lại identity tại trust boundary. Private field giảm đường sai nhưng không tạo proof tuyệt đối.

## Entity Equality

Nếu cần equality, Entity thường so identity:

```go
func (a *Account) SameIdentity(other *Account) bool {
	return other != nil && a.id == other.id
}
```

Không dùng deep equality để quyết định “cùng account”:

```go
reflect.DeepEqual(accountLoadedYesterday, accountLoadedToday)
```

Deep equality trả false khi balance/status đổi, dù identity không đổi. Nó phù hợp một số test snapshot, không phải domain identity semantics.

Mini-banking hiện không thêm `SameIdentity` vì use case so `AccountID` trước khi load. Không tạo method khi không có consumer thật.

## Pointer Receiver Hay Value Receiver?

`Account` có state transition nên dùng pointer receiver:

```go
func (a *Account) Withdraw(amount Money) error
```

`Money` trả value mới và dùng value receiver:

```go
func (m Money) Add(other Money) (Money, error)
```

Đây là Go design choice phù hợp semantics:

- Entity mutation có identity/lifecycle, pointer giúp transition tác động object đang được orchestration.
- Value Object nhỏ, immutable-style, copy theo value dễ reasoning.

Không biến pointer/value receiver thành rule DDD tuyệt đối. Struct lớn hoặc chứa slice/map cần xem aliasing/copy cost và ownership.

## Creation Và Rehydration

```go
func NewAccount(id AccountID, balance, limit Money) (*Account, error) {
	return RehydrateAccount(id, balance, limit, AccountStatusActive)
}
```

Trong mini-banking, `NewAccount` là convenience constructor cho example. Production creation có thể có semantics mạnh hơn:

```go
func OpenAccount(id AccountID, currency Currency) (*Account, AccountOpened, error)
```

Nó có thể ép balance bằng zero, status active và ghi domain event.

Rehydration khôi phục Entity đã tồn tại:

```go
func RehydrateAccount(
	id AccountID,
	balance Money,
	limit Money,
	status AccountStatus,
) (*Account, error)
```

Rehydration vẫn validate. Database là nguồn dữ liệu bền vững nhưng không phải nguồn định nghĩa valid domain state. Migration lỗi hoặc manual SQL có thể tạo row sai; adapter phải nhận error thay vì đưa object hỏng vào core.

## State Transition, Không Phải Setter

Wrong:

```go
func (a *Account) SetBalance(balance Money) {
	a.balance = balance
}
```

Setter nói mechanism “gán value”, không nói intent. Caller phải tự nhớ overdraft, frozen, audit và event rule.

Correct API nói business action:

```go
func (a *Account) Withdraw(amount Money) error
func (a *Account) Deposit(amount Money) error
func (a *Account) Freeze()
func (a *Account) Activate()
```

Không phải mọi method đều cần error. `Freeze` hiện idempotent và luôn tạo valid state. Nếu business cấm freeze account đã closed hoặc cần reason, signature phải tiến hóa để thể hiện rule.

## Identity Trong Database

Database primary key và domain identity thường trùng, nhưng không bắt buộc:

```text
internal bigint PK: tối ưu join/storage
account number:     business identifier, có thể đổi/reissue
public UUID:        identifier expose ra API
```

Nếu account number có thể đổi, dùng nó làm Entity identity sẽ gây nhầm. Nếu regulation coi account number là identity vĩnh viễn, model khác. Không chọn identity chỉ vì column đang là primary key.

Repository phải bảo toàn identity:

```text
FindByID(A) -> Account{id: A}
Save(Account{id: A}) -> update A, không biến thành B
```

Optimistic version là concurrency metadata, không phải Entity identity.

## Production Scenario: Duplicate Customer Records

Hai customer records có cùng tên/ngày sinh chưa chắc cùng người. Merge theo attribute equality có thể gộp nhầm identity. Domain cần policy riêng: verified document, source system, manual review hoặc merge command có audit.

Entity concept giúp nhìn ra vấn đề, nhưng không tự giải quyết identity resolution. Database unique constraint, external identity provider và operational reconciliation vẫn cần.

## Testing Strategy

Test Entity tập trung vào transition:

- Valid transition đổi đúng state.
- Invalid transition trả đúng domain error.
- Rejected operation không mutate state.
- Boundary values.
- Constructor/rehydration reject invalid state.
- Identity không đổi sau behavior.

Không mock Entity. Tạo object thật bằng constructor và quan sát public behavior.

## Khi Nào Không Cần Entity?

- Một immutable configuration snapshot không có lifecycle.
- DTO chuyển data qua boundary.
- Read model cho màn hình/report.
- Value như `Money`, `DateRange`, `EmailAddress` được định nghĩa bởi attributes.

Gắn ID giả vào mọi struct để biến nó thành Entity làm model khó hơn mà không thêm meaning.

## Mastery Questions

1. Hai Account cùng balance/currency có bằng nhau không? Equality nào đang hỏi?
2. Một database row có primary key có luôn là Domain Entity không?
3. Vì sao account number có thể không phải identity ổn định?
4. Tại sao Entity thường dùng pointer receiver trong Go, nhưng đây không phải luật tuyệt đối?
5. Constructor và rehydration khác nhau về semantics nào?
6. Setter làm caller gánh invariant ra sao?
7. Tại sao private field vẫn không loại bỏ nhu cầu validate khi rehydrate?
