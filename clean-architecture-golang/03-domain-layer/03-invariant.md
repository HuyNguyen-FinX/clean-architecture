# Invariant Và Valid State

## Invariant Là Gì?

Invariant là mệnh đề phải đúng tại mọi thời điểm observable mà domain coi object/aggregate là hợp lệ.

Mini-banking:

```text
balance >= -overdraftLimit
```

Rule không chỉ được check ở HTTP request. Nó phải đúng sau constructor, rehydration và mọi state transition.

## Validation Khác Invariant

| Loại | Ví dụ | Owner thường gặp |
|---|---|---|
| Transport validation | JSON parse được, field `amount` tồn tại | HTTP adapter |
| Application validation | from/to không cùng account, actor được phép gọi use case | Application |
| Domain invariant | Không withdraw vượt overdraft, frozen không withdraw | Domain |
| Persistence constraint | Unique idempotency key, numeric/check constraint | Database adapter/schema |

Một rule có thể được enforce ở nhiều nơi để fail sớm và defense in depth, nhưng phải có một nguồn semantics chính.

Ví dụ handler check `amount > 0` giúp response nhanh/đẹp. Domain vẫn phải reject non-positive Money movement vì Kafka consumer hoặc internal caller có thể bypass HTTP.

## Ba Level

### Level 1: Trực Giác

Invariant định nghĩa “hàng rào” quanh state hợp lệ. Method được phép thay đổi state nhưng không được trả control cho caller khi object đã vượt hàng rào.

### Level 2: Backend Engineer

Guard rails trong Go:

- Private fields.
- Constructors và rehydration factories trả error.
- Behavior methods thay setter.
- Validate trước khi assign.
- Rejected operation không để partial mutation.
- Tests cho boundary values.

### Level 3: Architecture / Consistency

Invariant dẫn đến transaction/concurrency question:

- Invariant nằm trong một object instance hay cần nhiều aggregate?
- State được đọc có còn current khi commit không?
- Database constraint nào là last line of defense?
- Eventual consistency có cho phép temporary violation không?

Domain code bảo vệ in-memory transition; transaction/lock/version bảo vệ state trước concurrent writers.

## Public Field Phá Hàng Rào

```go
type Account struct {
	Balance        int64
	OverdraftLimit int64
}

account.Balance = -999_999
```

Không compiler/test nào buộc caller gọi validation. Mỗi code path có thể tạo state bất hợp lệ.

Private fields:

```go
type Account struct {
	balance        Money
	overdraftLimit Money
}
```

Caller chỉ có getters trả value và behavior methods. Search mutation được thu hẹp vào package domain.

Private field không bảo vệ khỏi:

- Constructor sai.
- Method mutate trước rồi mới validate nhưng không rollback.
- Pointer/slice/map internal bị leak.
- Database row corrupt đi qua unsafe rehydration.
- Concurrent stale snapshot.

Encapsulation là cơ chế, invariant reasoning mới là mục tiêu.

## Constructor Cũng Là State Transition

Constructor cũ của mini-banking từng có lỗ hổng:

```go
func NewAccount(id AccountID, balance, limit Money) (*Account, error) {
	if limit.IsNegative() {
		return nil, ErrInvalidOverdraftRule
	}
	return &Account{balance: balance, overdraftLimit: limit}, nil
}
```

Input sau pass:

```text
balance = -100.000
limit   =   50.000
```

`Withdraw` có thể hoàn hảo nhưng aggregate sinh ra đã sai. Correct constructor tính minimum balance và reject:

```go
minimum, err := overdraftLimit.Negate()
if err != nil {
	return nil, err
}
tooLow, err := balance.LessThan(minimum)
if err != nil {
	return nil, err
}
if tooLow {
	return nil, ErrInvalidOverdraftRule
}
```

Regression test nằm trong `account_test.go`.

## Validate Trước Khi Mutate

Wrong:

```go
func (a *Account) Withdraw(amount Money) error {
	a.balance, _ = a.balance.Sub(amount)
	if tooLow(a.balance, a.overdraftLimit) {
		return ErrInsufficientBalance
	}
	return nil
}
```

Caller nhận error nhưng object đã bị mutate. Nếu caller log/reuse/save object, invalid state thoát ra.

Correct dùng candidate state:

```go
next, err := a.balance.Sub(amount)
if err != nil {
	return err
}
tooLow, err := next.LessThan(minimumBalance)
if err != nil {
	return err
}
if tooLow {
	return ErrInsufficientBalance
}

a.balance = next
return nil
```

Assignment là bước cuối sau mọi guard.

## Boundary Values

Rule:

