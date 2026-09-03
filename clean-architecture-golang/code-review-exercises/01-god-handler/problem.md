# Code Review 01: Transfer Service Tưởng Như Sạch

## Bối cảnh

Team vừa refactor handler: đã có Service, Repository interface, constructor, context và transaction. Production đôi lúc balance lệch, Kafka thiếu/duplicate và client thấy lỗi SQL.

Đọc [starter/transfer.go](./starter/transfer.go). Code compile:

~~~bash
cd starter
go test ./...
go vet ./...
~~~

## Nhiệm vụ

Viết review theo format:

~~~text
[Severity] Title
Failure scenario:
Why:
Minimal safe change:
Test proving the change:
~~~

Tìm ít nhất:

- một lỗi transaction dù code có BeginTx;
- một invariant bị bypass;
- một lỗi context/lifecycle;
- một lỗi async delivery;
- một transport semantic leak;
- một error leak;
- một idempotency/concurrency gap;
- một misleading interface ownership issue.

## Constraints

- Không rewrite toàn service trong một PR.
- Giữ external HTTP request/status trước.
- PostgreSQL là source of truth.
- Kafka delivery cần at-least-once, duplicate chấp nhận nếu consumer idempotent.

## Câu hỏi

1. Ba findings nào phải sửa trước?
2. Characterization/integration tests nào thêm trước?
3. Tách Domain behavior ở commit nào?
4. EventPublisher port có đủ giải dual write không?
5. Transaction handle nên đi qua context, explicit repositories hay UoW?

Chỉ mở [solution](./solution.md) sau khi đã tự viết review.
