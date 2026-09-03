# Starter

Handler cố ý bỏ decode error, mất request context và leak internal error. Baseline test chỉ cover happy path.

~~~bash
go test ./...
~~~
