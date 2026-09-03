# Solution

~~~text
application/  consumer port và GetBalance
memory/       concrete Store
httpapi/      delivery adapter
composition/  config, object graph, cleanup
~~~

~~~bash
go test -race ./...
go vet ./...
~~~
