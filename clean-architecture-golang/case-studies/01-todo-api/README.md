# Case Study 01: Todo API - Kiến Trúc Tăng Trưởng Theo Độ Phức Tạp

Todo API là phép thử tốt cho sự thực dụng: domain ban đầu rất nhỏ, vì vậy một kiến trúc nhiều layer có thể đắt hơn vấn đề nó giải quyết. Case study này không hỏi "Clean Architecture có dùng được không?" mà hỏi "boundary nào đã có giá trị ở thời điểm nào?".

## Bối Cảnh Và Yêu Cầu

Phiên bản đầu có năm hành vi: tạo, xem, sửa, xóa và hoàn thành Todo. `Title` bắt buộc, `DueDate` có thể rỗng. Sau ba tháng, sản phẩm thêm:

- Todo hoàn thành không được sửa nội dung; chỉ được reopen bởi owner.
- Due date không được lùi về quá khứ, trừ thao tác import của admin.
- Client mobile gửi lại request khi mất mạng nên create phải idempotent.
- Hai thiết bị có thể cùng sửa một Todo.
- Khi Todo sắp quá hạn, hệ thống phát event cho Notification Service.

Điểm đáng chú ý là kiến trúc hợp lý ở ngày đầu có thể không còn hợp lý ở tháng thứ ba.

## Quyết Định Phiên Bản A: CRUD Nhỏ

Với một service nội bộ, một database, ba người dùng và không có rule ngoài `title != empty`, cấu trúc sau là đủ:

~~~text
todo/
  handler.go
  store.go
  model.go
~~~

`handler` parse HTTP và gọi một concrete `Store`. `Store` giữ SQL. `model.go` có struct dữ liệu chung nếu API và schema thật sự cùng tiến hóa. Đây không phải thất bại kiến trúc; nó là quyết định mua ít flexibility vì chi phí thay đổi đang thấp.

Không nên tạo `TodoEntity`, `TodoDTO`, `TodoRecord`, `TodoRepository`, `TodoRepositoryImpl`, presenter và factory nếu mỗi lớp chỉ copy cùng bốn field. Số abstraction không đo chất lượng architecture.

## Điểm Bẻ Gãy

Khi `MarkCompleted`, `Reopen` và optimistic concurrency xuất hiện, handler bắt đầu phải biết state transition. Nếu cả HTTP import lẫn cron worker đều tự viết rule, hành vi sẽ lệch nhau. Đây là lúc tách core nhỏ đem lại giá trị.

~~~go
type Todo struct {
	id      string
	title   string
	status  Status
	dueDate *time.Time
	version int64
}

func (t *Todo) Complete(now time.Time) error {
	if t.status == Completed {
		return ErrAlreadyCompleted
	}
	t.status = Completed
	return nil
}
~~~

Field private làm mọi state transition phải đi qua behavior. `now` được đưa vào như value thay vì gọi `time.Now()` để test deterministic. Domain vẫn không cần `context.Context`, HTTP status hay SQL tag.

## Kiến Trúc Phiên Bản B

~~~mermaid
flowchart LR
    HTTP["HTTP handler"] --> APP["Todo use cases"]
    CRON["Deadline worker"] --> APP
    APP --> DOMAIN["Todo"]
    APP --> PORT["TodoStore port"]
    PG["PostgreSQL adapter"] -.implements.-> PORT
~~~

Compile-time dependency:

~~~text
http -> application -> domain
worker -> application -> domain
postgres -> application + domain
~~~

Runtime của `CompleteTodo` vẫn đi từ use case tới PostgreSQL, nhưng application chỉ gọi interface do nó sở hữu. Thay database không buộc rule `Complete` đổi.

~~~go
type Store interface {
	Find(ctx context.Context, id string) (*domain.Todo, error)
	Save(ctx context.Context, todo *domain.Todo, expectedVersion int64) error
}

type CompleteTodo struct{ store Store }

func (uc CompleteTodo) Execute(ctx context.Context, id string) error {
	todo, err := uc.store.Find(ctx, id)
	if err != nil {
		return err
	}
	wantVersion := todo.Version()
	if err := todo.Complete(time.Now()); err != nil {
		return err
	}
	return uc.store.Save(ctx, todo, wantVersion)
}
~~~

