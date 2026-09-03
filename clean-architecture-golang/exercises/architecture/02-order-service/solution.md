# Solution tham khảo: Order Service

## Model

Order là Aggregate Root; OrderItem chỉ mutate qua Order. Invariants: quantity dương, total=sum lines-discount, transition Pending→Confirmed/Cancelled hợp lệ.

Coupon có hai dạng:

- rule nội bộ ổn định: PromotionPolicy Domain Service;
- external Campaign context: Gateway/ACL trả DiscountDecision đã normalize.

## Ports

~~~go
type OrderRepository interface {
	FindByID(context.Context, OrderID) (*Order, error)
	Save(context.Context, *Order) error
}

type InventoryGateway interface {
	Reserve(context.Context, ReservationRequest) (ReservationResult, error)
	Release(context.Context, ReservationID) error
}
~~~

## Workflow

Không giữ DB transaction trong network:

~~~text
CreateOrder
→ transaction: save Pending + outbox OrderSubmitted
→ process manager: reserve inventory
→ authorize payment
→ transaction: Confirm + outbox OrderConfirmed
~~~

Failure reserve → Rejected/Cancelled. Payment success rồi inventory failure cần void/refund compensation. Compensation có thể fail, state ManualReview.

## Idempotency

Create key + cart/request hash. Inventory reservation có stable operation ID. Event handlers inbox by event ID.

## Events

Domain: OrderConfirmed. Integration v1 mapper chỉ publish fields consumers need; PII/address redacted. Outbox local transaction.

## Consistency

Order total/status strong within Aggregate. Inventory/payment cross-context eventual/pending. UI trả 202 + order status nếu workflow async.

## Tests

- total/transition property/table tests;
- process manager transition/failure;
- Repository transaction/outbox Postgres;
- provider/inventory contract;
- duplicate/out-of-order events;
- E2E eventual status.

## Alternative

Monolith cùng Postgres có thể transaction Order+Inventory nếu cùng bounded context/team và scale. Đừng introduce Saga chỉ vì future microservices.

## Operations

Metrics pending age, reservation timeout, compensation failures, outbox age. Reconciliation Order vs Payment/Inventory.
