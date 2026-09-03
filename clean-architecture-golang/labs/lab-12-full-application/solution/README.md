# Solution

~~~text
domain/       Account invariant
application/  TransferMoney, UnitOfWork, history contracts
memory/       copy-on-write transaction + read projection
httpapi/      strict-ish transport mapping
cmd/api/      composition root
~~~

~~~bash
go test -race ./...
go vet ./...
go run ./cmd/api
~~~

Outbox ở solution là durable intent trong transaction profile, chưa phải delivered event. Publisher worker/inbox xem Lab 09.
