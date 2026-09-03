# Case Study 02: E-commerce Order - Aggregate Và Workflow Phân Tán

Order Service cho thấy hai lớp vấn đề khác nhau: invariant bên trong một Order và workflow qua Inventory, Payment, Shipping. Clean Architecture giúp cô lập policy khỏi transport/provider; nó không biến distributed transaction thành local transaction.

## Bối Cảnh

Khách tạo Order từ cart, áp coupon, reserve inventory, authorize payment rồi confirm. Họ có thể cancel trước khi giao. Giá tại checkout phải được snapshot; Product Service đổi giá sau đó không làm lịch sử Order đổi.

Giả định:

- Inventory và Payment là service độc lập, không dùng chung transaction PostgreSQL.
- Broker có at-least-once delivery.
- Một Order tối đa 100 items.
- Payment provider có thể timeout sau khi đã authorize.
- Product catalog và coupon có owner khác Order.

## Domain Model Và Invariant

`Order` là Aggregate Root; `OrderItem` chỉ thay đổi qua Order. Các invariant quan trọng:

- Draft Order phải có ít nhất một item trước khi submit.
- Quantity dương; `(product_id, unit_price, quantity)` được snapshot.
- `total = sum(line_total) - discount`, không âm.
- Chỉ `Pending` mới được confirm; `Confirmed` không sửa item.
- Cancel phải idempotent về state, nhưng side effect hoàn inventory cần deduplicate riêng.

~~~go
func (o *Order) Confirm(reservationID, authorizationID string) error {
	if o.status != Pending {
		return fmt.Errorf("confirm order: %w", ErrInvalidTransition)
	}
	if reservationID == "" || authorizationID == "" {
		return ErrMissingFulfilmentEvidence
	}
	o.reservationID = reservationID
	o.authorizationID = authorizationID
	o.status = Confirmed
	o.events = append(o.events, OrderConfirmed{OrderID: o.id})
	return nil
}
~~~

Domain không gọi Inventory. Nó kiểm tra evidence mà application thu được. Nếu coupon rule thay đổi mỗi ngày và do Promotion context sở hữu, Order chỉ giữ discount đã được quyết định; một `PromotionGateway` lấy quote trước khi tạo snapshot.

## Boundary

~~~mermaid
flowchart LR
    HTTP["Checkout HTTP"] --> UC["PlaceOrder workflow"]
    UC --> ORDER["Order aggregate"]
    UC --> REPO["OrderRepository"]
    UC --> INV["InventoryGateway"]
    UC --> PAY["PaymentGateway"]
    PG["PostgreSQL"] -.implements.-> REPO
    IHTTP["Inventory HTTP client"] -.implements.-> INV
    PSDK["Provider SDK adapter"] -.implements.-> PAY
    WORKER["Outbox worker"] --> KAFKA["Kafka"]
~~~

Gateway interface thuộc application vì workflow là consumer. Provider DTO, retry header và SDK error dừng ở adapter. `Order` chỉ biết domain values như `ReservationID`.

## Vì Sao Không Giữ Transaction Qua Network

Flow sai:

~~~text
BEGIN -> INSERT order -> call Inventory -> call Payment -> COMMIT
~~~

Network có thể mất giây hoặc không trả lời. Trong thời gian đó connection và lock bị giữ; retry provider có thể tạo side effect, DB rollback không hoàn tác được remote authorization. Atomicity của PostgreSQL không bao trùm HTTP.

Hai lựa chọn hợp lý:

| Cách | Khi phù hợp | Giá phải trả |
|---|---|---|
| Synchronous orchestration + compensation | checkout cần phản hồi nhanh, dependencies ổn định | timeout budget chặt, phải xử lý ambiguous outcome |
| Asynchronous Saga/process manager | workflow dài, chấp nhận pending | state machine, event ordering, UX và operations phức tạp hơn |

## State Machine Của Saga

~~~text
Draft -> PendingInventory -> PendingPayment -> Confirmed
              |                   |
              v                   v
         Rejected             CompensationPending -> Cancelled
~~~

Application ghi Order state và outbox command trong một local transaction. Consumer của reply dùng inbox/idempotency key trước khi chuyển state. Mỗi transition kiểm tra current state để event cũ không kéo Order lùi.

Không dùng một generic `HandleEvent(any)` service. Các command `ReserveInventory`, `AuthorizePayment`, `ConfirmOrder`, `CancelOrder` có input/output và failure semantics riêng.

