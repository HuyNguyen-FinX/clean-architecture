# 19 Logging Observability

## Tại sao cần học?

Production system cần log, metrics và tracing để hiểu chuyện gì đang xảy ra. Nhưng observability framework không nên xâm nhập domain.

## Logging

Structured logging nên có request ID, actor, use case, aggregate ID và error cause phù hợp. Tránh log duplicate cùng một lỗi ở nhiều layer.

## Metrics

Metrics nên trả lời:

- Latency từng endpoint/use case.
- Error rate.
- DB query duration.
- Kafka lag.
- Retry count.

## Tracing

Trace có thể xuyên HTTP, use case, PostgreSQL và Kafka. Domain method không cần nhận tracer.

## Anti-pattern

- Domain import OpenTelemetry.
- Log PII hoặc secret.
- Log mọi layer cùng một error làm nhiễu incident investigation.

## Mastery Check

- [ ] Tôi biết observability thuộc adapter/application boundary.
- [ ] Tôi biết tránh làm domain phụ thuộc tracing framework.
- [ ] Tôi biết correlation ID/request ID đi qua context.
