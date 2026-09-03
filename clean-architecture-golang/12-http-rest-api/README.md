# HTTP REST API trong Go: contract chặt, core độc lập

HTTP Handler là public compatibility boundary. Một handler production cần chống malformed/oversized input, map error ổn định, truyền cancellation, bảo vệ internal detail và vẫn không chứa business invariant.

## Kết quả học tập

- thiết kế route/method/status/idempotency semantics;
- strict JSON decode bằng net/http;
- map DTO → command và error → safe response;
- cấu hình server timeouts/lifecycle;
- test contract bằng httptest;
- xử lý cancellation, retry và ambiguous response.

## 1. Problem: decode được chưa đủ an toàn

~~~go
var req TransferRequest
_ = json.NewDecoder(r.Body).Decode(&req)
err := useCase.Execute(context.Background(), req)
http.Error(w, err.Error(), 500)
~~~

Các failure:

- decode error bị bỏ;
- body không giới hạn;
- unknown/trailing fields có thể lọt;
- request cancellation bị mất;
- DTO đi thẳng vào application;
- mọi error thành 500 và leak detail;
- response có thể WriteHeader sau khi body bắt đầu.

## 2. Ba level

### Level 1: trực giác

Handler nhận “gói hàng” HTTP, kiểm bao bì, dịch nội dung thành command, gọi use case rồi đóng gói outcome.

### Level 2: Backend Engineer

Quan tâm method, content type, body size, JSON, headers, timeout, status, response schema, middleware, tests.

### Level 3: Architecture/API Design

HTTP contract tiến hóa độc lập domain. Status/code/idempotency/versioning là public semantics. DTO tách core khỏi transport shape; application error taxonomy tách workflow khỏi status code.

## 3. Runtime và source dependency

~~~mermaid
sequenceDiagram
    participant C as Client
    participant H as HTTP Adapter
    participant U as Use Case
    participant R as Repository
    C->>H: POST /transfers + JSON
    H->>H: limit, decode, map
    H->>U: Execute(request context, command)
    U->>R: calls through port
    R-->>U: result/error
    U-->>H: outcome
    H-->>C: stable status + JSON
~~~

Compile time: http adapter → application → domain. Application không import net/http.

## 4. Strict decoder

Mini-banking implementation ở [handler.go](../examples/mini-banking/internal/account/delivery/http/handler.go):

~~~go
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errors.New("body must contain one JSON value")
	}
	return nil
}
~~~

Production decoder cần cân nhắc:

- require application/json hay chấp nhận empty Content-Type;
- body max theo endpoint;
- unknown field compatibility;
- empty body;
- một JSON object duy nhất;
- number precision nếu vượt int64/float;
- error detail public vừa đủ.

DisallowUnknownFields giúp phát hiện typo nhưng có thể làm forward compatibility khó hơn với tolerant readers. Chọn theo API policy, không máy móc.

## 5. DTO và mapping

~~~go
type transferRequest struct {
	FromAccountID string
	ToAccountID   string
	Amount        int64
	Currency      string
}

cmd := application.TransferMoneyCommand{
	FromAccountID: req.FromAccountID,
	ToAccountID:   req.ToAccountID,
	Amount:        req.Amount,
	Currency:      req.Currency,
}
~~~

Mapping vài dòng là intentional boundary. Không dùng domain.Account làm request vì client không được gửi balance/status tùy ý.

## 6. Status semantics

| Outcome | HTTP |
|---|---:|
| Transfer created synchronously | 201 |
| Accepted async, chưa hoàn tất | 202 |
| Malformed/invalid input | 400 |
| Missing/invalid auth | 401 |
| Authenticated nhưng forbidden | 403 |
| Account không tồn tại | 404 |
| Conflict với state | 409 |
| Unsupported media type | 415 |
| Rate limited | 429 |
| Deadline/upstream unavailable | 503/504 theo policy |
| Unknown internal | 500 |

Không chọn 200 cho mọi thứ. Nhưng REST không yêu cầu một status duy nhất cho domain error; consistency trong API quan trọng hơn tranh luận giáo điều.

## 7. Stable error response

~~~json
{
  "code": "insufficient_balance",
  "message": "account balance is insufficient"
}
~~~

Code là machine-readable contract. Message có thể localization/change. Internal error:

~~~go
return http.StatusInternalServerError, errorResponse{
	Code: "internal_error",
	Message: "an internal error occurred",
}
~~~

Không đưa SQL, DSN, stack, account secret vào response. Log outer boundary với request/trace ID và wrapped cause.

## 8. Validation layers

Handler kiểm wire syntax. Command/domain kiểm meaning. Ví dụ amount=0 có thể bị handler reject để UX tốt, nhưng Account/Money vẫn phải reject để entry point khác không bypass.

Tránh duplicated constants drift. Nếu maximum transfer là business policy thay đổi theo account tier, không hardcode ở struct tag.

## 9. Context và timeout

Luôn dùng r.Context:

~~~go
err := h.transfer.Execute(r.Context(), cmd)
~~~

Server timeouts:

~~~go
server := &http.Server{
	Addr:              ":8080",
	Handler:           routes,
	ReadHeaderTimeout: 5 * time.Second,
	ReadTimeout:       10 * time.Second,
	WriteTimeout:      15 * time.Second,
	IdleTimeout:       60 * time.Second,
}
~~~

