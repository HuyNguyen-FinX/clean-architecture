# 04 Use Case / Application Layer

> Application Layer biến một mục tiêu của actor/system thành workflow có boundary rõ. Nó orchestration domain behavior và ports, nhưng không biết JSON, HTTP status, SQL query, Kafka client hoặc Redis key.

## Kết Quả Học Tập

Sau chapter này, bạn có thể:

- Phân biệt Application Service/Use Case với Domain Service.
- Thiết kế command/result không phụ thuộc transport.
- Đặt validation, authorization, transaction và idempotency đúng scope.
- Định nghĩa outgoing ports nhỏ theo nhu cầu use case.
- Truyền `context.Context` qua I/O flow mà không đưa request lifecycle vào Entity.
- Test orchestration bằng fake/stub/spy có chủ đích.
- Phân tích partial failure, timeout, retry và ambiguous outcome.
- Nhận ra service pass-through hoặc God Service.

## 1. Problem: Use Case Bị Mất Trong Handler

Một transfer endpoint thường phải:

```text
parse JSON
authenticate actor
validate command
load sender/receiver
withdraw/deposit
save atomic
record transfer
publish intent
map response
```

Nếu handler sở hữu toàn bộ, thêm Kafka consumer hoặc gRPC sẽ copy workflow. Nếu Entity sở hữu toàn bộ, `Account` phải biết repository, transaction và account khác. Nếu repository sở hữu toàn bộ, persistence adapter trở thành nơi quyết định business flow.

Application Use Case là boundary cho **một mục tiêu có ý nghĩa với actor**:

```text
TransferMoney
ApproveLoan
CapturePayment
ConfirmOrder
FreezeAccount
```

Nó giữ thứ tự orchestration và application policy ở một nơi tái sử dụng được bởi nhiều delivery adapters.

## 2. Ba Level

### Level 1: Trực Giác

Use Case là đạo diễn:

- Nhận intent đã được adapter dịch.
- Lấy domain objects cần thiết.
- Gọi behavior đúng thứ tự.
- Phối hợp side effects/transaction.
- Trả outcome không gắn transport.

Đạo diễn không tự diễn mọi vai. Invariant vẫn ở Domain; SQL vẫn ở repository adapter.

### Level 2: Backend Engineer

Một Use Case Go thường là struct với constructor và một method:

```go
type TransferMoneyUseCase struct {
	accounts AccountRepository
	tx       Transactor
}

func (uc *TransferMoneyUseCase) Execute(
	ctx context.Context,
	cmd TransferMoneyCommand,
) error
```

Dependencies được inject. Command là plain Go data. Method nhận context vì repository/transaction có I/O.

### Level 3: Architecture

Application Layer sở hữu:

- Workflow boundary.
- Capability contracts cần từ outside.
- Transaction/idempotency semantics ở mức use case.
- Authorization/application validation.
- Outcome taxonomy ổn định cho callers.

Nó không sở hữu mechanism: SQL transaction object, HTTP header parsing, Kafka retries hoặc provider SDK.

## 3. Responsibility Map

| Concern | Owner thường gặp | Ví dụ |
|---|---|---|
| JSON/body limit | HTTP adapter | decode `transferRequest` |
| Actor extraction | Middleware/delivery | JWT -> ActorID |
| Actor được transfer account này? | Application policy/authorization port | ownership/permission |
| Amount positive/currency valid | Domain Value Object | `NewPositiveMoney` |
| Balance/overdraft/frozen | Account | `Withdraw` |
| Load A/B và save cả hai | Use Case | orchestration |
| BEGIN/COMMIT/rollback API | DB Transactor adapter | pgx implementation |
| Row locking/version | Repository adapter + port semantics | `FOR UPDATE`/version |
| HTTP 409 | HTTP adapter | map outcome |
| Kafka serialization/topic | Kafka adapter | integration event |

Một concern có thể cần nhiều layer phối hợp. Idempotency policy thuộc application; unique constraint là infrastructure enforcement; header extraction thuộc HTTP.

## 4. Command Và Result

HTTP DTO:

```go
type transferRequest struct {
	FromAccountID string `json:"from_account_id"`
	ToAccountID   string `json:"to_account_id"`
	Amount        int64  `json:"amount"`
	Currency      string `json:"currency"`
}
```

Application command:

```go
type TransferMoneyCommand struct {
	FromAccountID string
	ToAccountID   string
	Amount        int64
	Currency      string
}
```

Hai struct hiện giống gần nhau nhưng ownership khác:

- HTTP DTO có JSON contract/version.
- Command có application intent và có thể được tạo từ HTTP, gRPC, Kafka hoặc CLI.