Đoạn trên rút gọn để trình bày boundary. Production code nên inject `Clock`, chuẩn hóa ID và wrap error có operation context.

## Concurrency Và HTTP Contract

Database có cột `version`. Update dùng compare-and-swap:

~~~sql
UPDATE todos
SET status = $1, version = version + 1
WHERE id = $2 AND version = $3;
~~~

Nếu `RowsAffected == 0`, adapter trả `ErrConflict`; HTTP map thành `409 Conflict` hoặc `412 Precondition Failed` nếu contract dùng `If-Match`. Domain không biết status code. Client phải reload thay vì âm thầm ghi đè thay đổi từ thiết bị khác.

HTTP handler chỉ chịu trách nhiệm:

1. Parse path/body/header và giới hạn body.
2. Validate shape như field thiếu hoặc timestamp sai format.
3. Map sang command.
4. Gọi use case.
5. Map lỗi ổn định sang response, không lộ SQL.

## Idempotency Và Event

`CreateTodo` nhận `Idempotency-Key`. Trong cùng transaction, application ghi Todo và `(actor_id, key, request_hash, response)`. Cùng key/cùng hash trả lại kết quả cũ; cùng key/khác hash là conflict. Khi tạo reminder event, transaction đồng thời insert outbox row. Không publish broker trực tiếp rồi hy vọng DB commit.

Nếu hệ thống vẫn chỉ có vài request/ngày, yêu cầu này có thể được bỏ. Idempotency và outbox là response cho failure model cụ thể, không phải huy hiệu production-ready.

## Failure Walkthrough

| Failure | Kết quả mong muốn | Boundary xử lý |
|---|---|---|
| JSON sai | `400`, không gọi use case | HTTP adapter |
| Todo không tồn tại | lỗi typed/not-found | Store -> application -> HTTP mapping |
| Hai client cùng update | một thành công, một conflict | PostgreSQL adapter + API contract |
| Commit xong nhưng response mất | retry trả cùng kết quả | idempotency record |
| Broker down | Todo vẫn commit, event còn pending | outbox worker |
| Worker nhận event hai lần | notification deduplicate | consumer đích |

## Testing Strategy

- Domain table test: complete, complete lần hai, reopen đúng/sai actor, due date boundary.
- Use-case test với fake store: not-found, save conflict, dependency error, không save khi domain từ chối.
- Repository integration test trên PostgreSQL thật: mapping, unique key, version conflict, rollback.
- HTTP contract test bằng `httptest`: malformed JSON, unknown field, body dư, error mapping.
- Outbox integration test: commit cùng Todo, rollback cùng Todo, publish retry.
- Race test không thay thế test database concurrency; chạy cả `go test -race` và hai transaction thật.

## Quan Sát Production

Log `operation`, `todo_id`, `request_id`, `error_kind`; không log toàn bộ title nếu chứa dữ liệu nhạy cảm. Metrics cần request latency/error rate, optimistic-conflict rate và outbox age. Conflict tăng đột biến có thể là UX retry sai, không nhất thiết database lỗi.

## Trade-off Và Quy Tắc Tiến Hóa

- Tách domain khi behavior phải được dùng nhất quán từ nhiều adapter.
- Tách DTO khi API contract và persistence thay đổi vì lý do khác nhau.
- Tạo port khi application cần bảo vệ khỏi một detail hoặc cần test double hữu ích.
- Giữ concrete dependency nếu module nhỏ và thay đổi cùng nhau.
- Đừng đoán trước mười provider; refactor khi có evidence từ rule, ownership hoặc volatility.

## Câu Hỏi Mastery

1. Điều kiện nào khiến một Todo từ record CRUD trở thành Entity có behavior?
2. `version` là domain concept, persistence detail hay cả hai trong context này?
3. Vì sao `If-Match` không nên đi xuyên vào method `Todo.Complete`?
4. Nếu chỉ có create/get và SQLite nhúng, bạn sẽ bỏ layer nào?
5. Outbox bảo đảm điều gì và không bảo đảm điều gì?

## Bài Thực Hành

Thiết kế hai pull request: PR đầu cho Version A tối giản; PR sau thêm `CompleteTodo` và optimistic concurrency mà không rewrite toàn service. Ghi Architecture Decision Record nêu trigger khiến bạn thêm từng boundary.
