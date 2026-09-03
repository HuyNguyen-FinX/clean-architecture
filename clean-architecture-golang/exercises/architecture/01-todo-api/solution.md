# Solution tham khảo: Todo API

## Assumptions

Team 3, internal API, PostgreSQL, dưới 100 RPS. V1 chủ yếu CRUD; MarkCompleted có rule Archived không thể complete và optimistic edit cần chống lost update.

## Thiết kế tối thiểu

~~~text
cmd/api
internal/todo/
  handler.go
  service.go
  postgres.go
  model.go
~~~

Không cần bốn layer/package con. Package todo vẫn giữ fields private nếu behavior xuất hiện.

~~~go
type Todo struct {
	id      ID
	title   string
	status  Status
	version uint64
}

func (t *Todo) MarkCompleted(now time.Time) error {
	if t.status == Archived {
		return ErrArchived
	}
	t.status = Completed
	return nil
}
~~~

Handler map JSON; Service orchestration/auth; Postgres mapping/query. Một Store interface cạnh Service chỉ cần khi unit test/adapter variation đáng giá:

~~~go
type Store interface {
	Find(context.Context, ID) (*Todo, error)
	Save(context.Context, *Todo, uint64) error
}
~~~

## Concurrency

PATCH nhận If-Match/version. UPDATE ... WHERE id=? AND version=?; zero row → conflict. HTTP 409/412 tùy contract.

## Models

V1 có thể dùng một read DTO cho simple query. Không dùng request DTO làm writable Entity vì client không được set status/version tùy ý.

## Tests

- MarkCompleted domain cases;
- Service auth/conflict fake;
- httptest mapping;
- PostgreSQL optimistic integration;
- migration.

## Evolution triggers

Recurring, assignment permissions, reminders và collaboration làm domain/application split đáng giá. Nếu vẫn CRUD, giữ phẳng.

## Alternative

Handler gọi generated SQL queries trực tiếp có thể hợp cho pure CRUD. Chỉ extract behavior route MarkCompleted. Clean Architecture không yêu cầu uniform layers cho mọi endpoint.

## Failure

Hai clients edit title: version prevents last-write silently. Reminder sending async cần outbox nếu guarantee quan trọng.
