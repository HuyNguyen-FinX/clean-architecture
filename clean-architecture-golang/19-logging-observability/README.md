# Logging và Observability: nhìn xuyên boundary mà không nhiễm Domain

Observability không phải “thêm logger”. Hệ thống cần trả lời request nào chậm, use case nào fail, query nào chờ lock, outbox nào backlog và error cause ở đâu, trong khi không log bí mật hoặc biến Domain thành client OpenTelemetry.

## Kết quả học tập

- phân vai logs, metrics, traces;
- instrument HTTP/Application/Infrastructure boundaries;
- dùng slog với structured fields;
- quản correlation, cardinality, sampling và PII;
- thiết kế SLI/SLO/alerts;
- điều tra incident theo trace/log/metric.

## 1. Problem

~~~go
func (a *Account) Withdraw(ctx context.Context, amount Money) error {
	span := trace.SpanFromContext(ctx)
	logger.Info("withdrawing", "balance", a.balance)
	metrics.WithdrawCounter.Inc()
	// rule
}
~~~

Domain rule phụ thuộc request lifecycle và telemetry vendors; test phải dựng context; mỗi gọi log có thể lộ balance. Instrumentation concern đã xâm nhập model.

## 2. Ba level

### Level 1

Logs kể sự kiện chi tiết; metrics cho xu hướng/alert; traces nối latency qua các bước.

### Level 2

Engineer đặt fields, labels, spans, propagation, sampling, retention và dashboards.

### Level 3

Observability là feedback architecture: boundary phải phát ra signal đủ để kiểm chứng runtime graph và guarantee. Instrumentation đặt quanh I/O/use case, không nằm trong business computation thuần.

## 3. End-to-end flow

~~~mermaid
flowchart LR
    HTTP["HTTP span"] --> APP["Use-case span"]
    APP --> PG["PostgreSQL span"]
    APP --> OUTBOX["Outbox write"]
    WORKER["Outbox worker span"] --> KAFKA["Kafka publish span"]
~~~

Trace context qua context.Context và message headers. Business correlation như TransferID vẫn là domain/application identity, không thay bằng trace ID.

## 4. Structured logging với slog

~~~go
logger.InfoContext(ctx, "transfer completed",
	"operation", "transfer_money",
	"transfer_id", result.ID,
	"duration_ms", duration.Milliseconds(),
)
~~~

Error boundary:

~~~go
logger.ErrorContext(ctx, "request failed",
	"operation", "transfer_money",
	"error", err,
	"trace_id", traceID(ctx),
)
~~~

Không format một JSON string thủ công. Handler/logger config ở composition root; domain không nhận slog.Logger.

## 5. Log một lần, wrap nhiều nơi

Repository wrap cause; application wrap workflow context; outer boundary log một lần. Inner layer chỉ log nếu nó recover/continue hoặc phát sinh một operational event riêng.

Duplicate logs làm incident count sai và noise. Cần preserve errors.Is/As khi wrap.

## 6. Field policy và PII

Cho phép:

- operation, route template, status category;
- trace/request/event IDs;
- hashed/redacted account ID;
- dependency, attempt, duration;
- error category.

Cẩn thận:

- balance/amount theo compliance;
- user/account identifiers;
- IP/user agent;
- raw Kafka payload/body;
- SQL args;
- token, DSN, password tuyệt đối không.

Data classification/retention/access là architecture concern.

## 7. Metrics

Golden signals:

- request rate;
- error rate;
- latency distribution;
- saturation.

Ví dụ:

~~~text
http_server_requests_total{route,method,status_class}
http_server_duration_seconds{route,method}
usecase_duration_seconds{operation,outcome}
db_pool_acquire_seconds
db_transaction_retries_total{reason}
outbox_oldest_pending_seconds
kafka_consumer_lag
~~~

Dùng histogram cho latency. Counter monotonic. Gauge cho current backlog/connections.

## 8. Cardinality

Không dùng account_id, request_id, raw URL hoặc error message làm metric label. Hàng triệu series làm backend metrics tốn bộ nhớ/khó query.

Route template /accounts/{id}, không raw /accounts/A-100. High-cardinality ID để ở trace/log.

## 9. SLI, SLO và alert

SLI transfer có thể là tỷ lệ non-user-error hoàn tất trong 500ms. Business rejection insufficient balance không nhất thiết tính availability failure.

SLO ví dụ 99.9% trong 30 ngày. Alert theo burn rate tốt hơn “có một 500”. Outbox age/ledger reconciliation có SLI riêng vì HTTP success chưa nghĩa event delivered.

## 10. Tracing

Span boundaries:

- HTTP/gRPC/Kafka ingress;
- use case;
- DB query/transaction;
- Redis;
- external call;
- Kafka publish/consume.

Không span mỗi getter/domain method. Quá nhiều spans tăng cost và che critical path.

Attributes cần stable semantic names; không record secrets. Record exception/error status theo SDK conventions.

