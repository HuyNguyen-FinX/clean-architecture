# Solution

~~~text
domain/       invariant tests
application/  fake/spy use-case tests
memory/       detached ownership adapter
httpapi/      httptest contract
~~~

~~~bash
go test -race ./...
go vet ./...
~~~

Test gap cố ý: không có PostgreSQL nên suite chưa chứng minh SQL, rollback hoặc row locking.
