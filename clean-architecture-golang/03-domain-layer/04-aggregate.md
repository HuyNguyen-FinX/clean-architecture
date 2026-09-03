# Aggregate Và Consistency Boundary

## Aggregate Không Phải “Một Nhóm Struct Có Liên Quan”

Order có OrderItems không tự động đủ để kết luận đây là một Aggregate. Aggregate là cluster domain objects được xem như một unit để bảo vệ invariant. Một object là Aggregate Root; code ngoài đi qua root để thay đổi members.

```text
Order (Aggregate Root)
├── OrderItem
├── OrderItem
└── ShippingAddress
```

Lý do gom không phải vì chúng vẽ gần nhau. Lý do là rule cần consistency:

```text
order total = tổng item totals
confirmed order không được sửa item
shipping address phải valid trước confirm
```

## Ba Level

### Level 1: Trực Giác

Aggregate là phạm vi mà một “người gác cửa” có thể nói: sau operation này, mọi rule bên trong vẫn đúng.

Code ngoài không lấy `OrderItem` rồi tự đổi quantity; nó gọi `order.ChangeQuantity(productID, quantity)`.

### Level 2: Backend Engineer

Aggregate ảnh hưởng:

- Repository thường load/save theo Aggregate Root.
- Transaction thường giữ một Aggregate nhất quán.
- Internal entities/value objects không bị mutation trực tiếp từ ngoài.
- Cross-aggregate reference thường dùng ID.
- Concurrency token/version thường gắn với root.

### Level 3: Architecture / Domain Modeling

Aggregate là trade-off giữa consistency và concurrency:

```text
Aggregate lớn
  + enforce nhiều invariant trong một transaction
  - load nhiều data, lock rộng, contention cao

Aggregate nhỏ
  + update độc lập, scale/concurrency tốt hơn
  - cross-aggregate rule cần application workflow/eventual consistency
```

## Aggregate Root

Root sở hữu API thay đổi toàn aggregate:

```go
type Order struct {
	id     OrderID
	status OrderStatus
	items  []OrderItem
}

func (o *Order) AddItem(product ProductID, price Money, quantity int) error
func (o *Order) Confirm() error
```

Getter trả mutable slice nội bộ sẽ phá boundary:

```go
func (o *Order) Items() []OrderItem {
	return o.items // caller có thể sửa backing array
}
```

Trả copy hoặc read-only view:

```go
func (o *Order) Items() []OrderItem {
	return append([]OrderItem(nil), o.items...)
}
```

Nếu `OrderItem` chứa pointer/map/slice, shallow copy chưa đủ; ownership phải được thiết kế rõ.

## Account Là Aggregate Gì?

Trong mini-banking hiện tại:

```text
Account (Aggregate Root)
├── Money balance
├── Money overdraftLimit
└── AccountStatus
```

`Money` và `AccountStatus` là values, không phải child entities. Aggregate chỉ có một Entity `Account`, nhưng vẫn là Aggregate. Aggregate không bắt buộc là object graph lớn.

Invariant mà root bảo vệ:

- Currency nhất quán.
- Balance không thấp hơn overdraft limit.
- Frozen status chặn withdrawal.

## Có Nên Gom Account A, Account B Và Transfer?

Transfer money liên quan:

```text
Account A
Account B
Transfer
```

Gom cả ba thành một Aggregate nghe có vẻ giúp atomicity, nhưng identity/lifecycle không tự nhiên:

- Account A tồn tại trước/sau transfer và tham gia nhiều transfer.
- Account B cũng vậy.
- Mỗi transfer khác lại muốn “sở hữu” cùng account.
- Load aggregate transfer sẽ phải load hai account và có thể history.
- Hot account làm mọi aggregate chồng lấn, không còn ownership đơn nhất.

Thiết kế thường hợp lý hơn:

```text
Account A aggregate
Account B aggregate
Transfer aggregate/record

Application use case orchestration local DB transaction
```

Cross-aggregate transaction không bị DDD “cấm” về mặt vật lý. Guideline một transaction một aggregate giúp giữ boundary nhỏ, nhưng banking transfer có consistency requirement mạnh có thể cập nhật nhiều aggregate trong một local transaction. Hãy ghi rõ exception và contention/deadlock strategy.

## Consistency Boundary Không Luôn Trùng Storage Boundary

Một aggregate có thể map nhiều tables:

```text
orders
order_items
order_addresses
```

Repository load/save như một unit dù relational storage tách rows.

Ngược lại, một table có thể chứa nhiều Aggregate instances. Table không định nghĩa aggregate. ORM association cũng không định nghĩa consistency boundary.

## Repository Theo Aggregate

Repository thường trả root:

```go
type OrderRepository interface {
	FindByID(ctx context.Context, id OrderID) (*Order, error)
	Save(ctx context.Context, order *Order) error
}
```

Tránh repository cho internal item chỉ để CRUD:

```go
type OrderItemRepository interface {
	UpdateQuantity(ctx context.Context, itemID string, quantity int) error
}
```

Nếu code ngoài update item không qua Order, invariant `confirmed order không đổi item` bị bypass.

Ngoại lệ: bulk/reporting/query path có thể đọc table riêng mà không rehydrate aggregate, miễn nó không bypass command-side rule để mutate.

## Cross-Aggregate Reference

Giữ ID thường rõ ownership hơn pointer:

```go
type Transfer struct {
	fromAccountID AccountID
	toAccountID   AccountID
}
```

Thay vì:

```go
type Transfer struct {
	from *Account
	to   *Account
}
```

