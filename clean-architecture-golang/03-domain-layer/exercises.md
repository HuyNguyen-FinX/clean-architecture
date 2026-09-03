# Domain Layer Exercises

Không có lời giải trong file này. Hãy commit hoặc lưu reasoning của bạn trước khi so với implementation mini-banking và các chapter sau.

## Cách Làm

Với mỗi bài:

1. Viết invariant bằng câu/mệnh đề trước code.
2. Phân loại Entity, Value Object, Aggregate, Domain Service, Application Service.
3. Viết public API mong muốn.
4. Liệt kê invalid states và failure paths.
5. Viết tests trước hoặc song song implementation.
6. Ghi trade-off và assumption.

## Bài 1: Hoàn Thiện `Money`

Mở [`money.go`](../examples/mini-banking/internal/account/domain/money.go) nhưng chưa sửa ngay.

Requirement mới:

- Tính fee theo basis points.
- Rounding theo half-up tới minor unit.
- Không cho kết quả overflow.
- Currency phải được giữ nguyên.

Nhiệm vụ:

- Quyết định `BasisPoints` có nên là Value Object.
- Thiết kế signature không dùng `float64`.
- Viết bảng boundary test.
- Giải thích policy rounding thuộc domain nào và vì sao.

Challenge: phân bổ 100 VND cho 3 recipients sao cho tổng vẫn đúng 100 và remainder policy deterministic.

## Bài 2: Account Lifecycle

Thêm status `closed` với rules:

- Active/frozen có thể close khi balance = 0.
- Closed không withdraw/deposit.
- Closed không activate lại.
- Freeze active; freeze frozen là idempotent.

Nhiệm vụ:

- Vẽ state transition diagram.
- Không thêm `SetStatus`.
- Viết error types/categories.
- Test valid/invalid transition và state unchanged.
- Quyết định rehydration historical status unknown xử lý thế nào.

## Bài 3: Overdraft Limit Approval

Overdraft limit không được thay tùy ý:

```text
requested limit <= approved limit from risk policy
new limit không được làm current balance invalid
```

Nhiệm vụ:

- Xác định rule nào thuộc Account, rule nào cần Domain Service/application gateway.
- Thiết kế `ApprovedOverdraft` Value Object nếu hợp lý.
- Không truyền provider SDK response vào domain.
- Viết flow lấy risk result rồi apply limit.

Giải thích vì sao `SetOverdraftLimit(amount)` là API yếu.

## Bài 4: Order Aggregate

Requirement:

- Order pending có thể thêm/xóa item.
- Mỗi product chỉ xuất hiện một dòng; add lần hai tăng quantity.
- Quantity 1-99.
- Tổng Order bằng tổng item.
- Confirmed Order không sửa item.

Nhiệm vụ:

- Chọn Entity/Value Object cho OrderItem.
- Chỉ expose mutation qua Order root.
- Trả Items mà caller không mutate backing data.
- Viết repository port theo aggregate.
- Viết tests cho invariant và aliasing.

Challenge: coupon phụ thuộc customer tier và campaign service. Tách pure domain rule khỏi external data acquisition.

## Bài 5: Aggregate Boundary Trong Banking

Bạn có:

```text
Customer
Account
Card
Transfer
```

Requirement:

- Freeze Customer phải chặn card payment trong tối đa 5 giây.
- Account transfer cần strong balance consistency.
- Card name update không được conflict với deposit.
- Transfer history cần query theo Customer.

Nhiệm vụ:

- Đề xuất aggregate boundaries.
- Mô tả references bằng IDs.
- Chọn strong/eventual consistency cho từng rule.
- Phân tích hot Customer/account contention.
- Thiết kế read model history mà không load toàn aggregate graph.

Không có đáp án duy nhất. Rubric là reasoning về invariant, lifecycle, transaction và load/lock size.

## Bài 6: Domain Event Và Outbox

Thêm `AccountFrozen` event.

Nhiệm vụ:

- Event chỉ sinh khi active chuyển frozen.
- Gọi freeze lần hai không tạo duplicate theo policy.
- Không import Kafka trong domain.
- Thiết kế mapper sang `AccountFrozenV1`.
- Vẽ transaction/outbox/publisher flow.
- Liệt kê failure nếu worker publish rồi crash trước mark.

Viết test domain riêng với test outbox integration trong design, không dùng một test mock tất cả.

## Bài 7: Review Invalid State

Review code:

```go
type Subscription struct {
	ID        string
	Plan      string
	StartsAt  time.Time
	EndsAt    *time.Time
	Cancelled bool
}

func (s *Subscription) SetPlan(plan string)       { s.Plan = plan }
func (s *Subscription) SetEndsAt(at *time.Time)   { s.EndsAt = at }
func (s *Subscription) SetCancelled(value bool)   { s.Cancelled = value }
```

Requirement:

- End phải sau start.
- Cancelled subscription phải có cancellation time/reason.
- Enterprise plan không downgrade khi invoice pending.

Nhiệm vụ:

- Liệt kê invalid states.
- Đề xuất Entity API theo behavior.
- Xác định invoice pending data lấy ở đâu.
- Chỉ ra rule nào không thể bảo vệ chỉ bằng Subscription instance.

## Bài 8: Domain Debugging

Incident:

```text
Một account persisted với balance -90.000, overdraft limit 50.000.
Không log nào ghi withdrawal vượt limit.
```

Lập investigation plan:

- Creation/rehydration paths.
- Direct mutation/ORM hooks/manual SQL.
- Concurrent requests và isolation.
- Integer overflow.
- Migration/legacy data.
- Regression tests theo từng hypothesis.

Không kết luận “domain bug” trước khi phân biệt object invariant với persisted/concurrent system invariant.

## Self-review Rubric

- [ ] Model dùng language của requirement, không language của table/framework.
- [ ] Identity và equality được định nghĩa rõ.
- [ ] Mỗi invariant có creation và transition guards.
- [ ] Rejected operation giữ state unchanged.
- [ ] Aggregate Root là mutation entry point.
- [ ] Cross-aggregate consistency strategy explicit.
- [ ] Domain Service không làm I/O.
- [ ] Domain Event tách khỏi integration schema.
- [ ] Test không mock domain objects.
- [ ] Có phần “khi nào thiết kế này quá nặng”.