## Persistence

Các bảng tối thiểu:

- `orders(id, customer_id, status, total_minor, currency, version, ...)`.
- `order_items(order_id, line_no, product_id, unit_price_minor, quantity)`.
- `workflow_steps(order_id, step, status, external_ref, attempt, ...)` nếu cần audit workflow.
- `outbox(id, aggregate_id, event_type, payload, occurred_at, published_at)`.
- `inbox(consumer, message_id, processed_at)`.

Repository load/save toàn Aggregate. Một query analytics như doanh thu theo SKU không nên hydrate hàng nghìn Aggregate; nó thuộc read model/DAO riêng.

## Concurrency

Hai request cancel/confirm có thể đua. Dùng version column hoặc row lock ngắn quanh transition local. Lock không ngăn provider hoàn tất muộn, nên handler của reply vẫn phải kiểm tra state và quyết định ignore, apply hoặc compensation.

Ordering chỉ được tin trong cùng Kafka partition key. Chọn `order_id` làm key để events của cùng Order có thứ tự; vẫn phải chống duplicate.

## Failure Walkthrough

| Điểm hỏng | Trạng thái | Recovery |
|---|---|---|
| Inventory từ chối | `Rejected` | ghi reason, không gọi Payment |
| Inventory timeout không rõ kết quả | `PendingInventory` | query by operation ID trước retry |
| Payment authorize xong nhưng response mất | `PendingPayment` | idempotency key + reconciliation |
| Payment thất bại sau reserve | `CompensationPending` | outbox release inventory |
| DB commit, Kafka down | state bền, outbox pending | worker retry/backoff |
| Confirmed event duplicate | consumer có inbox | skip side effect lần hai |
| Compensation thất bại lâu | workflow không được giả là Cancelled hoàn toàn | alert/manual repair |

## Error Và API

HTTP `POST /orders` có idempotency key. Kết quả có thể là `201 Confirmed`, `202 Pending`, hoặc lỗi business ổn định. `provider timeout` không được map bừa thành "payment failed" nếu kết quả chưa biết. Trả operation/order ID để client poll.

Domain errors không chứa HTTP code. Application phân biệt rejected, conflict, dependency unavailable và unknown outcome. Delivery quyết định status và public message.

## Testing Strategy

- Domain test cho total, currency, duplicate line, transition matrix và event emission.
- Use-case test bằng fake gateway theo script: inventory fail, payment unknown, compensation scheduled.
- Contract test adapter để provider code/JSON map đúng error taxonomy.
- PostgreSQL integration test cho aggregate mapping, optimistic conflict, outbox atomicity.
- Consumer test duplicate/out-of-order event và crash trước/sau inbox commit.
- End-to-end test happy path và compensation với broker/database thật ở CI chọn lọc.
- Chaos drill: tăng payment latency, dừng Kafka, replay message, theo dõi backlog.

## Observability

Trace theo `order_id` và `workflow_id`, nhưng không dùng chúng làm metric label. Metrics: transition count, age của Order theo pending state, compensation backlog, provider outcome unknown, outbox oldest age. Log phải có state cũ/mới và operation ID; tránh dữ liệu thẻ.

## Trade-off Và Phương Án Đơn Giản Hơn

Nếu Inventory, Payment và Order vẫn là module trong cùng modular monolith/database, một local transaction có thể là lựa chọn tốt hơn Saga. Tách microservice chỉ vì sơ đồ đẹp sẽ thêm eventual consistency mà không tạo business value.

Aggregate quá lớn làm mỗi thay đổi item tranh chấp cùng row/version. Aggregate quá nhỏ lại đẩy invariant sang orchestration. Ranh giới được chọn từ invariant và contention thực tế, không từ foreign key.

## Câu Hỏi Mastery

1. Price snapshot thuộc Order hay Product?
2. Vì sao Payment timeout không đồng nghĩa Payment thất bại?
3. Outbox loại bỏ failure window nào? Nó có tạo exactly-once end-to-end không?
4. Event `OrderConfirmed` đến trước `InventoryReserved` thì consumer làm gì?
5. Khi nào nên giữ Order/Inventory trong modular monolith?

## Bài Thực Hành

Vẽ sequence cho cả happy path và payment ambiguous. Viết transition table gồm current state, input event, next state, side effect, hành vi khi duplicate. Sau đó chỉ ra transaction boundary của từng transition.
