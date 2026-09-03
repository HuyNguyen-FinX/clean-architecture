# Solution

CachedReader là decorator; ExpiringCache là adapter dùng Clock được inject. Cache error fail-open sang source vì đây chỉ là performance cache.

~~~bash
go test -race ./...
go vet ./...
~~~
