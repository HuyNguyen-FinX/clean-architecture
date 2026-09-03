# External Service Gateway: timeout, retry và ambiguous outcome

External API là boundary không đáng tin: latency, schema, status, rate limit và outage nằm ngoài control. Gateway port nên nói bằng application intent, còn adapter hấp thụ SDK/HTTP details và anti-corruption mapping.

## Kết quả học tập

- thiết kế intent-based gateway port;
- map vendor model/error thành application semantics;
- phân bổ timeout/retry/circuit-breaker budget;
- xử lý non-idempotent và unknown outcome;
- test adapter bằng httptest contract;
- thiết kế reconciliation/observability.

## 1. Problem

~~~go
type PaymentUseCase struct {
	stripe *stripe.Client
}

func (uc *PaymentUseCase) Execute(ctx context.Context, cmd Command) error {
	params := &stripe.PaymentIntentParams{...}
	_, err := uc.stripe.PaymentIntents.New(params)
	return err
}
~~~

Application biết SDK, vendor fields/status và retry semantics. Test business workflow mock SDK details. Đổi provider làm core đổi.

## 2. Ba level

### Level 1

Gateway là translator/protective boundary giữa intent của app và API provider.

### Level 2

Adapter quản HTTP client, auth, request/response, timeout, retry, rate limit, error mapping và metrics.

### Level 3

Remote call không nằm trong local transaction. Timeout có thể tạo ambiguous outcome. Application cần state machine, idempotency và reconciliation thay vì giả vờ synchronous atomicity.

## 3. Port theo intent

~~~go
type PaymentGateway interface {
	Authorize(
		ctx context.Context,
		request AuthorizationRequest,
	) (AuthorizationResult, error)
}
~~~

Không:

~~~go
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}
~~~

HTTPClient là dependency hữu ích bên trong adapter test, nhưng không phải application port nếu app muốn Authorize.

## 4. Anti-Corruption Layer

~~~text
Application AuthorizationRequest
→ Vendor request mapper
→ HTTP/SDK
→ Vendor response
→ status/error mapper
→ Application AuthorizationResult
~~~

Provider field “captured”, “requires_action”, “declined” không đi thẳng vào domain enum nếu meanings không đồng nhất.

## 5. HTTP client production

~~~go
client := &http.Client{
	Timeout: 2 * time.Second,
	Transport: transport,
}
~~~

Reuse client/transport; không tạo per request. Cấu hình dial/TLS/header/response timeouts tùy control. Luôn close response body; giới hạn body; kiểm status trước decode; không log Authorization.

## 6. Timeout budget

Request budget 2s:

~~~text
ingress 100ms
DB 300ms
provider attempts + backoff 1200ms
response margin 400ms
~~~

Ba retry mỗi retry timeout 2s là sai budget. Mỗi attempt deadline là min(parent remaining, per-attempt cap).

## 7. Retry

Retry GET/idempotent operation hoặc POST có provider idempotency key. Classify:

- network connect/reset trước response: có thể retry;
- 429/503 với Retry-After: bounded;
- 4xx business decline: không;
- context cancel: không;
- malformed response: thường không retry mù;
- timeout sau request gửi: outcome có thể unknown.

Backoff exponential + jitter + cap. Tránh cả SDK và application cùng retry.

## 8. Idempotency key

Application tạo stable operation key, adapter map sang provider header. Cùng key phải cùng request parameters. Persist relationship local operation ↔ provider key/ID để inquiry/reconciliation.

Random key mới mỗi retry phá idempotency.

## 9. Ambiguous outcome

~~~text
POST authorize
provider charge thành công
response mất
client timeout
~~~

Không trả “declined”. Result cần Unknown/Pending:

~~~go
type AuthorizationStatus string

const (
	Authorized AuthorizationStatus = "authorized"
	Declined   AuthorizationStatus = "declined"
	Pending    AuthorizationStatus = "pending"
)
~~~

Application persist Pending và enqueue inquiry. Reconciliation hỏi provider bằng idempotency key/operation ID.

## 10. Circuit breaker

Breaker ngăn gọi dependency đang lỗi liên tục, bảo vệ thread/goroutine/pool và giảm recovery load. State closed/open/half-open.

