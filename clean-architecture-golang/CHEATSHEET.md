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

## Decision Flow Cho Một Use Case

~~~text
1. Actor muốn đạt outcome nào?
2. Invariant nào phải luôn đúng?
3. Dữ liệu nào cần strong consistency?
4. I/O nào là local DB, I/O nào qua network?
5. Retry/duplicate/timeout có thể xảy ra ở đâu?
6. Application cần capability nào từ bên ngoài?
7. Adapter nào implement từng capability?
8. Test nào chứng minh từng guarantee?
~~~

Nếu chưa trả lời 1-5, khoan tạo interface/folder. Boundary nên xuất phát từ ownership và failure model.

## Failure Checklist

| Câu hỏi | Cơ chế thường cân nhắc |
|---|---|
| Hai writers cùng sửa? | row lock, version, serializable + retry |
| Commit xong response mất? | idempotency key + status lookup |
| DB commit nhưng Kafka down? | transactional outbox |
| Consumer nhận duplicate? | inbox/idempotent business effect |
| External timeout không rõ outcome? | stable operation ID + inquiry/reconciliation |
| Retry storm? | classification, budget, backoff+jitter, load shedding |
| Process bị SIGTERM? | cancel intake, drain bounded, close resource |
| Log/trace quá nhiều cardinality? | route/error class labels; ID chỉ log/trace |

## Code Review Nhanh

- Tìm imports từ `domain/application` tới framework/client cụ thể.
- Tìm public mutable fields làm bypass behavior.
- Theo transaction handle đến từng Repository call; `BEGIN` tồn tại chưa chứng minh write dùng transaction.
- Tìm goroutine không owner, `context.Background()` trong request flow và ignored errors.
- Kiểm tra HTTP/Kafka DTO có bị dùng làm Entity không.
- Kiểm tra retry có stable identity và chỉ retry lỗi transient không.
- Kiểm tra publish side effect có failure window với DB commit không.
- Đòi test thật cho SQL/locking/schema; mock không chứng minh database semantics.

## Khi Nào Giữ Đơn Giản

Một handler + concrete store có thể đủ khi CRUD nhỏ, một actor, một DB và failure cost thấp. Dấu hiệu tăng kiến trúc:

- Rule được lặp giữa HTTP/worker.
- Một operation chạm nhiều records/repositories.
- API, schema và domain đổi vì lý do khác nhau.
- External outcome, retry hoặc idempotency trở thành business concern.
- Module có owner/release cadence độc lập.

Thêm đúng boundary đang cần; không nhân bản bốn layer đồng loạt cho mọi feature.
