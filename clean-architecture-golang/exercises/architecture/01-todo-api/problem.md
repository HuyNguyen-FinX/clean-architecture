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
