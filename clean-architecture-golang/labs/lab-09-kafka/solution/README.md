# Solution

~~~text
consumer/  strict envelope mapper, inbox/use-case ports, error classification
memory/    ProcessOnce teaching adapter
outbox/    repository + publisher worker
~~~

~~~bash
go test -race ./...
go vet ./...
~~~

Không có Kafka client trong core exercise: client callback chỉ cần gọi Consumer.Handle rồi map Permanent/Retryable sang DLQ/retry/offset policy.
