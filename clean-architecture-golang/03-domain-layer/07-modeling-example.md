# Modeling Walkthrough: Mini-banking

Phần này đi từ requirement đến model. Mục tiêu không phải chứng minh model hiện tại là thiết kế duy nhất; mục tiêu là làm rõ từng quyết định và assumption.

## 1. Requirement Ban Đầu

```text
Một account có balance, currency, overdraft limit và status.
Account active có thể deposit/withdraw.
Account frozen không được withdraw nhưng vẫn nhận deposit.
Balance sau withdrawal không thấp hơn âm overdraft limit.
Transfer chuyển amount dương từ account A sang account B cùng currency.
```

Các câu còn mơ hồ cần hỏi domain expert:

- Overdraft limit có đổi được không? Ai duyệt?
- Frozen có chặn incoming transfer không?
- Currency của account có đổi được không?
- Fee được trừ thêm hay nằm trong amount?
- Transfer giữa currency khác nhau có được phép qua FX không?
- Balance là source of truth hay được derive từ ledger?
- Account closed có nhận reversal không?

Model không thể chính xác hơn requirement. Đừng che câu hỏi product bằng abstraction code.

## 2. Tìm Identity, Values Và Transitions

Candidate classification:

| Concept | Classification | Reasoning |
|---|---|---|
| Account | Entity/Aggregate Root | Có identity/lifecycle, state đổi qua thời gian |
| AccountID | Identity Value | Type-safe identity representation |
| Money | Value Object | Equality theo amount + currency, arithmetic rules |
| Currency | Value Object | Normalize/validate code |
| AccountStatus | Value | State set hữu hạn |
| Transfer | Candidate Entity/Aggregate | Có operation identity/status/lifecycle khi thêm history/idempotency |

Đừng tự động biến mọi noun thành struct. `Withdrawal` có thể chỉ là behavior/event, hoặc Entity nếu business cần lifecycle/identity riêng.

## 3. Viết Invariant Trước Code

```text
I1: AccountID != empty
I2: Money currency có shape hợp lệ
I3: balance.currency == overdraftLimit.currency
I4: overdraftLimit.amount >= 0
I5: balance.amount >= -overdraftLimit.amount
I6: amount movement > 0
I7: frozen account cannot withdraw
```

Phân loại:

- I1-I5 phải đúng khi Account được tạo/rehydrate.
- I3, I5-I7 phải đúng qua behavior.
- Transfer thêm rule from != to ở application command vì nó liên quan hai identities, không state nội bộ một Account.

## 4. Thiết Kế Value Object Trước Primitive API

```go
type Money struct {
	amount   int64
	currency Currency
}
```

API tối thiểu từ consumer hiện tại:

```go
func NewMoney(amount int64, currency string) (Money, error)
func NewPositiveMoney(amount int64, currency string) (Money, error)
func (m Money) Add(other Money) (Money, error)
func (m Money) Sub(other Money) (Money, error)
func (m Money) LessThan(other Money) (bool, error)
func (m Money) Equal(other Money) bool
```

Không thêm multiply/divide/allocate trước khi fee/tax use case cần và rounding policy rõ.

## 5. Thiết Kế Aggregate API Theo Intent

```go
type Account struct {
	id             AccountID
	balance        Money
	overdraftLimit Money
	status         AccountStatus
}
```

Read API:

```go
func (a *Account) ID() AccountID
func (a *Account) Balance() Money
func (a *Account) OverdraftLimit() Money
func (a *Account) Status() AccountStatus
```

Command API:

```go
func (a *Account) Deposit(amount Money) error
func (a *Account) Withdraw(amount Money) error
func (a *Account) Freeze()
func (a *Account) Activate()
```

Không expose `SetBalance`, `SetCurrency` hoặc pointer tới fields.

## 6. Code `Withdraw` Theo Candidate State

```go
func (a *Account) Withdraw(amount Money) error {
	if a.status == AccountStatusFrozen {
		return ErrAccountFrozen
	}
	if !amount.IsPositive() {
		return ErrInvalidAmount
	}

	next, err := a.balance.Sub(amount)
	if err != nil {
		return err
	}

	minimum, err := a.overdraftLimit.Negate()
	if err != nil {
		return err
	}
	tooLow, err := next.LessThan(minimum)
	if err != nil {
		return err
	}
	if tooLow {
		return ErrInsufficientBalance
	}

	a.balance = next
	return nil
}
```

Thứ tự reasoning:

1. Check transition permission.
2. Check operation input semantics.
3. Tính candidate state bằng Value Object operation.
4. Check invariant trên candidate.
5. Assign một lần cuối.

Nếu step 1-4 lỗi, object không đổi.

## 7. Constructor Và Rehydration

Application/test mở account qua:

```go
NewAccount(id, balance, overdraftLimit)
```

Database adapter sau này có status persisted nên gọi:

```go
RehydrateAccount(id, balance, overdraftLimit, status)
```

Cả hai đi qua cùng invariant validation. Production `OpenAccount` có thể tách riêng nếu creation policy buộc zero balance hoặc cần domain event.

