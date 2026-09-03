# Bài 1: Thiết Kế Todo API

## Requirement

Thiết kế API:

- Create Todo.
- Get Todo.
- Update Todo.
- Delete Todo.
- Mark Todo as completed.

Todo có:

- ID.
- Title.
- Status.
- Due date optional.

## Nhiệm vụ

1. Chọn project structure.
2. Xác định domain rule thật sự.
3. Quyết định có cần tách Entity, DTO, DB model không.
4. Quyết định có cần repository interface không.
5. Chỉ ra điểm nào có nguy cơ over-engineering.

## Câu hỏi

- Đây có phải domain-rich system không?
- Nếu chỉ có CRUD, Clean Architecture đầy đủ có đáng không?
- `MarkCompleted` có phải business behavior không?

## Bối Cảnh Tiến Hóa

Thiết kế V1 cho team ba người, dưới 100 RPS và một PostgreSQL. Sau đó đánh giá lại khi có mobile offline retry, hai thiết bị cùng edit, rule "completed không được đổi due date" và reminder worker.

Không được mặc định mọi requirement của V2 phải có từ ngày đầu. Ghi rõ trigger nào khiến bạn thêm Entity behavior, interface, DTO riêng, optimistic locking hoặc outbox.

## Failure Injection

- Hai PATCH cùng đọc version 7 rồi ghi title khác nhau.
- Database commit create thành công nhưng response bị mất.
- Reminder broker down trong 30 phút.
- Client gửi JSON có unknown field `status=completed`.

Với mỗi failure, ghi expected behavior, owner boundary và dữ liệu cần lưu để recover.

## Deliverables

1. Hai package tree cho V1 và V2, kèm import arrows.
2. API contract cho create/complete/update conflict.
3. Pseudocode hoặc Go signature của behavior và Store.
4. Transaction/idempotency decision; có thể chọn không dùng nhưng phải giải thích.
5. Test matrix theo domain, HTTP và PostgreSQL.
6. ADR tối đa một trang: complexity nào bạn cố ý chưa mua.

## Self-review

- Có lớp nào chỉ copy bốn field mà chưa bảo vệ volatility không?
- Handler/worker có tự viết cùng state rule hai lần không?
- Test concurrency có chạy hai database transaction thật không?
- Error domain có chứa HTTP status không?
