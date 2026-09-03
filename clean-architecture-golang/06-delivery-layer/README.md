# Delivery Layer: dịch protocol thành application intent

Delivery Layer là nhóm driving adapters đưa input từ HTTP, gRPC, Kafka consumer, CLI hoặc scheduler vào Application. “Mỏng” không có nghĩa chỉ forward; adapter vẫn sở hữu protocol contract, parsing, authentication extraction, limits và error mapping.

## Kết quả học tập

- phân biệt transport DTO, command và domain model;
- đặt validation/error/auth đúng boundary;
- truyền context nhưng không kéo framework context vào core;
- thêm delivery adapter mới mà không sửa domain;
- test adapter bằng contract/transport harness;
- xử lý overload, timeout và malformed input.

## 1. Problem

~~~go
func TransferHandler(w http.ResponseWriter, r *http.Request) {
	var account domain.Account
	json.NewDecoder(r.Body).Decode(&account)
	db.Exec("UPDATE accounts ...")
	if account.Balance < 0 { /* rule */ }
	kafka.Publish(...)
}
~~~

Handler sở hữu HTTP, SQL, domain rule và event delivery. Khi thêm gRPC, team phải copy rule hoặc gọi lại handler giả lập. Test một rule cần HTTP + DB + Kafka.

Chuỗi coupling:

~~~text
protocol DTO = domain entity
        ↓
external caller có thể shape core state
        ↓
handler giữ business decisions
        ↓
entry point mới phải duplicate
        ↓
consistency giữa adapters bị lệch
~~~

## 2. Ba level

### Level 1: trực giác

Delivery Adapter là phiên dịch viên. Nó hiểu ngôn ngữ bên ngoài và chuyển thành câu lệnh application.

### Level 2: Backend Engineer

Adapter xử lý:

- route/method/content type/body limit;
- decode và syntax validation;
- auth claims/correlation data;
- DTO → command;
- call use case với request context;
- application result/error → protocol response;
- middleware transport-level.

### Level 3: Architecture

Delivery là policy-free ở cấp business, nhưng chứa protocol policy. HTTP status, protobuf compatibility hay Kafka offset đều là quyết định quan trọng của adapter. Compile-time dependency đi vào application; runtime request đi từ adapter đến use case.

## 3. Boundary flow

~~~mermaid
flowchart LR
    EXT["External request/message"] --> PARSE["Parse + limits"]
    PARSE --> DTO["Transport DTO"]
    DTO --> MAP["Map"]
    MAP --> CMD["Application command"]
    CMD --> UC["Use Case"]
    UC --> RESULT["Result/error"]
    RESULT --> RESPONSE["Protocol response/ack"]
~~~

Không truyền http.Request, gin.Context hay generated protobuf message vào use case. context.Context là exception có chủ đích cho cancellation/deadline/request metadata.

## 4. DTO khác command khác Entity

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

DTO phản ánh external schema và backward compatibility. Command phản ánh use-case intent. Entity giữ private state/invariant.

Không phải lúc nào cũng cần ba struct. Nếu read-only endpoint có cùng projection và không có security/evolution pressure, reuse DTO/result có thể hợp lý. Tách khi models có lý do thay đổi khác nhau.

## 5. Validation ownership

| Kiểm tra | Nơi |
|---|---|
| JSON/protobuf malformed | Delivery |
| body quá lớn/content type sai | Delivery |
| required field/format wire | Delivery hoặc command constructor |
| actor có quyền gọi operation | Application |
| amount dương/currency hợp lệ | Domain Value Object |
| không vượt overdraft | Account invariant |

Đừng chỉ validate domain rule bằng struct tag. Một Kafka consumer hoặc internal call có thể bypass HTTP.

## 6. Authentication và authorization

Middleware có thể verify JWT/signature và đưa identity đã xác thực vào typed request context. Adapter map thành application Actor:

~~~go
actor := application.Actor{Subject: claims.Subject, Roles: claims.Roles}
err := useCase.Execute(r.Context(), actor, cmd)
~~~

HTTP middleware sở hữu token/cookie. Application sở hữu rule “actor có được transfer từ account này không”. Domain có thể sở hữu permission nếu nó là core business concept.

Không truyền raw JWT xuống domain.

## 7. Context propagation

~~~text
HTTP/Kafka lifecycle
→ Application
→ Repository/Gateway
→ driver/client
~~~

Context mang cancellation, deadline, trace/request metadata. Không nhét optional business parameters vào context. Không gọi context.Background giữa chain. Account.Withdraw không cần context vì computation thuần không chờ I/O.

Kafka consumer cần context gắn với process/partition/message lifecycle; cancel đúng lúc shutdown nhưng phải quyết định ack/offset semantics.

## 8. Error mapping

Delivery map stable semantics:

~~~text
ErrAccountNotFound       → HTTP 404 / gRPC NotFound
ErrInsufficientBalance   → HTTP 409 / gRPC FailedPrecondition
Malformed JSON           → HTTP 400
Unknown infrastructure   → safe 500 / gRPC Internal
~~~

