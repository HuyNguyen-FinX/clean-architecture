# Solution

~~~text
domain/                 Aggregate và invariant
application/            use case + consumer-owned port
infrastructure/memory/  adapter
repositorytest/         reusable contract suite
~~~

Chạy:

~~~bash
go test -race ./...
go vet ./...
~~~

Contract suite không thay thế PostgreSQL integration test. Nó chỉ buộc các adapter chia sẻ những semantics mà application đã công bố.
