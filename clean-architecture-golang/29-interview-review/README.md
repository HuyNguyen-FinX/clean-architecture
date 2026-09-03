# 29 Interview Review

## Tại sao cần học?

Review giúp chuyển kiến thức thành khả năng giải thích. Senior engineer không chỉ viết code sạch, mà còn bảo vệ quyết định architecture bằng trade-off rõ ràng.

## Chủ đề ôn tập

- Dependency Rule.
- Compile-time dependency vs runtime flow.
- Interface placement.
- Entity vs DTO vs DB model.
- Transaction boundary.
- Repository vs DAO.
- HTTP/gRPC/Kafka adapter.
- Testing từng layer.
- Over-engineering.

## Dạng câu hỏi

```text
Nếu Use Case phụ thuộc *sql.DB, Dependency Rule có bị vi phạm không?
Tại sao gọi repository từ use case không đồng nghĩa use case phụ thuộc PostgreSQL?
Kafka consumer thuộc layer nào?
Transaction nên bắt đầu ở đâu?
Khi nào không cần tách DTO và Entity?
```

## Mastery Check

- [ ] Tôi trả lời được bằng nguyên lý, không chỉ bằng tên pattern.
- [ ] Tôi đưa được ví dụ Go cụ thể.
- [ ] Tôi phân tích được cả lợi ích và chi phí.