Domain không import net/http hoặc grpc/status. Unknown error không trả err.Error cho caller. Internal cause đi vào log/trace một lần ở outer boundary.

## 9. HTTP, gRPC, Kafka consumer là ngang hàng

~~~mermaid
flowchart LR
    HTTP["HTTP adapter"] --> UC["Application Use Case"]
    GRPC["gRPC adapter"] --> UC
    KAFKA["Kafka consumer adapter"] --> UC
    CLI["CLI adapter"] --> UC
~~~

Kafka consumer thường được xếp infrastructure theo organizational naming, nhưng về Ports and Adapters nó là driving adapter. Tên folder ít quan trọng hơn direction và responsibility.

## 10. Ack/response là protocol decision

HTTP trả response sau use case. Kafka consumer chỉ commit offset khi outcome phù hợp:

- success: ack/commit;
- retryable failure: không commit, retry/backoff;
- permanent invalid message: DLQ/quarantine rồi commit theo policy;
- context shutdown: dừng mà không làm mất message.

Use case không gọi CommitMessage của Kafka client.

## 11. Middleware boundary

Phù hợp:

- request ID;
- auth parsing;
- access log;
- panic recovery;
- CORS/compression;
- transport rate limit;
- timeout/body limit.

Không phù hợp:

- overdraft rule;
- loan approval;
- account ownership query ẩn lặp cho mọi route;
- transaction business workflow.

Middleware order là behavior và cần test, nhất là recovery/telemetry/auth.

## 12. Wrong và correct

Wrong:

~~~go
func (h *Handler) Transfer(c *gin.Context) {
	h.db.Begin()
	// domain rule + SQL + Kafka
}
~~~

Correct:

~~~go
func (h *Handler) Transfer(w http.ResponseWriter, r *http.Request) {
	var req transferRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeProblem(w, invalidRequest(err))
		return
	}
	err := h.transfer.Execute(r.Context(), mapCommand(req))
	writeOutcome(w, err)
}
~~~

Handler vẫn có logic, nhưng logic đó là protocol orchestration.

## 13. Testing strategy

- HTTP: httptest, strict body/content-type/error response.
- gRPC: bufconn/in-process server, status code và protobuf mapping.
- Kafka: unit test message mapper/outcome policy; integration test broker cho offset/order.
- CLI: inject Reader/Writer, kiểm exit code/output.
- Contract: cùng use-case fake để chứng minh adapters map cùng intent.

Không cần mock domain object trong adapter test; fake use-case ghi lại command thường đủ.

## 14. Production scenario

5.000 TPS, payload lớn, client disconnect:

1. body limit chặn memory amplification;
2. server timeout tạo request budget;
3. request context cancel query;
4. use case không start nếu decode invalid;
5. unknown error response không lộ DSN;
6. metrics phân biệt 4xx client và 5xx dependency;
7. idempotency bảo vệ retry sau ambiguous response.

Delivery không giải quyết balance consistency, nhưng phải truyền tín hiệu và không phá guarantee application.

## 15. Debug/investigation

Khi HTTP và gRPC cho kết quả khác:

1. capture sanitized transport input;
2. so command sau mapping;
3. so actor/deadline;
4. gọi use case trực tiếp với cùng command;
5. so error mapping;
6. kiểm version/default của schema;
7. thêm cross-adapter contract test.

## 16. Khi nào handler có thể đơn giản hơn?

Tiny CRUD có thể handler → store nếu không có reusable business workflow. Không cần tạo UseCase cho endpoint health/static. Nhưng khi logic được dùng từ HTTP và Kafka hoặc cần transaction/business invariant, boundary trả chi phí rõ.

## 17. Bài tập

1. Thêm gRPC adapter cho Transfer mà không sửa domain.
2. Viết strict JSON decoder reject unknown/trailing body.
3. Thiết kế Kafka ack policy cho retryable/permanent error.
4. So sánh command nhận primitive và Value Object.
5. Viết contract test HTTP/gRPC map cùng application error.

## 18. Mastery questions

1. “Thin handler” vẫn sở hữu policy nào?
2. Tại sao protobuf message không phải Entity?
3. Authorization nằm delivery hay application trong từng case?
4. Kafka consumer là delivery hay infrastructure?
5. context nên và không nên mang gì?
6. Khi nào reuse DTO/result hợp lý?
7. Vì sao unknown error không được trả thẳng?
8. Entry point mới chứng minh boundary tốt ra sao?

## Further reading

- Alistair Cockburn, Ports and Adapters.
- Go net/http và context documentation.
- gRPC status/error model.
- Kafka consumer delivery semantics.

## Quality gate

- [x] Problem, ba level và dependency flow
- [x] DTO/command/entity, validation/auth/context/errors
- [x] Multiple driving adapters và ack semantics
- [x] Wrong/correct Go examples
- [x] Tests, production failure, debugging
- [x] Trade-off, exercises, mastery
