# Cheatsheet

## Clean Architecture Trong Một Câu

Clean Architecture = Boundaries + Dependency Direction + Business Rules Independence + Separation of Policy and Details.

## Dependency Rule

Source-code dependency đi từ ngoài vào trong:

```text
delivery -> application -> domain
infrastructure -> application -> domain
```

Không để:

```text
domain -> postgres
domain -> http
application -> gin
application -> kafka client
```

## Runtime Flow Không Phải Source Dependency

Runtime có thể là:

```text
HTTP Handler -> Use Case -> PostgreSQL Repository -> PostgreSQL
```

Nhưng source dependency nên là:

```text
handler imports application
application imports domain
postgres adapter imports application and domain
domain imports only standard/value packages
```

## Interface Trong Go

Ưu tiên:

```go
type AccountRepository interface {
	FindByID(ctx context.Context, id domain.AccountID) (*domain.Account, error)
	Save(ctx context.Context, account *domain.Account) error
}
```

Tránh:

```go
type IAccountRepository interface{}
type AccountRepositoryImpl struct{}
```

Rule of thumb:

- Accept interfaces, return structs.
- Interface nên nhỏ và nằm gần consumer khi nó phục vụ boundary của consumer.
- Không tạo interface nếu chỉ có một implementation, không cần fake trong test, và không có boundary cần bảo vệ.

## Layer Responsibilities

Domain:

- Entity, Value Object, Aggregate, Domain Service, Domain Error.
- Không biết HTTP, SQL, Kafka, Redis, config, logger framework.

Application:

- Use Case, orchestration, transaction boundary, port interface.
- Biết domain và abstraction nó cần.

Delivery:

- HTTP, gRPC, CLI, Kafka consumer.
- Parse input, map DTO, call use case, map output/error.

Infrastructure:

- PostgreSQL, Redis, Kafka producer, external API client, file storage.
- Implement port.

Composition Root:

- Thường là `cmd/api/main.go`.
- Wire config, DB, repository, use case, handler.

## Entity vs DTO vs DB Model

Không mặc định gom mọi thứ vào một struct:

```go
type User struct {
	ID int64 `json:"id" db:"id"`
}
```

Tách khi:

- Domain có behavior/invariant riêng.
- API contract khác database schema.
- Queue message cần versioning.
- Bạn muốn đổi persistence mà ít ảnh hưởng domain.

Gom tạm được khi:

- CRUD rất đơn giản.
- Struct không có behavior đáng kể.
- Team chấp nhận coupling để giảm ceremony.

## Transaction Boundary

Với use case chạm nhiều aggregate/repository:

```text
Use Case starts transaction
  load aggregate
  call domain behavior
  save changes
commit
```

Repository không nên tự mở transaction riêng nếu use case cần atomicity xuyên nhiều repository.

## Những Mùi Over-Engineering

- Mỗi struct có một interface tương ứng.
- Package nhiều hơn behavior.
- Service chỉ pass-through sang repository.
- Generic repository CRUD cho domain giàu nghiệp vụ.
- Domain entity có tag `json`, `db`, `gorm`.
- Use case phụ thuộc framework chỉ để tiện parse request.
- Test toàn mock nhưng không test invariant.
