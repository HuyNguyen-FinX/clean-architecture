# 12 HTTP REST API

## Tại sao cần học?

HTTP là delivery detail phổ biến nhất. Handler tốt giúp API contract rõ ràng mà không làm business logic phụ thuộc transport.

## Flow

```text
HTTP Request
↓
Request DTO
↓
Use Case Command
↓
Use Case
↓
Response DTO
↓
HTTP Response
```

## Go Implementation

Ưu tiên bắt đầu với `net/http` hoặc router nhẹ như `chi`. Framework nặng chỉ nên dùng khi lợi ích routing, middleware, ecosystem lớn hơn coupling thêm vào.

## Error Mapping

Domain/Application error được map ở HTTP layer:

```text
ErrNotFound -> 404
ValidationError -> 400
Conflict -> 409
Unknown -> 500
```

Domain không biết status code.

## Mastery Check

- [ ] Tôi biết handler không chứa business rule.
- [ ] Tôi biết request DTO khác use case command.
- [ ] Tôi biết test HTTP bằng `httptest`.