## 11. Propagation

HTTP trace headers được middleware extract/inject. Kafka producer đưa trace context vào headers; consumer tạo/link span.

Async event có thể dùng parent-child hoặc span link tùy thời gian/lifecycle. Không giữ in-memory context đến worker sau restart; serialize standard propagation headers, business IDs.

## 12. Sampling

Head sampling quyết định sớm, rẻ nhưng có thể bỏ rare error. Tail sampling giữ trace lỗi/chậm nhưng backend phức tạp. Có thể:

- baseline small percentage;
- keep errors/high latency;
- deterministic by trace ID;
- tăng tạm trong incident.

Logs/metrics/traces có cost budget; không “100% mãi mãi” theo phản xạ.

## 13. Instrumentation qua decorator/middleware

~~~go
type observedTransfer struct {
	next   Transfer
	clock  Clock
	metric TransferMetrics
}

func (o observedTransfer) Execute(ctx context.Context, cmd Command) (err error) {
	start := o.clock.Now()
	defer func() {
		o.metric.Observe(o.clock.Now().Sub(start), classify(err))
	}()
	return o.next.Execute(ctx, cmd)
}
~~~

Decorator không cần đổi core use case. Repository adapter đo query/lock/pool details vì nó biết operation hạ tầng.

## 14. Domain signals vẫn quan sát được

Domain trả typed error hoặc Domain Event. Application boundary classify outcome. Không cần Domain gọi metrics để đếm insufficient balance.

~~~text
Account.Withdraw → ErrInsufficientBalance
Application decorator → outcome=business_rejection
HTTP mapper → 409
~~~

## 15. Health vs telemetry

Liveness chỉ nói process sống. Readiness nói có nhận traffic đúng guarantee. Metrics dashboard không thay health probe; health probe không thay alert.

Không ping tất cả dependencies mỗi request health. Cache down có thể degraded; DB down có thể unready; Kafka down với outbox có thể vẫn accept nhưng backlog alert.

## 16. Error observability

Giữ wrapped cause và category. Stack trace thường ở panic/error boundary, không thêm stack vào mọi wrap. Aggregate/sampling repeated error để tránh log storm.

## 17. Production scenario

Transfer p99 tăng:

1. metric cho thấy usecase latency, HTTP status vẫn 201;
2. trace chậm ở pgx acquire hay SELECT FOR UPDATE;
3. DB metrics chỉ ra hot account lock wait;
4. log có hashed account/transfer ID;
5. outbox age bình thường, loại Kafka;
6. mitigation load shed/partition/workflow.

Không có boundary spans, team chỉ thấy “API chậm”.

## 18. Incident investigation

Từ alert:

1. xác nhận SLI/burn rate và blast radius;
2. split route/operation/outcome;
3. mở exemplar trace chậm/lỗi;
4. theo dependency spans;
5. query logs bằng trace/business ID;
6. kiểm recent deploy/config/schema;
7. mitigate;
8. ghi timeline và bổ sung missing signal.

Observability phải hỗ trợ câu hỏi, không phải dashboard trang trí.

## 19. Testing

- slog handler capture kiểm field/redaction;
- metric recorder fake kiểm classification, không khóa implementation sequence;
- trace exporter in-memory kiểm parent relation quan trọng;
- middleware test request ID/context;
- load/cardinality test;
- chaos dependency để xác minh alerts.

Không assert timestamp/serialized log toàn dòng nếu không phải contract.

## 20. Khi nào instrumentation đơn giản đủ?

Service nhỏ có thể bắt đầu slog JSON + HTTP metrics + dependency metrics. Không cần inject custom Logger interface vào mọi package hoặc tự viết tracing SDK. Dùng OpenTelemetry ở adapters khi distributed path cần.

## 21. Bài tập

1. Thêm slog request middleware không log body.
2. Thiết kế metrics labels tránh account ID.
3. Tạo trace HTTP → use case → DB.
4. Viết SLI cho outbox delivery.
5. Điều tra giả lập pool exhaustion.

## 22. Mastery questions

1. Logs/metrics/traces trả lời câu hỏi khác nhau nào?
2. Vì sao account ID không là metric label?
3. Domain Error giúp metrics mà không import SDK ra sao?
4. HTTP 201 có chứng minh Kafka delivered không?
5. Sampling trade-off?
6. Readiness khi Kafka down + outbox hoạt động?
7. Log ở mọi layer gây failure gì?
8. Business ID khác trace ID?

## Further reading

- Go log/slog documentation.
- OpenTelemetry specification và semantic conventions.
- Prometheus instrumentation/cardinality guidance.
- Google SRE Workbook về SLO và burn-rate alerts.

## Quality gate

- [x] Logs/metrics/traces mental model
- [x] Instrumentation boundaries + Go examples
- [x] Correlation, PII, cardinality, sampling
- [x] SLI/SLO/health/async propagation
- [x] Production investigation/tests
- [x] Trade-off, exercises, mastery
