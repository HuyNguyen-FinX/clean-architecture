# Glossary

## Abstraction

Abstraction là cách che bớt chi tiết để code phụ thuộc vào khả năng cần dùng thay vì phụ thuộc vào cơ chế cụ thể. Trong Go, abstraction thường là interface nhỏ do consumer định nghĩa.

## Adapter

Adapter là code chuyển đổi giữa thế giới bên ngoài và core application. HTTP handler, gRPC server, PostgreSQL repository, Kafka producer, Redis cache client đều là adapter.

## Aggregate

Aggregate là một cụm domain object có invariant cần được bảo vệ nhất quán. Aggregate Root là object bên ngoài được phép tham chiếu trực tiếp.

## Application Layer

Application Layer chứa use case và orchestration: load dữ liệu, gọi domain behavior, quản lý transaction, gọi port, trả output. Nó không nên biết HTTP status code hoặc SQL query cụ thể.

## Boundary

Boundary là ranh giới giữa policy và detail. Boundary tốt giúp code bên trong ít bị ảnh hưởng khi database, framework, message broker hoặc transport thay đổi.

## Business Rule

Business Rule là luật nghiệp vụ làm hệ thống có ý nghĩa. Ví dụ: tài khoản không được withdraw vượt overdraft limit. Rule này thuộc domain, không thuộc PostgreSQL hoặc HTTP.

## Clean Architecture

Clean Architecture là cách tổ chức dependency để business rules độc lập với framework, UI, database và external services. Nó là tư duy về boundary, không phải cấu trúc folder cố định.

## Cohesion

Cohesion đo mức độ các phần trong một module cùng phục vụ một lý do thay đổi. Cohesion cao thường làm code dễ hiểu và dễ đổi hơn.

## Coupling

Coupling là mức độ một module bị ràng buộc với module khác. Coupling không xấu tuyệt đối, nhưng coupling sai hướng làm business logic bị kéo theo chi tiết kỹ thuật.

## Dependency Inversion

Dependency Inversion là nguyên tắc để policy không phụ thuộc trực tiếp vào detail. High-level code phụ thuộc vào abstraction; low-level adapter implement abstraction đó.

## Dependency Injection

Dependency Injection là cách đưa dependency từ bên ngoài vào object, thường qua constructor trong Go. `main.go` thường đóng vai trò Composition Root.

## Dependency Rule

Dependency Rule nói rằng source-code dependency chỉ được đi từ ngoài vào trong. Runtime call có thể đi qua adapter ra database, nhưng import dependency của core không được trỏ ra detail.

## Domain Entity

Domain Entity là object có identity và behavior nghiệp vụ. Entity không nên chỉ là struct chứa field rồi để service sửa tự do.

## Domain Service

Domain Service chứa domain logic không tự nhiên thuộc về một Entity hoặc Value Object cụ thể, nhưng vẫn là business rule thuần domain.

## DTO

DTO là data shape dùng để đi qua boundary, ví dụ HTTP request/response, Kafka message hoặc gRPC message. DTO không nhất thiết là Domain Entity.

## Gateway

Gateway là port/adapter đại diện cho external system, ví dụ payment provider, core banking API hoặc email service.

## Infrastructure

Infrastructure là chi tiết kỹ thuật: database, queue, cache, file system, HTTP client, cloud SDK, config loader.

## Repository

Repository là abstraction cho việc lấy và lưu aggregate theo ngôn ngữ domain. Repository không chỉ là CRUD wrapper quanh SQL.

## Unit of Work

Unit of Work gom nhiều thay đổi vào một transaction boundary. Trong Go production, nó thường được thể hiện bằng Transaction Manager hoặc function `WithinTx`.

## Use Case

Use Case là một hành động application có ý nghĩa với actor hoặc system, ví dụ `TransferMoney`, `CreateOrder`, `ApproveLoan`. Use case orchestration chứ không nên nhồi toàn bộ domain rule vào service.

## Value Object

Value Object là object được định danh bằng giá trị, thường immutable theo convention. Ví dụ `Money`, `Email`, `AccountNumber`.

## Anti-Corruption Layer (ACL)

