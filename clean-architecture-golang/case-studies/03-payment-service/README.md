# Case Study 03: Payment Service

Payment Service giúp học idempotency, external API failure và webhook.

Trọng tâm:

- `Payment` aggregate và status transition.
- `PaymentGateway` port.
- Provider SDK trong infrastructure adapter.
- Webhook là delivery adapter.
- Ambiguous result, retry và reconciliation.

Kết luận chính: Clean Architecture giữ provider detail ra ngoài core, nhưng idempotency và reconciliation mới là phần sống còn trong production.
