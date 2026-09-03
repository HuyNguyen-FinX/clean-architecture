# Solution: Payment Service

Provider SDK thuộc infrastructure adapter. Use case chỉ phụ thuộc gateway port:

```go
type PaymentGateway interface {
	Authorize(ctx context.Context, req AuthorizationRequest) (AuthorizationResult, error)
	Capture(ctx context.Context, req CaptureRequest) (CaptureResult, error)
}
```

Webhook là delivery adapter. Nó map provider payload sang application command như `HandlePaymentWebhookCommand`.

Idempotency không chỉ là database unique key. Cần lưu request key, operation, provider reference và trạng thái xử lý. Với error ambiguous, cần reconciliation hoặc query provider trước khi retry capture.