ACL map ngôn ngữ/model của external system sang model nội bộ để thay đổi hoặc sự kỳ quặc của provider không lan vào core. Ví dụ Payment adapter đổi provider status `requires_action` thành `AuthorizationPending`; raw SDK type không phải Domain Entity. Xem [External Services](./16-external-services/README.md).

## Application Service

Application Service triển khai một use case: authorize, load, mở transaction, gọi Domain behavior, lưu và tạo side-effect intent. Nó khác Domain Service ở chỗ orchestration biết ports/I/O; Domain Service giữ business rule thuần. Xem [Application Layer](./04-usecase-application-layer/README.md).

## Bounded Context

Bounded Context là ranh giới trong đó một ubiquitous language/model có meaning nhất quán. `Customer` ở Lending có thể là Applicant risk snapshot, còn ở Support là contact profile; ép dùng chung một struct làm semantic coupling. Bounded Context không bắt buộc là microservice.

## Compile-time Dependency

Quan hệ được tạo bởi import/type reference khi build source. Nếu `postgres` implement interface application, runtime application gọi Postgres nhưng compile-time arrow là `postgres -> application`. Đây là trọng tâm của Dependency Rule. Xem [Dependency Rule](./02-dependency-rule/README.md).

## Composition Root

Nơi tạo concrete dependencies, cấu hình object graph và sở hữu lifecycle, thường là `cmd/api/main.go`. Đây là nơi hợp lệ để import HTTP, pgx, Kafka và application cùng lúc. Composition Root không phải Service Locator: core nhận dependency qua constructor thay vì tự tìm global object.

## Consistency Boundary

Phạm vi invariant phải đúng ngay sau một operation/transaction. Aggregate thường được thiết kế như consistency boundary, nhưng distributed workflow qua nhiều Aggregates/Services thường chỉ đạt eventual consistency. Aggregate quá lớn tăng contention; quá nhỏ làm invariant mạnh không có nơi bảo vệ.

## Data Dependency

Coupling do shape/meaning của dữ liệu, kể cả không có import source. Hai services dùng JSON cùng field vẫn phụ thuộc semantic contract. Vì vậy "không import nhau" chưa đủ chứng minh decoupling.

## Domain Error

Lỗi diễn đạt business outcome/invariant như `ErrInsufficientBalance`. Nó không chứa HTTP status, gRPC code hay Kafka retry flag. Adapter/application map lỗi theo contract của từng boundary. Xem [Error Handling](./17-error-handling/README.md).

## Domain Event

Fact có meaning bên trong domain, ví dụ `MoneyTransferred`. Nó có thể chưa phải Kafka message. Application/outbox mapper tạo Integration Event versioned, chọn field cần publish và loại dữ liệu nhạy cảm. Domain Event và broker payload không nên bị đồng nhất máy móc.

## Driving Adapter

Adapter khởi phát use case: HTTP handler, gRPC server, CLI, scheduler, Kafka consumer. Nó map external input sang command và map result/error ngược ra protocol. Từ "driving" nói runtime role, không cho phép nó sở hữu business rule.

## Driven Adapter

Adapter được application gọi qua port: PostgreSQL repository, Kafka publisher, Redis cache, provider client. Runtime arrow đi ra ngoài; compile-time adapter implement/import port đi vào trong.

## Eventual Consistency

Các state liên quan có thể tạm thời khác nhau nhưng hội tụ qua event/retry/reconciliation. Nó cần pending state, freshness/SLO và recovery rõ; không có nghĩa "cuối cùng chắc sẽ đúng" nếu không có durable intent và operator path.

## Idempotency

Cùng một logical operation được thực hiện lặp lại mà business effect không nhân lên. Key cần scope, canonical request hash và durable arbitration. Idempotency khác transaction: transaction bảo vệ atomicity một attempt; idempotency nối nhiều attempts. Xem [Transaction Management](./11-transaction-management/README.md).

## Inbox Pattern

Consumer lưu message identity cùng transaction với business effect. Redelivery sau crash trở thành known duplicate. Check Redis rồi ghi PostgreSQL không atomic nên vẫn có failure window nếu business effect nằm ở PostgreSQL.

## Integration Event

