# Solution: Banking Transfer

Domain:

```text
Account.Withdraw
Account.Deposit
Money
Transfer status
```

Application:

```text
TransferMoneyUseCase
  validate idempotency
  begin transaction
  load accounts for update
  call Withdraw/Deposit
  save accounts
  save transfer record
  write outbox event
  commit
```

PostgreSQL adapter xử lý row locking bằng SQL detail như `FOR UPDATE`. Domain không biết lock. Idempotency thường là application/infrastructure phối hợp: use case quyết định policy, database enforce unique key.

Clean Architecture giúp giữ business rule độc lập và testable. Nó không tự giải quyết race condition, retry safety hoặc message delivery; bạn vẫn cần transaction, lock, idempotency và outbox.
