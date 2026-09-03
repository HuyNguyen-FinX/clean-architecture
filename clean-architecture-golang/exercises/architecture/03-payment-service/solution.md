# Solution tham khảo: Payment Service

## Payment Aggregate

~~~text
Created → Authorizing → Authorized → Captured
                    ↘ Declined
Authorizing → PendingInquiry
Captured → Refunding → Refunded
~~~

Invalid transitions rejected; refund là operation mới, không technical rollback.

## Gateway

~~~go
type PaymentGateway interface {
	Authorize(context.Context, AuthorizationRequest) (AuthorizationResult, error)
	Capture(context.Context, CaptureRequest) (CaptureResult, error)
	Inquire(context.Context, ProviderOperationID) (ProviderStatus, error)
}
~~~

Port nói intent, không SDK request. Adapter maps provider errors/status/idempotency header.

## Idempotency

Persist merchant+operation+key, canonical request hash, local PaymentID, provider key/reference và stable response. Same key khác hash → conflict; concurrent duplicate serialized by unique constraint/lock.

## Ambiguous outcome

Timeout sau provider nhận request không phải Declined. Mark PendingInquiry, enqueue inquiry, reconcile. Không retry với key mới.

## Webhook

Driving adapter verifies signature/timestamp, stores provider event inbox, maps command. Duplicate/out-of-order handled by event ID + transition/version. Return 2xx only after event safely recorded/applied per provider retry contract.

## Local transaction

State + idempotency + outbox atomically. Network call ngoài DB lock. Worker/process manager applies result.

## Errors

Declined is business result; ProviderUnavailable retryable; UnknownOutcome state; malformed webhook transport/permanent. HTTP maps stable codes.

## Tests

- state transitions;
- same/different idempotency key;
- httptest provider timeout/status/header;
- webhook signature/duplicate/order;
- Postgres transaction;
- reconciliation fixture;
- sandbox integration.

## Alternative

For synchronous low-risk provider with strong idempotency, Application can call gateway then persist; still handle crash between call/persist via inquiry. Architecture cannot erase remote ambiguity.

## Observability

Local/payment/provider IDs, attempt timeline, pending age, webhook lag, reconciliation mismatch; never log card/token.
