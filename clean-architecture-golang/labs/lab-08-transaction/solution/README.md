# Solution

Memory Store implement cả AccountRepository và Transactor. Transaction giữ working copy trong private context value rồi commit atomically khi callback thành công.

~~~bash
go test -race ./...
go vet ./...
~~~

Adapter này dành cho học/test, không thay integration test PostgreSQL.
