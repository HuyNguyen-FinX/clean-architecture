# Solution

handler.go chứa protocol policy; Transfer interface đại diện use case consumer. Test xác minh strict parsing, mapping, safe error và context propagation.

~~~bash
go test -race ./...
go vet ./...
~~~