Khi trivial CRUD chỉ có một caller, reuse struct có thể hợp lý. Tách không phải mặc định; tách khi contract có reason to change độc lập.

Result nên trả information caller cần, không trả transport response:

```go
type TransferMoneyResult struct {
	TransferID string
	Status     string
}
```

HTTP adapter quyết định encode JSON/status 201. Kafka consumer có thể chỉ cần ack. Mini-banking V3 chưa có Transfer Entity nên `Execute` mới trả `error`; result sẽ xuất hiện khi V5 thêm record/history.

## 5. Walkthrough `TransferMoneyUseCase`

Full code: [`transfer_money.go`](../examples/mini-banking/internal/account/application/transfer_money.go).

### Step 1: Parse Boundary Primitives Thành Domain Values

```go
fromID, err := domain.NewAccountID(cmd.FromAccountID)
if err != nil {
	return err
}

toID, err := domain.NewAccountID(cmd.ToAccountID)
if err != nil {
	return err
}
if fromID == toID {
	return domain.ErrSameAccountTransfer
}

amount, err := domain.NewPositiveMoney(cmd.Amount, cmd.Currency)
if err != nil {
	return err
}
```

Invalid command bị reject trước khi transaction/I/O. Rule same account liên quan hai IDs của operation nên nằm ở application trong model hiện tại.

### Step 2: Mở Transaction Boundary

```go
return uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
	// toàn bộ reads/writes cần atomicity
})
```

Use Case biết scope cần atomic; adapter biết cách BEGIN/COMMIT/ROLLBACK.

### Step 3: Load Aggregates

```go
sender, err := uc.accounts.FindByID(txCtx, fromID)
if err != nil {
	return err
}
receiver, err := uc.accounts.FindByID(txCtx, toID)
if err != nil {
	return err
}
```

Production adapter phải định nghĩa locking semantics. Interface `FindByID` hiện chưa nói load-for-update, nên V5 sẽ làm contract explicit.

### Step 4: Gọi Domain Behavior

```go
if err := sender.Withdraw(amount); err != nil {
	return err
}
if err := receiver.Deposit(amount); err != nil {
	return err
}
```

Use Case không tự tính balance/overdraft. Error path dừng trước save.

### Step 5: Save

```go
if err := uc.accounts.Save(txCtx, sender); err != nil {
	return err
}
return uc.accounts.Save(txCtx, receiver)
```

Closure trả error để transactor rollback. Memory `NoopTransactor` hiện không rollback; README công bố rõ limitation này. PostgreSQL adapter phải test rollback thật.

## 6. Application Service Vs Domain Service

| Application Service | Domain Service |
|---|---|
| Orchestration actor goal | Pure business rule không thuộc một Entity |
| Nhận context + command | Nhận domain values/entities |
| Gọi repositories/gateways | Không I/O |
| Quyết định transaction/idempotency flow | Tính/đánh giá policy domain |
| Có thể gọi Domain Service | Không gọi Application Service |

`TransferMoneyUseCase` là Application Service. `TransferFeePolicy.Fee(amount, sameBank)` có thể là Domain Service.

Wrong:

```go
type TransferService struct {
	db     *sql.DB
	kafka  *kafka.Writer
	logger *zap.Logger
}
```

Tên `Service` không làm code thuộc Domain. Dependencies và responsibility mới quyết định.

## 7. Outgoing Ports

```go
type AccountRepository interface {
	FindByID(ctx context.Context, id domain.AccountID) (*domain.Account, error)
	Save(ctx context.Context, account *domain.Account) error
}

type Transactor interface {
	WithinTransaction(
		ctx context.Context,
		fn func(context.Context) error,
	) error
}
```

Port nhỏ theo nhu cầu use case:

- Không `Delete`, `List`, `Count` nếu transfer không dùng.
- Không `pgx.Tx` hoặc SQL row.
- Domain language ở signatures.
- Transaction mechanism được invert.

Không cần interface cho every helper. Pure calculator concrete type có thể inject trực tiếp.

## 8. Constructor Và Dependency Invariant

Current constructor:

```go
func NewTransferMoneyUseCase(
	accounts AccountRepository,
	transactor Transactor,
) *TransferMoneyUseCase {
	if accounts == nil {
		panic("application: nil account repository")
	}
	if transactor == nil {
		panic("application: nil transactor")
	}
	return &TransferMoneyUseCase{accounts: accounts, transactor: transactor}
}
```

Dependencies thiếu là programming/configuration error ở composition time, không phải runtime business outcome. Panic ở constructor là lựa chọn phù hợp cho example; một constructor trả error cũng hợp lý nếu dependency được build từ dynamic plugin/config.

Phiên bản cũ tự thay nil transactor bằng no-op. Điều đó nguy hiểm:

```text
wiring quên transaction
        ↓
application âm thầm tiếp tục
        ↓
tests/happy path pass
        ↓
production partial save
```

No-op implementation giờ nằm ở memory adapter và tên nói rõ không có rollback guarantee. Production composition không được chọn nhầm mà không thấy.

Go interface có typed-nil nuance: một interface chứa `(*Repo)(nil)` không bằng nil. Constructor cần contract/documentation; tránh inject typed nil. Khi cần, concrete constructors phải không trả typed nil success.

## 9. Context Boundary

```text
HTTP/Kafka context
    ↓
UseCase.Execute(ctx, cmd)
    ↓
Transactor/Repository/Gateway(ctx, ...)
```

Use Case không tạo `context.Background()` giữa flow vì sẽ làm mất cancellation/deadline/trace. Nó có thể tạo timeout child khi application policy cần budget riêng:

```go
gatewayCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
defer cancel()
```

Domain method không cần context cho pure transition:

```go
sender.Withdraw(amount)
```

Không nhét dependencies vào `ctx.Value`; constructor giữ chúng explicit.

## 10. Transaction Boundary

Tại sao Use Case thường gần transaction boundary?

```text
Repository chỉ biết Save(Account)
Use Case biết operation cần:
  load A
  load B
  debit
  credit
  save A
  save B
  record Transfer
```

Nếu mỗi repository method tự commit, không component nào thấy toàn bộ atomic operation.

Application quyết định scope; adapter implement mechanism. Nhưng network call trong DB transaction có thể giữ lock lâu. Capture payment, inventory reservation hoặc Kafka publish cần outbox/saga/reconciliation, không kéo transaction xuyên network.

## 11. Idempotency Là Application Policy

Client timeout sau commit rồi retry cùng intent. Transaction không nhận ra duplicate request.

Use Case production có thể:

```text
validate idempotency key
BEGIN
insert operation key (unique)
if existed: return stored result
execute transfer
save result + outbox
COMMIT
```

HTTP adapter lấy key từ header; application định nghĩa semantics/scope; DB unique constraint enforce concurrency. Không đặt toàn bộ idempotency vào middleware nếu middleware không biết business result/transaction.

## 12. Error Taxonomy

Use Case có thể propagate domain error hiện tại:

```text
ErrInvalidAmount
ErrInsufficientBalance
ErrAccountFrozen
```

Infrastructure failures nên được wrap giữ cause:

```go
return fmt.Errorf("load sender %s: %w", fromID, err)
```

Khi application lớn, stable application errors/outcomes giúp delivery không import mọi domain package. Không convert mọi error thành string. HTTP adapter mới map status và không expose raw DB message.

## 13. Testing Use Case

Full tests: [`transfer_money_test.go`](../examples/mini-banking/internal/account/application/transfer_money_test.go).

### Fake Repository

Fake giữ Account copies trong map, record saved IDs và inject error tại save thứ N. Nó mô phỏng capability domain-level, không mock SQL calls.

### Recording Transactor

```go
func (tx *recordingTransactor) WithinTransaction(
	ctx context.Context,
	fn func(context.Context) error,
) error {
	tx.calls++
	return fn(context.WithValue(ctx, transactionContextKey{}, true))
}
```

Fake repository record nếu bị gọi ngoài transaction context. Test chứng minh reads/writes nằm trong boundary.

### Behavior Matrix

- Happy path chuyển đúng balances.
- Invalid command không mở transaction.
- Domain reject không save.
- Repository error được propagate.
- Hai account được save.
- Constructor reject missing dependencies.

Unit test không chứng minh rollback thật. PostgreSQL integration test phải inject failure và đọc DB sau rollback.

## 14. Mock, Fake, Stub, Spy Trong Context Này

- **Stub**: trả Account/error cố định để lái branch.
- **Fake**: in-memory repository có behavior gần capability thật.
- **Spy**: recording transactor/saved IDs để quan sát interaction quan trọng.
- **Mock**: pre-program exact expectations; hữu ích khi protocol sequence là requirement, dễ over-specify.

Test smell:

```text
EXPECT Find A
EXPECT Find B
EXPECT Withdraw helper
EXPECT Save A
EXPECT Save B
EXPECT Log
EXPECT Metric
```

Nếu reorder implementation không đổi outcome mà test fail, test đang khóa sequence không phải behavior. Chỉ assert transaction/saves khi đó là contract quan trọng.

## 15. Production Failure Matrix

