# Solution: Notification Service

Notification Service có thể là workflow-heavy nhưng domain không phải lúc nào cũng giàu. Nếu chỉ nhận event rồi gửi provider, kiến trúc nhẹ là đủ.

Port:

```go
type Sender interface {
	Send(ctx context.Context, message Message) error
}
```

Provider adapters implement email/SMS/push. Retry, DLQ và idempotency thuộc worker/application/infrastructure boundary. Template có thể là application concern nếu chỉ render nội dung, hoặc domain concern nếu có rule nghiệp vụ phức tạp về communication policy.
