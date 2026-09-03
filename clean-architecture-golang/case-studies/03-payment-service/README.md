# Case Study 03: Payment Service - Idempotency Và Kết Quả Không Chắc Chắn

Payment là domain mà "retry khi lỗi" có thể thu tiền hai lần. Trọng tâm không phải bọc SDK bằng interface cho đẹp, mà là mô hình hóa lifecycle, danh tính operation và trạng thái `unknown` một cách trung thực.

## Scope Và Trust Boundary

Service hỗ trợ authorize, capture, void và refund; nhận webhook từ provider. Nó không lưu PAN/CVV. Tokenization, PCI scope, signature verification và secret rotation là yêu cầu kiến trúc, không chỉ security middleware.

Assumption:

- Provider hỗ trợ idempotency key nhưng retention có giới hạn.
- Webhook at-least-once, có thể đến trước response đồng bộ.
- Currency amount dùng minor unit integer.
- Một payment có thể capture/refund nhiều phần nếu business cho phép.

## Domain Model

Không dùng một boolean `paid`. State machine tối thiểu:

~~~text
Created -> AuthorizationPending -> Authorized -> CapturePending -> Captured
                |                     |              |              |
                v                     v              v              v
             Failed                Voided         Unknown      RefundPending
                                                                    |
                                                                    v
                                                                 Refunded
~~~

`Unknown` không nhất thiết là terminal state; nó biểu diễn local system chưa biết provider đã tạo side effect chưa. Reconciliation sẽ chuyển nó sang kết quả xác định.

Invariant ví dụ:

- Amount dương và currency được hỗ trợ.
- Tổng capture không vượt authorized amount.
- Tổng refund không vượt captured amount.
- Không void sau capture.
- Provider reference đã gắn không được đổi tùy ý.

~~~go
func (p *Payment) RecordCapture(amount Money, providerRef string) error {
	if p.status != CapturePending && p.status != Authorized {
		return ErrInvalidTransition
	}
	if amount.Currency() != p.authorized.Currency() || amount.Amount() > p.RemainingCapture().Amount() {
		return ErrCaptureExceedsAuthorization
	}
	p.captured = p.captured.AddMust(amount)
	p.status = Captured
	p.providerRef = providerRef
	return nil
}
~~~

Đây là conceptual snippet; production code không nên dùng `AddMust` nếu overflow có thể xảy ra.

## Ports Và Ownership

~~~go
type Gateway interface {
	Authorize(ctx context.Context, req AuthorizeRequest) (AuthorizeResult, error)
	Lookup(ctx context.Context, operationID string) (OperationResult, error)
}
~~~

Application sở hữu interface và result semantic. Adapter provider map SDK status, timeout và error code sang taxonomy nội bộ. Domain không import Stripe/Adyen SDK; application cũng không nhận raw webhook JSON.

~~~mermaid
flowchart LR
    API["Payment HTTP API"] --> APP["Payment use cases"]
    WEBHOOK["Verified webhook adapter"] --> APP
    APP --> DOMAIN["Payment aggregate"]
    APP --> GW["PaymentGateway port"]
    SDK["Provider adapter"] -.implements.-> GW
    PG["Payment repository"] -.implements.-> APP
~~~

## Authorize Flow An Toàn

1. Nhận merchant, order, amount, currency và idempotency key.
2. Trong transaction ngắn, insert operation với unique `(merchant_id, idempotency_key)` và request hash.
3. Nếu key đã có: cùng hash trả trạng thái cũ; khác hash trả conflict.
4. Commit `AuthorizationPending` trước khi gọi provider.
5. Gọi provider với stable operation ID làm idempotency key.
6. Ghi kết quả trong transaction mới.
7. Nếu timeout/connection reset sau khi request đã gửi, ghi `Unknown`, không tự kết luận failed.

Việc commit pending trước network call tạo khả năng recovery. Một worker có thể lookup provider bằng operation ID sau khi process crash.

## Webhook Là Delivery Adapter

Webhook adapter phải đọc raw body có giới hạn, verify signature/timestamp trước JSON mapping, chống replay và lưu `provider_event_id` unique. Nó map provider event thành application command. Raw payload không phải Domain Event; nó là external contract chưa được tin cậy.

Webhook và synchronous response có thể đua. Cả hai gọi transition idempotent, lock/version cùng Payment và lưu evidence. Transition cũ bị ignore có audit reason; không cập nhật state theo "last write wins".