Không dùng breaker để biến lỗi thành success. Không dùng một global breaker cho mọi endpoint/tenant nếu failure domain khác. Metrics/alerts và manual override cần cân nhắc.

## 11. Bulkhead và concurrency limit

Giới hạn concurrent calls tới provider để một dependency chậm không chiếm hết goroutine/connection. Queue phải bounded; khi đầy fail fast với retryable/unavailable semantics.

## 12. Rate limit

Provider 429 có thể có Retry-After. Central limiter tránh mỗi replica tự vượt quota tổng. Nhưng distributed limiter thêm dependency; quota partition/local budget có thể đơn giản hơn.

## 13. Error mapping

~~~go
switch response.StatusCode {
case http.StatusOK:
	return mapSuccess(response)
case http.StatusUnprocessableEntity:
	return AuthorizationResult{Status: Declined}, nil
case http.StatusTooManyRequests:
	return Result{}, retryable(ErrProviderRateLimited)
default:
	return Result{}, fmt.Errorf("provider status %d: %w", response.StatusCode, ErrProvider)
}
~~~

Không trả raw provider body; có thể chứa PII. Correlation ID provider nên giữ trong diagnostics/result metadata phù hợp.

## 14. Local transaction + remote call

Không giữ DB row lock trong provider latency. Workflow:

1. local transaction tạo Payment Pending + outbox command;
2. worker gọi provider với idempotency key;
3. local transaction apply result;
4. timeout Unknown dẫn tới inquiry;
5. reconciliation đảm bảo eventual terminal state.

Compensation không thật sự đảo mọi remote effect; refund là business operation mới có thể fail.

## 15. Testing

- mapper fixtures cho provider versions;
- httptest.Server kiểm method/header/body/status/timeout/body limit;
- fake clock/backoff;
- retry count/idempotency key stable;
- ambiguous timeout;
- circuit state;
- contract sandbox integration;
- chaos/network fault.

Mock Do expectation không chứng minh TLS/DNS/proxy, nhưng httptest đủ tốt cho HTTP adapter behavior.

## 16. Production scenario

Provider p99 tăng từ 300ms lên 5s:

- timeout giới hạn 800ms;
- retries có thể nhân traffic;
- breaker mở;
- bounded queue reject;
- payments chuyển Pending;
- inquiry/reconciliation xử lý;
- API trả 202 + operation ID thay vì 500 giả thất bại;
- SLO/alerts theo pending age.

## 17. Debug

1. local operation/idempotency key;
2. provider correlation ID;
3. attempt timeline/deadline;
4. request hash;
5. response status/truncated safe body;
6. breaker/limiter state;
7. local pending/outbox;
8. inquiry result.

Không log card/token/secret.

## 18. Khi nào không tạo gateway?

Simple read-only API dùng một chỗ có thể gọi typed client trực tiếp ở application nếu coupling chấp nhận. Gateway đáng giá khi vendor volatility, semantic mapping, testing hoặc provider swap/failure workflow là thật. Không tạo interface mirror 50 SDK methods.

## 19. Bài tập

1. Implement Authorize adapter bằng httptest.Server.
2. Test stable idempotency key qua retry.
3. Thiết kế Pending reconciliation.
4. Phân bổ 2s timeout budget.
5. Review circuit breaker placement.

## 20. Mastery questions

1. Intent port khác HTTPClient abstraction?
2. Timeout vì sao không luôn là failure?
3. Retry non-idempotent gây gì?
4. Breaker không giải quyết điều gì?
5. Tại sao không gọi provider trong DB transaction?
6. Refund có phải rollback không?
7. Khi nào trả 202 hợp lý?
8. Anti-Corruption Layer bảo vệ semantics nào?

## Further reading

- Go net/http Transport/Client docs.
- RFC 9110 retry/idempotent methods và Retry-After.
- Release It!, circuit breaker/bulkhead.
- Provider-specific idempotency documentation.

## Quality gate

- [x] Intent port và anti-corruption mapping
- [x] HTTP client/timeout/retry/idempotency
- [x] Ambiguous outcome/circuit/bulkhead
- [x] Workflow, tests, production/debug
- [x] Trade-off, exercises, mastery
