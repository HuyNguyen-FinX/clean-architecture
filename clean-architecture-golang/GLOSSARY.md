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
