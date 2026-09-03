# Solution tham khảo: Notification Service

## Complexity assessment

If event → render → send, domain is workflow/data-centric. Start:

~~~text
consumer adapter → HandleNotification use case → Sender port
                                      → Template renderer
                                      → Delivery store/inbox
~~~

No need Aggregate hierarchy.

## Ports

~~~go
type Sender interface {
	Send(context.Context, Message, IdempotencyKey) (DeliveryResult, error)
}

type TemplateRenderer interface {
	Render(TemplateID, Locale, Data) (Subject, Body, error)
}
~~~

Separate Email/SMS/Push adapters may implement channel-specific ports if capabilities differ; one lowest-common Sender can hide SMS/email semantics.

## Idempotency

EventID + notification type + recipient/channel. Store delivery state and provider reference. Provider timeout can be Unknown; inquire/reconcile if supported.

## Retry/DLQ

Transient 429/5xx/network bounded retry; invalid address/template/schema permanent; DLQ owner/replay. Respect provider rate/quota and Retry-After. Do not block partition indefinitely if retry topic policy chosen.

## Template ownership

Simple rendering is application/infrastructure. Regulatory contact preferences, quiet hours, locale fallback or channel eligibility can become domain policy.

## Security

PII minimization/encryption/redaction; do not log body/token. Unsubscribe/preferences and audit.

## Tests

- event mapper/duplicate;
- template fixtures;
- sender adapter httptest;
- retry/unknown;
- DLQ metadata/replay;
- rate limiter;
- end-to-end delivery state.

## Alternative

Managed notification provider may replace most service. Keep thin ACL/webhook/idempotency if vendor lock/failure semantics matter. Full Clean layers may not.