Pointer graph làm object nào sở hữu mutation mơ hồ, dễ load graph lớn và serialize vòng. Application load aggregates cần thiết qua repositories.

Không phải mọi in-memory relation đều cấm pointer. Trong cùng aggregate, pointer/value có thể hợp lý. Guideline ID nhắm vào reference **xuyên aggregate**.

## Transaction Boundary

Một aggregate operation trong memory là atomic ở mức method nếu không có concurrent mutation. Persistence cần transaction/version để tránh partial write.

Order save nhiều table:

```text
BEGIN
UPDATE orders
DELETE/UPSERT order_items
UPDATE address
COMMIT
```

Transfer hai Account:

```text
BEGIN
lock A và B theo thứ tự ổn định
load A và B
A.Withdraw
B.Deposit
save A và B
insert transfer
COMMIT
```

Aggregate reasoning cho biết rule; Transaction chapter sẽ chọn mechanism và xử lý commit failure/retry.

## Concurrency Và Aggregate Size

Giả sử `Customer` aggregate chứa 100 Accounts và mọi profile/account update cùng version:

- Đổi email conflict với deposit dù hai action không chung invariant.
- Load customer kéo 100 accounts.
- Hot customer tạo lock contention.

Đây là tín hiệu aggregate quá lớn. Nếu Customer status phải chặn toàn bộ account withdrawal tức thời, cần phân tích:

- Strong consistency thật sự cần đến mức nào?
- Copy status vào Account có chấp nhận propagation delay không?
- Application lock Customer và Account trong cùng transaction có deadlock risk gì?
- Có một authorization/policy service trước withdrawal không?

Không giải bằng câu “DDD nói aggregate nhỏ”. Business consistency requirement quyết định, performance/concurrency tạo constraint.

## Aggregate Quá Nhỏ

Nếu `OrderItem` mỗi cái là aggregate riêng:

- `Order` không kiểm soát total.
- Confirm có thể xảy ra trong lúc item update.
- Rule max 100 items cần distributed query/lock.
- Repository API biến thành CRUD rows.

Aggregate nhỏ chỉ tốt khi objects thực sự có lifecycle/consistency độc lập.

## Domain Event Trong Aggregate

Root có thể record fact sau successful transition:

```go
func (o *Order) Confirm(at time.Time) error {
	if o.status != StatusPending {
		return ErrInvalidOrderTransition
	}
	o.status = StatusConfirmed
	o.events = append(o.events, OrderConfirmed{
		OrderID:    o.id,
		OccurredAt: at,
	})
	return nil
}
```

Event không được record trước validation. Pull/clear event lifecycle phải phối hợp transaction để không mất hoặc publish duplicate. Domain event không đồng nghĩa Kafka message; xem bài 06.

## Production Scenario: Hot Account

Một merchant account nhận hàng nghìn transfer/giây. Nếu mỗi transfer lock account row để update balance:

- Account aggregate là consistency boundary hợp lý về rule.
- Nhưng account trở thành serialization point vật lý.
- Tăng application instances không loại bỏ row contention.

Các hướng thiết kế khác có thể gồm ledger append-only, partitioned balance calculation hoặc reservation model. Chúng thay domain model và consistency semantics, không chỉ đổi repository implementation.

Aggregate design phải sống cùng throughput requirement. Clean package boundary không chữa hot key.

## Failure Scenarios

### Save Một Phần Aggregate

Order root update thành confirmed nhưng item rows fail. Nếu repository không transaction, persisted aggregate vi phạm invariant dù object in-memory đúng.

### Concurrent Writers

Hai copies cùng version mutate hợp lệ; last-write-wins làm mất một transition. Thêm version check hoặc lock.

### Event Publish Trước Commit

Consumer thấy `OrderConfirmed` nhưng DB rollback. Record event/outbox trong cùng transaction, publish sau.

### Aggregate Load Thiếu Data

Repository lazy-load items sau khi transaction đóng; `Confirm` tính rule trên partial collection. Aggregate phải được load đủ state cần cho invariant, hoặc model/query phải thể hiện explicit incompleteness.

## Debug Questions

1. Invariant nào đang bị vi phạm?
2. Objects nào cần đọc cùng lúc để quyết định?
3. Ai được phép mutate child?
4. Repository load/save unit là gì?
5. Transaction/version bảo vệ unit nào?
6. Có path SQL nào update child trực tiếp không?
7. Aggregate có load/lock quá nhiều state không liên quan?
8. Cross-aggregate rule cần strong hay eventual consistency?

## Khi Nào Không Cần Aggregate Pattern Rõ Nét?

- CRUD/reference data không có invariant xuyên objects.
- Read model/reporting.
- Data pipeline transform immutable records.
- Prototype chưa biết consistency requirement.

Đừng tạo `AggregateRoot` interface/base struct chung chỉ để đánh dấu mọi Entity. Trong Go, behavior và repository boundary quan trọng hơn marker type.

## Mastery Questions

1. Vì sao “Order có nhiều OrderItem” chưa đủ chứng minh Aggregate?
2. Aggregate Root bảo vệ internal members bằng API nào?
3. Tại sao table boundary không phải aggregate boundary?
4. Có nên gom hai Account và Transfer thành một Aggregate? Phân tích identity/lifecycle/contention.
5. “Một transaction không được chạm hai aggregate” có phải luật tuyệt đối không?
6. Aggregate quá lớn ảnh hưởng lock/conflict thế nào?
7. Query read model có bắt buộc load Aggregate không?
8. Domain event được record trong root nhưng publish reliability thuộc boundary nào?