| Failure | Expected application behavior | Không được giả định |
|---|---|---|
| Sender not found | Trả stable not-found outcome | Không map HTTP trong use case |
| Domain reject | Không save, transaction rollback | Error đồng nghĩa HTTP 409 ở mọi adapter |
| Receiver save fail | Rollback mọi write | No-op/memory fake chứng minh rollback |
| Context canceled trước query | Dừng I/O, propagate cause | Cancellation chắc chắn nghĩa DB chưa commit |
| Commit ambiguous | Reconcile/idempotent retry | Blind retry an toàn |
| Kafka down | Persist outbox intent hoặc policy rõ | DB + Kafka atomic mặc định |

## 16. Debug / Investigation

Khi use case có bug:

1. Viết actor goal và expected outcome.
2. Vẽ runtime sequence, đánh dấu I/O và transaction.
3. Kiểm command có transport type không.
4. Kiểm domain rule có bị duplicate trong use case không.
5. Kiểm mọi repository call dùng đúng transaction context/Unit of Work.
6. Kiểm error trước/sau commit có semantics khác nhau không.
7. Kiểm retry/idempotency key và duplicate result.
8. Viết unit test cho orchestration, integration test cho mechanism.

## 17. Anti-patterns

### God Use Case

Một method 500 dòng làm validation, SQL, mapping, retry, publish và formatting. Tách domain behavior và adapters; không tách thành 20 pass-through services.

### Pass-through Service

```go
func (s *UserService) Get(ctx context.Context, id string) (*User, error) {
	return s.repo.Get(ctx, id)
}
```

Nếu không application policy, service có thể bỏ. Handler gọi query component/repository phù hợp trong tiny CRUD.

### Framework Command

Use case nhận `*http.Request`, `*gin.Context`, protobuf generated message hoặc Kafka message.

### Hidden Globals

Use Case đọc env, global DB/logger/service locator. Dependencies/lifecycle khó thấy và test isolation kém.

### Goroutine Fire-and-forget

Use case spawn goroutine publish event rồi trả success. Context có thể bị cancel, panic/error mất, process crash mất work. Dùng durable outbox/worker khi side effect cần guarantee.

## 18. Khi Nào Không Cần Use Case Struct Riêng?

- CRUD endpoint chỉ gọi một store và map result.
- CLI/script nhỏ tuổi đời ngắn.
- Function thuần không cần dependencies.

Một function có thể là use case:

```go
func CalculateQuote(input QuoteInput) (Quote, error)
```

Không cần suffix `UseCase` hoặc interface nếu package/context đã rõ. Tạo abstraction khi workflow/dependency/test boundary có giá trị.

## 19. Exercises

### Exercise 1: Thêm Frozen Transfer Test

Tạo frozen sender, gọi use case, assert `ErrAccountFrozen`, không account nào được save và transaction được gọi đúng một lần.

### Exercise 2: Authorization

Command có `ActorID`. Thiết kế port/policy kiểm actor sở hữu source account. Quyết định check trước hay trong transaction và race nào tồn tại nếu ownership đổi đồng thời.

### Exercise 3: Idempotency

Thiết kế command/result và `OperationRepository` để duplicate key trả cùng TransferID. Phân tích concurrent duplicate requests.

### Exercise 4: External Fee

Fee rule dùng domain policy, rate config được load ngoài. Tách application acquisition và domain calculation.

## 20. Mastery Questions

1. `TransferMoneyUseCase` khác `Account.Withdraw` về responsibility nào?
2. Tại sao invalid command nên fail trước transaction?
3. Vì sao repository error test không chứng minh rollback?
4. Idempotency nằm ở middleware có đủ không?
5. Khi nào command có thể reuse HTTP DTO?
6. Use Case có nên publish Kafka trực tiếp trong DB transaction?
7. Vì sao default nil transactor thành no-op nguy hiểm?
8. Một pass-through service có nên tồn tại không?
9. Context cancellation sau commit tạo ambiguity gì?
10. Interface nào nên do application sở hữu?

## 21. Further Reading

- [The Clean Architecture - Use Cases](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html): application-specific business rules và dependency direction.
- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/): application core được drive qua incoming adapters và gọi outside qua ports.
- [Go package `context`](https://pkg.go.dev/context): cancellation/deadline contract qua API boundaries.
- [Go Blog: Errors are values](https://go.dev/blog/errors-are-values): error flow idiomatic thay vì exception/service framework.

## 22. Quality Gate

- [x] Problem, mental model ba level và responsibility map.
- [x] Command/result, context, ports và constructor.
- [x] Full Go use case chạy được.
- [x] Domain Service vs Application Service.
- [x] Transaction/idempotency reasoning.
- [x] Wrong examples và anti-patterns.
- [x] Production failure matrix và debug workflow.
- [x] Fake/stub/spy/mock testing strategy.
- [x] Trade-off và khi nào bỏ use case layer.
- [x] Exercises, mastery questions và references.

Thực hành tại [`lab-02-usecase`](../labs/lab-02-usecase/README.md).
