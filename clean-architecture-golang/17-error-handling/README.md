# Error Handling: error là boundary language, không chỉ là string

Trong Go, error vừa giữ cause kỹ thuật vừa biểu diễn outcome mà boundary khác cần hiểu. Thiết kế tốt cho phép domain độc lập HTTP, application phân nhánh có chủ đích, adapter giữ diagnostics và client chỉ nhận public contract an toàn.

## Kết quả học tập

- phân loại Domain/Application/Infrastructure/Transport error;
- dùng sentinel, typed error, errors.Is/As và wrapping;
- map error tại boundary;
- giữ retryability và unknown/ambiguous semantics;
- tránh double logging và data leak;
- test error chain thay vì so string.

## 1. Problem

~~~go
return fmt.Errorf("HTTP 409: SQL update failed: insufficient balance")
~~~

Một error trộn:

- business rule;
- HTTP status;
- SQL operation;
- public message.

Domain dùng lại từ Kafka sẽ mang HTTP 409 vô nghĩa. Client thấy SQL detail. Code phân nhánh bằng string dễ vỡ khi wrap.

## 2. Ba level

### Level 1: trực giác

Error trả lời hai câu: chuyện gì xảy ra với caller, và nguyên nhân nào cần giữ để điều tra.

### Level 2: Backend Engineer

Go cần error chain, stable identity/type, contextual wrapping, mapping và logging policy.

### Level 3: Architecture

Mỗi boundary dịch vocabulary:

~~~text
pgx no rows → repository not found semantics
domain insufficient balance → application outcome
application outcome → HTTP 409 / gRPC FailedPrecondition
unknown cause → safe internal response + diagnostic trace
~~~

Không phải layer nào cũng wrap/map/log.

## 3. Taxonomy

### Domain error

Vi phạm rule/invariant:

~~~go
var ErrInsufficientBalance = errors.New("insufficient balance")
~~~

Không biết HTTP/Kafka/SQL.

### Application error

Workflow semantics: command invalid, idempotency conflict, authorization, retry exhausted.

~~~go
type ConflictError struct {
	Resource string
	Cause    error
}
~~~

### Infrastructure error

Timeout, driver failure, corrupt row, broker unavailable. Adapter wrap operation và giữ cause/SQLSTATE.

### Transport error

Malformed JSON, unsupported media type, invalid protobuf, route/method. Nó có thể không đi vào use case.

## 4. Sentinel error

~~~go
var ErrAccountNotFound = errors.New("account not found")

if errors.Is(err, ErrAccountNotFound) { ... }
~~~

Phù hợp khi caller chỉ cần category không data. Sentinel là public API; đổi semantics cần cẩn thận.

Không so:

~~~go
if err.Error() == "account not found" { ... }
~~~

Wrapping làm string đổi; localization/detail cũng phá.

## 5. Typed error

~~~go
type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Reason
}
~~~

Caller:

~~~go
var validation *ValidationError
if errors.As(err, &validation) {
	// map fields to public violations
}
~~~

Type hữu ích khi mapping cần structured data. Không nhồi stack/status/retry vào một God Error.

## 6. Wrapping

~~~go
account, err := repo.FindByID(ctx, id)
if err != nil {
	return fmt.Errorf("load source account %q: %w", id, err)
}
~~~

%w giữ chain cho errors.Is/As. Context nên thêm operation/resource không nhạy cảm. Tránh wrap cùng phrase ở mọi function tạo message dài vô nghĩa.

Không dùng %v nếu caller cần cause identity.

## 7. Join và cleanup error

Transaction callback fail và rollback cũng fail:

~~~go
return errors.Join(operationErr, fmt.Errorf("rollback: %w", rollbackErr))
~~~

errors.Is có thể tìm cả cause. Nhưng response vẫn dựa vào primary semantics; telemetry ghi cleanup failure riêng.

## 8. Error mapping

Mini-banking HTTP adapter:

~~~go
switch {
case errors.Is(err, domain.ErrAccountNotFound):
	return 404, publicError("account_not_found")
case errors.Is(err, domain.ErrInsufficientBalance):
	return 409, publicError("insufficient_balance")
default:
	return 500, publicError("internal_error")
}
~~~

Không trả err.Error ở default. Public code ổn định hơn internal message.

gRPC adapter map cùng application semantics sang codes.NotFound/FailedPrecondition/Internal. Kafka adapter map retryable/permanent outcome sang ack/retry/DLQ.

## 9. Retryable không đồng nghĩa 500

Typed wrapper:

~~~go
type RetryableError struct {
	Cause error
}

func (e *RetryableError) Error() string { return e.Cause.Error() }
func (e *RetryableError) Unwrap() error { return e.Cause }
~~~

