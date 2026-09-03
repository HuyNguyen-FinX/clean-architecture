# Case Studies

Case study là nơi áp dụng nguyên tắc vào hệ thống có context cụ thể. Mỗi case phân tích:

- Requirements.
- Domain.
- Use Cases.
- Boundaries.
- Interfaces.
- Infrastructure.
- Database.
- Transactions.
- Concurrency.
- Error Handling.
- Testing.
- Observability.
- Trade-offs.

Đừng đọc case study như một reference architecture để copy. Mỗi quyết định chỉ đúng dưới assumptions đã ghi; thay consistency, traffic, team ownership hoặc failure cost thì boundary có thể đổi.

## Cách Học Một Case

1. Chỉ đọc phần bối cảnh rồi tự vẽ model/import graph trước.
2. Liệt kê strong invariants và operations có side effect.
3. Đánh dấu ranh giới local transaction và network.
4. Lập crash matrix: trước/sau mỗi write, commit, publish và response.
5. Đọc phương án trong case, ghi điểm bạn chọn khác và evidence.
6. Trả lời mastery questions không nhìn bài.

## Trục So Sánh

| Case | Domain richness | Consistency chính | Distributed workflow | Bài học trung tâm |
|---|---|---|---|---|
| Todo | thấp -> vừa | optimistic edit | reminder tùy chọn | proportional architecture |
| Order | cao | Aggregate + eventual cross-service | Saga/compensation | state không được nói dối |
| Payment | cao/risk lớn | operation identity | provider + webhook | ambiguous outcome |
| Banking | cao/risk rất lớn | atomic posting | outbox | transaction + idempotency |
| Loan | cao/lifecycle dài | decision evidence | process manager | context/language/audit |
| Kafka worker | orchestration | inbox + offset | at-least-once | ack sau durable effect |
| Batch | policy + execution | per-item idempotency | checkpoint/lease | resume và backpressure |
| Microservice | tùy capability | data ownership | API/event | network không chữa coupling |

Không có cột "số layer" vì nó không phải biến thiết kế quan trọng.

## Danh sách

- `01-todo-api`
- `02-ecommerce-order`
- `03-payment-service`
- `04-banking-account`
- `05-loan-service`
- `06-kafka-worker`
- `07-batch-processing`
- `08-microservice`

## Bài Tổng Hợp

Chọn một case và thay một assumption:

- Todo tăng từ 100 RPS lên collaborative editing real-time.
- Order quay về cùng modular monolith/database.
- Payment provider không hỗ trợ idempotency hoặc lookup.
- Banking chuyển từ mutable balance sang double-entry ledger.
- Batch từ nightly snapshot thành continuous stream.

Viết lại decision record. Không được chỉ thêm Redis/Kafka/microservice; phải chỉ ra invariant, source of truth, transaction, recovery và operational owner thay đổi ra sao.

## Rubric

| Mức | Dấu hiệu |
|---:|---|
| 0 | kể tên pattern, không có assumptions |
| 1 | có diagram nhưng không có dependency/failure reasoning |
| 2 | model và ports hợp lý, transaction còn mơ hồ |
| 3 | failure/idempotency/testing rõ, trade-off có evidence |
| 4 | có migration, operations, limits và alternative đơn giản hơn |

Một thiết kế tốt không phải thiết kế giống tài liệu. Nó là thiết kế có chain reasoning kiểm chứng được từ requirement tới boundary và guarantee.