## Ambiguous Outcome

Các lỗi trước khi gửi request, như local validation, là safe-to-retry. Các lỗi sau khi byte có thể đã tới provider là ambiguous. Retry chỉ an toàn khi provider cam kết deduplicate cùng key và local key vẫn ổn định.

Nếu provider không có idempotency API, strategy có thể là:

- Lookup theo merchant reference trước khi retry.
- Đưa operation vào manual review khi không thể query.
- Chấp nhận at-most-once và UX pending tùy risk.
- Không giả vờ có exactly-once bằng một mutex trong process.

## Persistence Và Audit

- `payments`: aggregate hiện tại, amount/currency/status/version.
- `payment_operations`: mỗi authorize/capture/refund, idempotency hash, attempt, provider ref, outcome.
- `provider_events`: inbox webhook, signature metadata, processed state.
- `outbox`: event nội bộ như `PaymentCaptured`.
- Append-only audit log không thay thế authoritative state; quyền truy cập và retention phải rõ.

Sensitive data cần redaction từ adapter, logger và tracing. Không đưa raw Authorization header, signature secret hoặc toàn bộ provider payload vào log.

## Failure Matrix

| Failure | Không được làm | Recovery đúng |
|---|---|---|
| Validation fail | gọi provider | trả business/input error |
| Timeout trước connect | tạo key mới | retry cùng operation ID |
| Timeout sau write | mark failed | mark unknown, lookup/reconcile |
| DB fail sau provider success | gọi lại không key | reconcile bằng stable key/ref |
| Webhook duplicate | capture/refund lại | unique inbox + idempotent transition |
| Webhook giả | parse rồi xử lý | verify raw body trước mapping |
| Provider degraded | giữ DB tx mở | circuit/bulkhead có đo lường, pending workflow |

## Error Taxonomy

Phân biệt `Declined` (business outcome), `Unavailable` (dependency tạm thời), `UnknownOutcome`, `Conflict`, `InvalidTransition` và `Internal`. HTTP adapter có thể map declined thành stable 422/409 theo contract; không trả nguyên provider error cho client.

Retry policy chỉ áp dụng theo operation. Retry `GET lookup` khác retry `POST authorize`; không gắn một generic retry middleware vào mọi method.

## Testing Strategy

- Domain transition/property test cho partial capture/refund và giới hạn amount.
- Use-case fake gateway mô phỏng success, decline, timeout-before-send, timeout-after-send.
- Adapter contract test từ provider fixtures, bao gồm unknown enum và schema evolution.
- HTTP/webhook test signature, stale timestamp, duplicate event, body limit.
- PostgreSQL integration test unique idempotency, concurrent duplicate request, optimistic lock.
- Recovery test crash ở mỗi điểm: sau pending commit, sau provider response, trước result commit.
- Security test log capture để chứng minh secret/sensitive field bị redact.

## Observability Và Runbook

Metrics cần authorization outcome, unknown age, reconciliation success, webhook lag/duplicates và provider latency theo operation type. High-cardinality payment ID chỉ ở trace/log, không ở metric label. Alert trên số operation `Unknown` quá SLA, không chỉ HTTP 5xx.

Runbook phải cho phép operator lookup local operation/provider reference, chạy reconciliation idempotent, và lưu actor/reason khi manual override. Không sửa row trực tiếp không audit.

## Trade-off

Một checkout đơn giản qua hosted payment page có thể giữ Payment model mỏng hơn. Multi-provider routing, partial capture và dispute có thể đáng tách bounded context. Clean Architecture không yêu cầu microservice; module payment trong monolith vẫn có thể có port/provider adapter rõ.

## Câu Hỏi Mastery

1. Vì sao timeout là trạng thái kiến thức của hệ thống chứ không phải kết quả business?
2. Idempotency key cần scope và request hash nào?
3. Webhook event khác Domain Event ra sao?
4. Nếu synchronous response là success nhưng webhook failure đến sau, transition nào thắng?
5. Circuit breaker bảo vệ tài nguyên nào và có thể làm UX tệ hơn thế nào?

## Bài Thực Hành

Lập crash matrix cho authorize với mọi điểm trước/sau external call và DB commit. Thiết kế schema operation đủ để một process mới phục hồi mà không dựa vào memory của process cũ.
