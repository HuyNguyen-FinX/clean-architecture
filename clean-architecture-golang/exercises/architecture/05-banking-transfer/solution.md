# Solution tham khảo: Banking Transfer

## Domain

Money Value Object (minor units + Currency), Account Entity/Aggregate with private balance/status/overdraft, Transfer with immutable identity/status/history.

## Use case

~~~text
validate command/idempotency key
BEGIN
claim key + compare request hash
lock A/B in stable ID order
Withdraw/Deposit
save Accounts
insert Transfer
complete idempotency response
insert outbox MoneyTransferred.v1
COMMIT
~~~

Application owns operation boundary. PostgreSQL adapter owns SELECT FOR UPDATE/SQLSTATE/mapping. Domain owns amount/currency/overdraft/frozen invariants.

## Ports

Small AccountRepository + TransferRepository + Idempotency, coordinated by Transactor, or one explicit UnitOfWork callback. Both valid; compare hidden context vs interface ceremony.

## Concurrency

Pessimistic row lock or optimistic version. Stable order reduces opposite-transfer deadlock. Retry only classified abort with reload/jitter/budget. Mutex not cross-replica.

## Idempotency

Tenant+key unique; canonical hash; same key/same body replay TransferID/result; same key/different body conflict. Atomic with money effect. Status endpoint handles lost response.

## History/ledger

Simple Transfer table gives history; financial system should consider double-entry ledger debit+credit sum zero, balance projection and reconciliation.

## Event

Outbox intent same transaction; worker at-least-once; consumer inbox idempotent. No network in DB transaction.

## Tests

- Money overflow/currency;
- Account boundary/frozen;
- use-case fake;
- PostgreSQL rollback/lock/concurrent;
- duplicate HTTP key;
- outbox Kafka failure/duplicate;
- reconciliation/property debit+credit.

## Operations

Lock/pool latency, transfer outcomes, duplicate/conflict, outbox age, ledger imbalance. Hot account may require append ledger/asynchronous projection, changing consistency.

## Clean Architecture limits

It isolates policy/adapter and makes tests feasible. It does not choose isolation, prevent double spend, provide exactly-once or reconcile money.
