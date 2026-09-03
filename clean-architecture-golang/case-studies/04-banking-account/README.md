# Case Study 04: Banking Account

Banking Account dùng để học domain invariant và transaction boundary.

Trọng tâm:

- `Money` Value Object.
- `Account` Entity.
- Withdraw/deposit behavior.
- Transfer atomicity.
- Row locking và double spending.
- Transaction history.

Kết luận chính: rule balance thuộc domain, nhưng locking và transaction implementation thuộc adapter/application boundary.