Chỉ adapter hiểu SQLSTATE/HTTP status provider để classify. Application policy quyết định retry trong budget. Domain rejection không retry.

Không retry unknown non-idempotent outcome.

## 10. Ambiguous outcome

Payment request timeout có ba khả năng:

- provider chưa nhận;
- provider đang xử lý;
- provider đã charge nhưng response mất.

Gộp thành “failed” và retry ngay có thể double charge. Gateway cần outcome Unknown/lookup handle:

~~~go
type AuthorizationResult struct {
	Status    AuthorizationStatus
	ProviderID string
}
~~~

Application chuyển state pending-reconciliation thay vì giả vờ rollback remote.

## 11. Cancellation/deadline

Giữ errors.Is với context.Canceled/DeadlineExceeded qua wrap. HTTP mapping tùy policy, nhưng metrics cần tách caller cancel, server deadline và dependency timeout.

Không log canceled request như server error ở mức error nếu đó là normal client behavior; vẫn đo.

## 12. Logging policy

Rule thực dụng: log một error tại boundary có đủ context để hành động. Inner layer wrap, không log rồi return trừ khi nó xử lý/recover hoặc cần event độc lập.

Sai:

~~~text
repository logs ERROR
application logs ERROR
handler logs ERROR
middleware logs ERROR
~~~

Một failure tạo bốn alert.

Structured fields: operation, error, trace_id, safe resource ID, latency/retry count. Không log secret/raw body.

## 13. Panic

Panic dành cho programmer invariant hoặc unrecoverable initialization, không phải business rejection. HTTP recovery middleware ngăn process crash và trả safe 500, nhưng vẫn cần stack/metric.

Constructor panic nil required dependency có thể chấp nhận vì composition bug. Config invalid nên thường trả error để startup báo rõ.

## 14. Wrong patterns

- map mọi DB error thành not-found;
- stringify rồi mất cause;
- public response lộ driver;
- typed error có HTTP status trong domain;
- ignore cleanup/commit error;
- retry mọi error;
- error package toàn cục chứa mọi layer.

## 15. Testing errors

~~~go
if !errors.Is(err, domain.ErrInsufficientBalance) {
	t.Fatalf("got %v", err)
}
~~~

Test:

- identity survives wrapping;
- typed details map đúng;
- unknown cause không leak;
- retry classifier;
- context cause;
- HTTP/gRPC/Kafka outcomes;
- commit/rollback errors.

Không snapshot full internal error string nếu wording không phải contract.

## 16. Production scenario

Postgres timeout trong transfer:

1. adapter wrap operation và giữ context deadline;
2. transaction rollback;
3. application không retry nếu budget hết;
4. HTTP trả safe 503/504 hoặc 500 theo API policy;
5. log một lần có trace/pool/SQLSTATE;
6. client retry chỉ an toàn với idempotency key.

## 17. Debug

1. bắt đầu từ public code/request ID;
2. tìm trace/log outer boundary;
3. đi theo unwrap chain;
4. xác định layer đã map category;
5. kiểm errors.Is/As có bị %v làm đứt;
6. xem retry/commit ambiguity;
7. bổ sung test tại boundary gây mất semantics.

## 18. Khi nào một error string là đủ?

CLI nhỏ/internal script có thể chỉ cần fmt.Errorf. Không cần taxonomy lớn nếu không có caller programmatic hoặc nhiều boundary. Nhưng khi API, retry và observability dựa vào category, string là contract quá yếu.

## 19. Bài tập

1. Thiết kế typed IdempotencyConflict có request hash.
2. Map cùng errors sang HTTP/gRPC/Kafka.
3. Tìm một error chain bị đứt bởi %v.
4. Viết retry classifier cho PostgreSQL SQLSTATE.
5. Phân tích timeout external payment thành rejected/unknown.

## 20. Mastery questions

1. Sentinel và typed error khác nhau khi nào?
2. Layer nào map pgx.ErrNoRows?
3. Vì sao log-and-return ở mọi layer gây hại?
4. Unknown payment outcome khác failure thế nào?
5. errors.Join có ích ở cleanup ra sao?
6. Domain error chứa HTTP status phá reuse nào?
7. Test error string khi nào hợp lý?
8. Retryability thuộc adapter hay application?

## Further reading

- Go package errors, Error handling and Go 1.13 error wrapping.
- Go context package.
- RFC 9457 Problem Details nếu chọn HTTP problem format.
- PostgreSQL SQLSTATE appendix.

## Quality gate

- [x] Taxonomy, three-level reasoning
- [x] Sentinel/type/wrap/join/cancellation
- [x] Cross-boundary mapping và safe responses
- [x] Retry/ambiguous outcomes/logging/panic
- [x] Tests, production failure, debugging
- [x] Trade-off, exercises, mastery
