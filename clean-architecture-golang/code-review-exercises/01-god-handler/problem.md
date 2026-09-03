# Code Review 01: God Handler

## Code xấu

```go
func Transfer(w http.ResponseWriter, r *http.Request) {
	// parse HTTP
	// query DB
	// validate balance
	// update account
	// call Kafka
	// return JSON
}
```

## Nhiệm vụ

1. Liệt kê vấn đề.
2. Xác định business logic.
3. Xác định infrastructure detail.
4. Thiết kế boundary.
5. Refactor theo từng bước nhỏ.

## Gợi ý

Đừng bắt đầu bằng cách tạo folder. Bắt đầu bằng cách tìm rule:

```text
Khi nào transfer hợp lệ?
Balance thay đổi theo invariant nào?
Transaction cần bao phủ những thao tác nào?
Kafka publish có cần outbox không?
```
