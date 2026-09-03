# Solution

~~~text
domain/       business behavior
application/  command + UoW contract
memory/       transactional adapter + outbox
httpapi/      preserved HTTP contract
~~~

~~~bash
go test -race ./...
go vet ./...
~~~
