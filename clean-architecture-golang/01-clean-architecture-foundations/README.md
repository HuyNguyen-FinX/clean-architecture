# 01 Clean Architecture Foundations

Chapter này đang ở trạng thái `[IN PROGRESS]`: nội dung đã có tuyến học cốt lõi, nhưng sẽ còn được mở rộng bằng nhiều ví dụ và bài tập ở các increment sau.

## Tại sao cần học?

Clean Architecture trả lời một câu hỏi đơn giản nhưng khó làm đúng: làm sao để business rules không bị framework, database, transport và external service chi phối?

Nó không bắt đầu từ folder. Nó bắt đầu từ boundary.

## Mental Model

```text
Frameworks & Drivers
        ↓
Interface Adapters
        ↓
Application / Use Cases
        ↓
Domain / Entities
```

Mũi tên là source-code dependency. Code bên ngoài được phép import code bên trong. Code bên trong không import code bên ngoài.

## Nó giải quyết vấn đề gì?

Clean Architecture giúp khoanh vùng thay đổi:

- Đổi router HTTP không làm đổi use case.
- Đổi PostgreSQL adapter không làm đổi domain.
- Thêm gRPC cạnh REST không làm copy business logic.
- Test domain không cần database.
- Test use case có thể dùng fake repository.

## Dependency

Một transfer API có thể chạy như sau:

```text
HTTP Handler
↓
TransferMoneyUseCase
↓
AccountRepository interface
↑
PostgresAccountRepository
↓
PostgreSQL
```

Ở runtime, use case gọi object PostgreSQL repository. Ở source code, application chỉ biết interface. Đây là điểm khác biệt cốt lõi.

## Golang Implementation

Trong Go, Clean Architecture nên tận dụng:

- Package nhỏ, rõ responsibility.
- Interface nhỏ, thường nằm gần consumer.
- Constructor injection.
- `cmd/.../main.go` làm Composition Root.
- Không dùng prefix kiểu `IUserRepository`.
- Không tạo framework riêng trong project nếu chỉ cần function và struct.

## Trade-offs

Clean Architecture có giá trị khi hệ thống có nghiệp vụ, workflow và boundary thật. Nó có thể thừa nếu API chỉ CRUD đơn giản.

Luôn hỏi:

```text
Complexity được giảm là gì?
Complexity được thêm là gì?
Ai phải trả chi phí này?
```

## Mastery Check

- [ ] Tôi giải thích được Clean Architecture là boundary và dependency direction.
- [ ] Tôi biết framework là detail, không phải trung tâm.
- [ ] Tôi biết vì sao folder structure chỉ là implementation detail.
- [ ] Tôi có thể thêm adapter mới mà không sửa domain trong một ví dụ nhỏ.
