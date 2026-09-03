# Clean Architecture Golang

Repository này là một curriculum chuyên sâu để học Clean Architecture bằng Go theo hướng thực dụng: hiểu boundary, dependency direction, business rules và trade-off trước khi nghĩ đến folder structure.

Mục tiêu không phải là nhớ một sơ đồ có `entity/usecase/repository/handler`. Mục tiêu là có thể nhìn vào một backend thật và tự trả lời:

- Business logic đang nằm ở đâu?
- Source-code dependency có đi đúng hướng không?
- Interface này bảo vệ boundary hay chỉ làm code dài hơn?
- Domain có đang phụ thuộc vào database, HTTP, Kafka, Redis, framework hay config không?
- Transaction boundary có nằm ở nơi orchestration đủ ngữ cảnh không?
- Kiến trúc đang làm hệ thống rõ hơn hay đang over-engineering?

## Prerequisite

Bạn nên đã quen với:

- Go syntax, package, interface, `context.Context`, error handling.
- Backend concepts: HTTP API, database, transaction, queue, cache.
- Testing cơ bản trong Go.
- Một ít kinh nghiệm đọc code production hoặc hệ thống nhiều module.

Repository này không dạy lại Go từ đầu. Các ví dụ tập trung vào cách đặt boundary và dependency.

## Learning Path

1. Đọc [00 Software Architecture Foundations](./00-software-architecture-foundations/README.md) để thống nhất mental model về policy, detail, coupling và boundary.
2. Đọc [CONTENT_AUDIT.md](./CONTENT_AUDIT.md) để biết material nào đã đạt depth và khoảng trống production nào còn mở.
3. Đọc [ROADMAP.md](./ROADMAP.md) để đi theo từng phase.
4. Chạy ví dụ trong [examples/mini-banking](./examples/mini-banking) để thấy domain, use case, repository port, adapter và delivery layer làm việc cùng nhau.
5. Làm lab theo thứ tự trong [labs](./labs).
6. Làm [architecture exercises](./exercises) và [code review exercises](./code-review-exercises) để luyện phân tích trade-off.
7. Đọc [case studies](./case-studies) để nối nguyên tắc với hệ thống production.
8. Dùng [CHEATSHEET.md](./CHEATSHEET.md) và [GLOSSARY.md](./GLOSSARY.md) khi review code hoặc thiết kế module mới.
9. Theo dõi [PROGRESS.md](./PROGRESS.md) để biết phần nào đã hoàn thiện sâu, phần nào mới là khung học tập đang được mở rộng.

## Repository Map

```text
clean-architecture-golang/
├── README.md
├── ROADMAP.md
├── PROGRESS.md
├── CONTENT_AUDIT.md
├── GLOSSARY.md
├── CHEATSHEET.md
├── 00-software-architecture-foundations/
├── 01-clean-architecture-foundations/
├── 02-dependency-rule/
├── 03-domain-layer/
├── 04-usecase-application-layer/
├── 05-repository-pattern/
├── 06-delivery-layer/
├── 07-infrastructure-layer/
├── 08-dependency-injection/
├── 09-project-structure/
├── 10-database/
├── 11-transaction-management/
├── 12-http-rest-api/
├── 13-grpc/
├── 14-kafka-event-driven/
├── 15-redis-cache/
├── 16-external-services/
├── 17-error-handling/
├── 18-validation/
├── 19-logging-observability/
├── 20-testing/
├── 21-concurrency-golang/
├── 22-domain-driven-design/
├── 23-cqrs-event-driven/
├── 24-production-architecture/
├── 25-refactoring/
├── 26-anti-patterns/
├── 27-case-studies/
├── 28-system-design/
├── 29-interview-review/
├── examples/
├── labs/
├── exercises/
├── code-review-exercises/
└── case-studies/
```

Các chapter được đánh số theo thứ tự học. Số thứ tự không có nghĩa là package trong Go phải được chia giống hệt như vậy.

## Setup

Yêu cầu:

- Go 1.22 hoặc mới hơn.

Kiểm tra nhanh:

```bash
go version
```

Chạy test cho ví dụ banking:

```bash
cd examples/mini-banking
go test ./...
```

Chạy API mẫu:

```bash
cd examples/mini-banking
go run ./cmd/api
```

Gửi request transfer:

```bash
curl -X POST http://localhost:8080/transfers \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: demo-transfer-001' \
  -d '{"from_account_id":"A-100","to_account_id":"B-200","amount":500000,"currency":"VND"}'
```

Response trả `transfer_id`; gửi lại cùng key/body là replay và không chuyển tiền lần hai. Xem history tại `GET /accounts/A-100/transfers` và metrics tại `GET /metrics`. PostgreSQL/Kafka setup cùng guarantee matrix nằm trong [mini-banking README](./examples/mini-banking/README.md).

## Cách Học

Mỗi chapter nên được đọc theo 3 lớp:

- Level 1: nắm trực giác và vấn đề thực tế.
- Level 2: đọc code Go và hiểu package nào phụ thuộc package nào.
- Level 3: tự phân tích coupling, boundary, trade-off và khi nào nên bỏ bớt abstraction.

Sau mỗi chapter, hãy tự hỏi:

- Nếu thay PostgreSQL bằng adapter khác, domain có đổi không?
- Nếu thêm gRPC cạnh HTTP, use case có đổi không?
- Nếu bỏ interface này, test hoặc boundary nào bị ảnh hưởng?
- Nếu gom DTO, DB model và Entity thành một struct, coupling nào xuất hiện?

## Triết Lý

Clean Architecture là công cụ để bảo vệ business rules khỏi chi tiết thay đổi nhanh. Nó không phải checklist folder, không phải lý do để tạo interface cho mọi struct, và không phải cách biến Go thành Java.

Một CRUD nhỏ có thể chỉ cần cấu trúc đơn giản. Một hệ thống banking, payment, loan, order hoặc workflow phức tạp thường hưởng lợi nhiều hơn từ boundary rõ ràng. Kiến trúc tốt là kiến trúc trả tiền đúng mức cho complexity hiện tại và complexity sắp tới.
