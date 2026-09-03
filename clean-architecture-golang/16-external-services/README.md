# 16 External Services

## Tại sao cần học?

External API là detail không ổn định: timeout, retry, partial failure, version change. Clean Architecture đặt chúng sau gateway port để application nói bằng intent, không nói bằng SDK cụ thể.

## Ví dụ

```text
Payment Use Case
↓
PaymentGateway interface
↑
StripeAdapter
```

Hoặc:

```text
Loan Use Case
↓
CoreBankingGateway
↑
HTTPCoreBankingClient
```

## Nội dung trọng tâm

- Timeout và cancellation.
- Retry/backoff.
- Circuit breaker.
- Idempotency key.
- Mapping external error sang application error.

## Anti-pattern

- Use case import Stripe SDK.
- Domain dùng response model của external API.
- Retry không giới hạn.
- Không phân biệt timeout, rejected, unknown result.

## Mastery Check

- [ ] Tôi biết gateway port biểu diễn capability application cần.
- [ ] Tôi biết external API response không phải domain model.
- [ ] Tôi biết network failure cần policy riêng.
