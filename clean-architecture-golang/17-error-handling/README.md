# 17 Error Handling

## Tại sao cần học?

Error trong Clean Architecture vừa là kỹ thuật Go vừa là boundary language. Lỗi domain không nên biết HTTP status, lỗi infrastructure không nên rò driver detail lên API response.

## Phân loại

- Domain Error: vi phạm invariant hoặc rule.
- Application Error: workflow invalid, conflict, authorization ở use case.
- Infrastructure Error: database timeout, network failure.
- Transport Error: JSON invalid, route not found, gRPC status.

## Go Implementation

Sử dụng `errors.Is`, `errors.As`, wrapping có chủ đích. Error message log nội bộ có thể giàu detail hơn response public.

## Mapping

```text
ErrInsufficientBalance -> 409
ErrAccountNotFound -> 404
Invalid JSON -> 400
DB timeout -> 503 hoặc 500 tùy policy
```

## Anti-pattern

- Domain trả HTTP status.
- Adapter trả raw SQL error ra client.
- Mọi lỗi bị convert thành string.
- Log cùng lỗi ở mọi layer.

## Mastery Check

- [ ] Tôi biết lỗi thuộc layer nào.
- [ ] Tôi biết map error ở delivery.
- [ ] Tôi biết giữ nguyên cause cho observability.
