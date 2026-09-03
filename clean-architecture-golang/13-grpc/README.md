# 13 gRPC

## Tại sao cần học?

gRPC chứng minh lợi ích của boundary: thêm transport mới mà không sửa domain/use case.

## Flow

```text
HTTP ───┐
        ├──> Use Case
gRPC ───┘
```

gRPC server là delivery adapter giống HTTP handler. Protobuf message là DTO, không phải domain entity.

## Dependency

```text
delivery/grpc -> application
application -> domain
```

Không để application import generated protobuf nếu protobuf là transport contract. Có thể map protobuf request sang command ở gRPC adapter.

## Anti-pattern

- Dùng protobuf message làm domain entity.
- Đặt business validation trong generated handler.
- Trả raw domain error làm gRPC status không kiểm soát.

## Mastery Check

- [ ] Tôi biết thêm gRPC không nên sửa domain.
- [ ] Tôi biết protobuf là transport schema.
- [ ] Tôi biết map application error sang gRPC status.