Contract versioned được publish giữa bounded contexts/services. Nó thuộc public integration language, cần compatibility, privacy và retention. Không serialize nguyên Aggregate để làm integration event.

## Invariant

Điều kiện phải luôn đúng tại boundary business, ví dụ balance không thấp hơn overdraft limit. Required JSON field là transport validation; overdraft là invariant dù command đến từ HTTP, Kafka hay CLI. Xem [Domain Invariant](./03-domain-layer/03-invariant.md).

## Optimistic Concurrency

Đọc version rồi update với điều kiện expected version; conflict thì reload/reject/retry. Hợp khi conflict hiếm. Nó không "không có lock": database vẫn dùng synchronization nội bộ, nhưng application không giữ row lock suốt read-think-write.

## Outbox Pattern

Business state và message intent được ghi cùng local DB transaction. Relay publish rồi mark. Crash sau publish/trước mark có thể duplicate, nên outbox tạo at-least-once chứ không exactly-once end-to-end. Xem [Kafka/Event Driven](./14-kafka-event-driven/README.md).

## Pessimistic Locking

Khóa row trước khi quyết định, ví dụ `SELECT ... FOR UPDATE`. Dễ reason cho conflict cao nhưng làm request chờ, có deadlock và hot-key contention. Lock order ổn định giảm một lớp deadlock, không loại bỏ mọi deadlock.

## Policy Và Detail

Policy trả lời hệ thống phải quyết định gì; detail trả lời quyết định được thực hiện bằng cơ chế nào. `không cho rút quá overdraft` là policy; SQL, pgx và HTTP JSON là details. Detail đôi khi rất quan trọng cho correctness, nhưng vẫn nên bị khoanh khỏi business model.

## Port

Contract tại boundary mô tả capability bên trong cần hoặc cung cấp. Trong Go, port thường là interface nhỏ do consumer sở hữu. Không phải mọi concrete type cần port; interface vô nghĩa chỉ tăng navigation/coupling.

## Reconciliation

Quy trình so sánh authoritative states và sửa/chuyển manual review khi operation có outcome mơ hồ hoặc event bị trễ. Retry không thay reconciliation: retry có thể lặp command, còn reconciliation hỏi điều gì đã thực sự xảy ra.

## Runtime Call

Lời gọi thực khi chương trình chạy. `UseCase -> PostgresRepository` là runtime call hợp lệ dù source dependency được đảo bằng interface. Nhầm runtime với compile-time dẫn tới kết luận sai rằng Clean Architecture cấm core gọi database.

## Saga

Workflow nhiều local transactions được nối bởi command/event và compensation. Saga không rollback thời gian; compensation là business action mới, có thể thất bại và cần audit/manual repair. Xem [CQRS/Event Driven](./23-cqrs-event-driven/README.md).

## Semantic Dependency

Phụ thuộc vào meaning/assumption của module khác. Một package có thể không import provider nhưng vẫn coupled nếu core dùng status names và retry semantics riêng của provider. ACL và contract theo intent giảm loại coupling này.

## Transaction Boundary

Phạm vi thay đổi cần commit/rollback cùng nhau. Nó thường ở application use case vì repository riêng lẻ không biết toàn operation. Không kéo network call vào DB transaction chỉ để tạo cảm giác atomic; remote side effect cần idempotency/Saga/reconciliation.

## Ubiquitous Language

Ngôn ngữ chung giữa domain experts và engineers trong một Bounded Context, được phản ánh trong code, tests và events. Từ giống nhau ở hai context có thể có meaning khác; glossary toàn công ty không thay context map.

## Bản Đồ Quan Hệ

~~~text
Driving Adapter -> Application Service -> Domain
                         |
                         v
                    Port contract
                         ^
                         |
                    Driven Adapter

Aggregate -> protects Invariant -> defines Consistency Boundary
Use Case  -> owns Transaction Boundary -> coordinates Repositories/Gateways
Outbox    -> durable outgoing intent
Inbox     -> durable duplicate guard
~~~

Khi một thuật ngữ bị dùng như luật tuyệt đối, quay lại câu hỏi: ai sở hữu quyết định, guarantee cần là gì, failure nào đang được xử lý và chi phí abstraction có đáng không?