Timeout budget phải xét load balancer/client/database. WriteTimeout không thay business idempotency. Client disconnect sau commit tạo ambiguous outcome.

## 10. Idempotency-Key

Money transfer retry cần key:

~~~text
POST /transfers
Idempotency-Key: client-generated-key
~~~

Rules phải định nghĩa:

- required/format/length;
- scope theo tenant/operation;
- cùng key khác request hash trả conflict;
- concurrent duplicate chờ/replay thế nào;
- response nào được cache;
- retention;
- durable record atomic với transfer.

Handler parse header vào command. Application và durable store sở hữu state machine; in-memory middleware không đủ qua restart/replica.

## 11. Authentication/authorization

Middleware verify credential. Use case nhận Actor đã normalize. 401 và 403 không nên leak resource existence ngoài threat model.

Không tin X-User-ID trực tiếp trừ khi trusted proxy contract xác thực và strip client header.

## 12. Pagination

Transaction history nên dùng cursor:

~~~json
{
  "items": [],
  "next_cursor": "opaque-token"
}
~~~

Cursor encode stable sort keys như created_at + id, có signature/version nếu public. Offset dễ skip/duplicate khi dữ liệu mới chèn và query sâu tốn cost.

Read model không cần Aggregate Repository.

## 13. Response writing

Set headers trước WriteHeader. Encode response error sau WriteHeader không thể đổi status. Với small DTO ổn định, marshal trước rồi write:

~~~go
payload, err := json.Marshal(body)
if err != nil {
	return err
}
w.Header().Set("Content-Type", "application/json; charset=utf-8")
w.Header().Set("X-Content-Type-Options", "nosniff")
w.WriteHeader(status)
_, _ = w.Write(append(payload, '\n'))
~~~

Không phải handler nào cũng cần generic responder phức tạp.

## 14. Middleware order

Một order thường thấy:

~~~text
request ID
→ panic recovery
→ access log/metrics
→ security headers
→ body/timeout limits
→ authentication
→ authorization/use case
~~~

Recovery phải nằm ngoài code có thể panic; observability phải thấy recovered outcome. Test order cho request ID trong panic response/log.

## 15. Graceful shutdown

Khi SIGTERM:

1. readiness false;
2. HTTP server ngừng nhận connection mới;
3. Server.Shutdown với timeout;
4. đợi in-flight request;
5. stop workers;
6. close DB/client/telemetry.

Không dùng log.Fatal sau khi đã tạo defer cleanup vì os.Exit bỏ qua defer.

## 16. HTTP test bằng httptest

Test matrix:

- happy status/body/command mapping;
- malformed, unknown và trailing JSON;
- wrong method/media type;
- body too large;
- domain/application error mapping;
- unknown error không leak;
- canceled context propagation;
- headers/content type;
- nil dependency constructor.

Code chạy được: [handler_test.go](../examples/mini-banking/internal/account/delivery/http/handler_test.go).

Test fake use case, không cần DB. E2E riêng xác minh toàn graph.

## 17. Production failure scenario

Commit transfer thành công, client timeout trước response:

- server log thấy success;
- client không biết, retry;
- không có idempotency: double transfer;
- có durable key: replay same outcome;
- GET status endpoint theo transfer ID giúp reconciliation.

HTTP timeout là transport event, không chứng minh transaction rollback.

## 18. Debug

1. request ID/trace ID;
2. sanitized method/route/status/latency;
3. command mapping;
4. context deadline remaining;
5. use-case wrapped cause;
6. response code mapping;
7. proxy/load balancer timeout;
8. idempotency record/transfer status.

Không log raw body mặc định vì PII/secrets và cost.

## 19. Khi nào không cần REST abstraction lớn?

Internal service có một endpoint có thể dùng net/http trực tiếp. Không cần framework, generic Controller/BaseResponse hoặc OpenAPI generation nếu cost lớn hơn compatibility needs. Router/framework là adapter detail; chọn theo middleware/ecosystem/team.

## 20. Lab và bài tập

Làm [Lab 05: HTTP](../labs/lab-05-http/README.md):

1. strict decoder;
2. DTO mapping;
3. safe error taxonomy;
4. httptest matrix;
5. cancellation;
6. idempotency header challenge.

## 21. Mastery questions

1. Handler “mỏng” vẫn chịu những responsibility nào?
2. Unknown JSON field nên reject trong mọi API không?
3. 201 khác 202 về guarantee gì?
4. HTTP timeout sau commit tạo failure nào?
5. Idempotency state đặt ở middleware có đủ không?
6. Vì sao r.Context phải đi tới DB?
7. Struct tag validation thay domain invariant được không?
8. Offset pagination lỗi gì khi list thay đổi?

## Further reading

- Go net/http, encoding/json, httptest và context documentation.
- RFC 9110 HTTP Semantics.
- OWASP REST Security Cheat Sheet.
- OpenAPI specification khi contract cần generate/validate.

## Quality gate

- [x] Strict executable handler và tests
- [x] DTO/command, status, safe error model
- [x] Context, timeout, auth, pagination, idempotency
- [x] Runtime/source dependency
- [x] Production failure/debug/lifecycle
- [x] Trade-off, lab, mastery, references