```text
balance >= -limit
```

Với balance 100.000 và limit 50.000:

| Withdraw | Next balance | Kết quả |
|---:|---:|---|
| 149.999 | -49.999 | allow |
| 150.000 | -50.000 | allow |
| 150.001 | -50.001 | reject |

Test chỉ happy path và “rất lớn” có thể bỏ sót dấu `>` vs `>=`. Boundary-focused tests là specification chính xác hơn một paragraph.

## State Transition Invariant

Account status:

```text
active --Freeze--> frozen
frozen --Activate--> active
```

Policy hiện tại:

- Frozen account không withdraw.
- Frozen account vẫn deposit.
- Freeze/Activate idempotent.

Nếu domain thêm `closed`, transition graph cần rõ:

```text
active -> frozen -> active
active -> closed
frozen -> closed
closed -> không transition
```

Một `SetStatus(string)` cho phép `closed -> active` hoặc typo `frozon`. Behavior methods/factory giới hạn transition hợp lệ.

## Invariant Xuyên Nhiều Aggregate

“Tổng tiền debit bằng tổng tiền credit” trong transfer liên quan hai Account và Transfer record. Gom tất cả thành một aggregate khổng lồ có thể bảo vệ local invariant nhưng làm mọi transfer lock/chạm nhiều state.

Thường thiết kế:

- `Account` tự bảo vệ balance.
- Application transaction orchestration hai Account.
- Database locking/version bảo vệ concurrent write.
- `Transfer` record/idempotency bảo vệ operation identity.

Đây là application/system invariant được enforce bằng nhiều boundary. Không phải mọi business rule đều phải nằm trong một Entity method.

## Database Constraint Có Thay Domain Invariant Không?

PostgreSQL có thể có:

```sql
CHECK (overdraft_limit_minor >= 0)
CHECK (balance_minor >= -overdraft_limit_minor)
```

Constraint là defense in depth tốt:

- Chặn manual SQL/bad adapter.
- Bảo vệ data at rest.
- Enforce atomic với write.

Nhưng nếu chỉ dựa DB:

- Domain behavior trả raw constraint failure muộn.
- In-memory/test/other storage không có rule.
- Error mapping phụ thuộc constraint name.
- Business intent không hiện trong model.

Với invariant quan trọng, domain và DB có thể cùng enforce. Duplicate mechanism không nhất thiết duplicate source of truth nếu semantics được định nghĩa ở domain và schema là guard cuối.

## Failure Scenario: Stale Read

Hai request load balance 1.000.000, mỗi request rút 800.000. Mỗi aggregate instance giữ invariant, nhưng cả hai dựa trên cùng version cũ.

Pessimistic lock:

```sql
SELECT id, balance_minor, overdraft_limit_minor
FROM accounts
WHERE id = $1
FOR UPDATE;
```

Optimistic write:

```sql
UPDATE accounts
SET balance_minor = $1, version = version + 1
WHERE id = $2 AND version = $3;
```

Nếu affected rows = 0, version conflict. Application quyết định retry hay trả conflict dựa trên idempotency/latency policy.

Domain invariant không “thất bại”; scope của nó là object snapshot. Architecture phải nối scope đó với storage concurrency.

## Debug Checklist

Khi thấy invalid state:

1. Viết invariant bằng biểu thức cụ thể.
2. Tìm mọi constructor/rehydration/deserialization path.
3. Tìm mọi direct mutation trong package.
4. Kiểm rejected operation có mutate trước error không.
5. Kiểm aliasing của slice/map/pointer.
6. Kiểm DB constraints/migration/manual writes.
7. Kiểm transaction isolation, lock/version và retries.
8. Thêm test ở layer thấp nhất tái hiện được bug.

## Khi Nào Không Cần Encapsulation Nặng?

Read DTO hoặc report row không có state transition có thể dùng exported fields. Một tiny CRUD model có thể chấp nhận tags/public fields. Không thêm getter/setter vô nghĩa chỉ để giống object-oriented template.

Đầu tư private state khi có invariant đáng bảo vệ hoặc nhiều caller có nguy cơ bypass rule.

## Mastery Questions

1. Validation `amount required` khác invariant `balance >= -limit` thế nào?
2. Vì sao private field cần nhưng chưa đủ?
3. Tại sao constructor phải kiểm invariant giống transition methods?
4. Vì sao mutate rồi trả error là bug nguy hiểm?
5. Domain test pass nhưng concurrent withdrawal vẫn double spend vì sao?
6. Khi nào nên enforce cùng invariant trong domain và database?
7. Invariant xuyên hai Account có buộc hai Account thành một Aggregate không?
