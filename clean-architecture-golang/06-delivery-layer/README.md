# 06 Delivery Layer

## Tại sao cần học?

Delivery Layer là nơi hệ thống nhận input từ bên ngoài: HTTP, gRPC, CLI, Kafka consumer, cron job. Nó chuyển đổi protocol-specific input thành command/query cho application.

## Vấn đề

Nếu business rule nằm trong handler, bạn phải copy rule khi thêm gRPC hoặc worker. Delivery adapter nên mỏng nhưng không vô dụng: nó chịu trách nhiệm protocol, DTO, auth extraction, request validation cơ bản và error mapping.

## Flow

```text
HTTP Request
↓
Handler
↓
Request DTO
↓
Use Case Command
↓
Application
↓
Response DTO
↓
HTTP Response
```

## Dependency

Delivery import application. Application không import delivery.

## Anti-pattern

- Handler gọi SQL trực tiếp.
- Handler chứa business invariant.
- Domain error chứa HTTP status code.
- Dùng framework context như input của use case.

## Mastery Check

- [ ] Tôi biết handler nên map protocol sang use case.
- [ ] Tôi biết thêm delivery adapter mới không nên sửa domain.
- [ ] Tôi biết lỗi domain/application được map sang transport ở delivery.