Một anti-pattern phổ biến là dùng unexported “unsafe constructor” cho repository. Nó nhanh nhưng biến database thành đường bypass invariant. Nếu historical data có state cũ không hợp lệ theo rule mới, cần migration/versioned model policy rõ, không im lặng nạp object sai.

## 8. Domain Error Là Outcome, Không Là Response

```go
var ErrInsufficientBalance = errors.New("insufficient balance")
```

Domain error giúp caller phân nhánh bằng `errors.Is`, không parse string. Nó không chứa 409.

Khi cần data, dùng typed error:

```go
type InsufficientBalanceError struct {
	Available Money
	Requested Money
}

func (e *InsufficientBalanceError) Error() string {
	return "insufficient balance"
}
```

Chỉ expose available/requested nếu policy bảo mật cho phép. Domain detail có thể bị transport log/response làm lộ thông tin nếu mapping cẩu thả.

Sentinel phù hợp khi caller chỉ cần category. Typed error phù hợp khi structured detail có ý nghĩa. Chapter Error Handling sẽ đào sâu wrapping/taxonomy.

## 9. Application Orchestration Hai Aggregates

```go
return uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
	sender, err := uc.accounts.FindByID(txCtx, fromID)
	// ...
	receiver, err := uc.accounts.FindByID(txCtx, toID)
	// ...

	if err := sender.Withdraw(amount); err != nil {
		return err
	}
	if err := receiver.Deposit(amount); err != nil {
		return err
	}

	if err := uc.accounts.Save(txCtx, sender); err != nil {
		return err
	}
	return uc.accounts.Save(txCtx, receiver)
})
```

Domain không có `Transfer(sender, receiver)` global service vì current workflow còn transaction/repository concern. Có thể tách pure Domain Service nếu transfer eligibility/fee cần rule xuyên accounts, nhưng loading/saving vẫn ở application.

`NoopTransactor` của memory version không bảo đảm atomicity. Nếu save receiver lỗi sau sender save, state partial. README phải nói rõ guarantee hiện tại thay vì dùng interface name để giả vờ production-ready.

## 10. Test Từ Invariant Matrix

Đừng viết test theo method count. Lập bảng behavior:

| State | Operation | Input | Expected |
|---|---|---|---|
| active | withdraw | positive, within limit | balance giảm |
| active | withdraw | exact limit | allow |
| active | withdraw | beyond limit | error, state unchanged |
| frozen | withdraw | positive | frozen error, unchanged |
| frozen | deposit | positive same currency | allow theo policy |
| any | deposit/withdraw | zero/negative | invalid amount |
| any | movement | other currency | currency mismatch |

Money matrix thêm normalization, equality, overflow/underflow và zero value.

## 11. Dependency Review

Domain source imports:

```text
errors
fmt
math
strings
```

Không có:

```text
context
net/http
database/sql
pgx
kafka
redis
otel
```

Không phải vì standard library luôn “inner-safe”. `net/http` vẫn là transport detail. Architecture fitness test chặn third-party/project imports trong domain của example; code review vẫn cần bắt semantic leaks.

## 12. Production Evolution

Model hiện tại là balance-based account. Khi yêu cầu tăng:

### V4: Persistence

Thêm row mapper và `RehydrateAccount`, database constraints mirror critical invariant.

### V5: Transaction/Concurrency

Lock/version Account; save sender/receiver/Transfer atomic; test concurrent withdrawal.

### V8: Idempotency

`Transfer` có ID/key/status. Application trả kết quả cũ cho duplicate key, database unique constraint enforce.

### V9-V10: Event/Outbox

Domain/Application tạo `MoneyTransferred`; adapter map sang integration event; outbox đảm bảo publish intent không mất.

### Ledger Evolution

Nếu audit/accounting cần immutable ledger, balance có thể trở thành derived/cached value từ entries. Đây là thay đổi domain model lớn, không chỉ đổi SQL adapter. Các invariant và consistency boundary phải được modeling lại.

## 13. Điều Còn Cố Ý Chưa Làm

- Không hỗ trợ FX/rounding/fee.
- Không có Transfer aggregate/history.
- Không có Clock/domain event trong Account.
- Không có closed/pending status graph.
- Không có PostgreSQL lock/version.
- Không có transaction thật trong memory adapter.
- Không có idempotency/outbox.

Giới hạn explicit giúp người học phân biệt code minh họa một concept với production guarantee.

## 14. Câu Hỏi Thiết Kế

1. Nếu frozen account không được deposit, method nào/test nào đổi?
2. Nếu overdraft limit thay đổi cần approval, `SetOverdraftLimit` có đủ không?
3. Transfer fee thuộc Account, Transfer hay Domain Service?
4. Nếu account multi-currency, `Account` aggregate hiện tại còn phù hợp không?
5. Nếu balance derive từ ledger, repository API và concurrency strategy đổi ra sao?
6. Khi nào `Transfer` trở thành Entity thay vì chỉ application operation?
7. Rule from != to thuộc Account hay application? Vì sao?
