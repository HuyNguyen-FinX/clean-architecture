# Clean Architecture Anti-Patterns: code “đúng form” nhưng sai lực

Anti-pattern production thường không lộ liễu. Code có interface, constructor và folder đẹp nhưng boundary vẫn sai, guarantee bị che hoặc ceremony lớn hơn domain.

## Kết quả học tập

- review dependency/semantic/ownership, không nhìn tên;
- nhận diện 12 anti-patterns bằng Go;
- phân biệt justified exception;
- refactor theo pressure thực;
- tránh over-engineering Todo CRUD.

## 1. Interface everywhere

~~~go
type IUserService interface { CreateUser(...) }
type UserServiceImpl struct { ... }
~~~

Chỉ một implementation, interface do producer đặt, consumer vẫn cần toàn API. Cost: navigation/mock/renames.

Refactor: trả concrete type; interface nhỏ xuất hiện ở consumer khi cần capability/inversion.

Exception: public plugin API hoặc multiple implementations planned bằng evidence.

## 2. Generic Repository

~~~go
type Repository[T any] interface {
	Find(context.Context, map[string]any) ([]T, error)
	Update(context.Context, map[string]any, map[string]any) error
}
~~~

Xóa domain language/type safety; query DSL leak. Account cần load-for-consistency, history cần projection, không cùng generic CRUD contract.

Generic helper nội bộ adapter cho repetitive scan có thể được, nhưng port vẫn semantic.

## 3. Repository là DAO đổi tên

~~~go
type AccountRepository interface {
	SelectAccountRow(ctx context.Context, id string) (postgres.AccountRow, error)
}
~~~

Interface không đảo ownership vì application phụ thuộc persistence model. Map row ở adapter; contract trả domain/read model.

## 4. Anemic Domain

~~~go
type Account struct {
	Balance int64
	Status string
}

func (s *AccountService) Withdraw(a *Account, n int64) {
	a.Balance -= n
}
~~~

Mọi caller có thể bypass. Đưa invariant/transition vào private-state Account. Transaction/I/O vẫn ở Application.

Exception: CRUD/data pipeline domain không có behavior, anemic record có thể chân thật hơn fake rich model.

## 5. God Application Service

~~~go
type Service struct {
	db, cache, kafka, email, risk, payment any
}
~~~

Một type có 50 methods, interface lớn, transaction/retry boundaries mờ. Tách theo use case/capability; shared helper chỉ cho cohesive policy.

Đừng tách mỗi 10 dòng thành service mới; cohesion trước file size.

## 6. Domain import framework

~~~go
func (a *Account) Withdraw(c *gin.Context, db *gorm.DB, amount int64) error
~~~

Domain bị request/ORM shape; tests nặng; entry point khác không reuse. Domain nhận Value Object; adapters/application quản context/I/O.

## 7. HTTP error trong Domain

~~~go
type DomainError struct {
	Status int
	Code string
}
~~~

Kafka/gRPC bị ép vocabulary HTTP. Domain error stable semantic; adapter map protocol.

Exception: nếu “status” là business status, tên rõ, không net/http constants.

## 8. Transaction per Repository method

Hai Save commit riêng làm transfer partial. Use case/UoW sở hữu multi-step atomic boundary. Single SQL method atomic vẫn có thể tự transaction.

## 9. Hidden transaction in context without discipline

Context tx tiện nhưng caller dùng context.Background thoát transaction. Contract/test/nested policy thiếu.

Refactor options:

- private typed key + integration test;
- explicit Tx repositories;
- UoW callback nhận capabilities.

Không cấm tuyệt đối; công bố trade-off.

## 10. Fire-and-forget side effect

~~~go
go publisher.Publish(context.Background(), event)
return nil
~~~

Không ownership/retry/shutdown/durability. Dùng synchronous guarantee hoặc outbox/owned worker.

Best-effort analytics có thể drop nếu policy/metric rõ.

## 11. Retry mọi lỗi

~~~go
for {
	if err := call(); err == nil { break }
}
~~~

Retry business reject, canceled, non-idempotent; outage thành load storm. Classify, bound, jitter, budget, idempotency và metrics.

## 12. DTO cho mọi boundary máy móc

Request → HandlerDTO → ApplicationDTO → DomainDTO → RowDTO, cùng fields, mapper chỉ copy. Cost lớn hơn decoupling.

Tách khi models có change/security/nullability/behavior khác. Reuse có chủ đích trong CRUD nhỏ.

## 13. Shared/common/utils dump

Package shared import mọi context và chứa errors/models/helpers. Blast radius/ownership mờ. Di chuyển concept về owner; duplicate vài dòng nếu coupling không đáng.

## 14. Service Locator/global singleton

~~~go
var Container map[string]any

func Execute() {
	repo := Container["repo"].(Repository)
}
~~~

Hidden dependency/runtime panic/test interference. Constructor injection/composition root.

Global immutable constants/logger default đôi khi chấp nhận, nhưng mutable core dependencies cần ownership.

## 15. Mapping everything to interface{}

Generic event/error/result map làm compile-time contract mất. Dùng typed command/event; envelope generic chỉ ở integration boundary.

## 16. Mock-heavy tests

~~~text
EXPECT logger
EXPECT metric
EXPECT repo exact order
EXPECT publisher
~~~

Test implementation choreography. Dùng fake/state assertions; spy cho meaningful output; integration cho SQL.

## 17. Layer pass-through

~~~text
Controller → Service → UseCase → Manager → Repository → DAO
~~~

Mỗi method forward same args. Merge lớp không có policy/translation/ownership. Boundary phải mua được isolation hoặc language.

## 18. Framework abstraction fantasy

Tự viết UniversalRouter/UniversalORM/UniversalQueue để “đổi vendor”, rồi chỉ hỗ trợ lowest common denominator và duplicate bugs.

Abstraction theo capability cần thay/test, không wrap toàn framework.

## 19. Clean Architecture = microservices

Tách services trước data/team/domain boundary thêm network/consistency. Modular monolith có package/data ownership thường tốt hơn. Distributed boundary đắt và khó undo.

## 20. Folder-driven design

Tạo entity/usecase/adapter trước hiểu requirements. Result là placeholder interfaces. Bắt đầu actor/use case/rules/failure, rồi structure.

## 21. Todo API: Version A và B

Version A:

~~~text
todo/
  handler.go
  service.go
  store.go
~~~

Version B:

~~~text
entity/usecase/port/repository/adapter/gateway/
presenter/controller/mapper/factory/
~~~

Nếu Todo chỉ CRUD, B tăng navigation/codegen/mocks mà không bảo vệ volatility. Khi có collaborative permissions, recurring scheduling, state rules và multiple entries, tách selected boundaries dần.

Proportionality là architecture skill.

## 22. Smell diagnosis questions

1. Business rule nằm đâu và có thể bypass?
2. Interface owner là consumer?
3. Adapter semantics có tương đương?
4. Transaction/retry/idempotency guarantee rõ?
5. Change một API/schema chạm bao package?
6. Test đang verify behavior hay calls?
7. Có layer nào chỉ forward?
8. Exception có documented reason/exit condition?

## 23. Production scenario

Service “clean” publish Kafka qua port sau DB commit, không outbox. Kafka down làm error, client retry; DB effect lặp. Interface đúng direction nhưng distributed guarantee sai.

Architecture review phải hỏi runtime failure, không chỉ compile imports.

## 24. Refactor strategy

- lock behavior;
- remove pass-through;
- move rule vào owner;
- shrink consumer interface;
- isolate model mapping;
- move transaction to operation;
- replace fire-and-forget with outbox;
- add fitness/contract/integration tests;
- measure before more abstraction.

## 25. Khi anti-pattern có thể là trade-off

Application import pgx trong service nhỏ có thể chấp nhận nếu domain trivial, team nhỏ, integration tests mạnh. Viết decision/risk. Anti-pattern không là syntax ban; nó là recurring solution có harmful context.

## 26. Bài tập

Làm [Code Review 01](../code-review-exercises/01-god-handler/problem.md). Tìm cả subtle issues: split transaction, framework error, lost context, duplicate Kafka.

## 27. Mastery questions

1. Interface nào là abstraction giả?
2. Pass-through layer gây cost gì?
3. Port đúng direction vẫn sai guarantee ra sao?
4. Khi anemic model hợp lý?
5. Context transaction cần guard gì?
6. DTO duplication khi nào trả cost?
7. Global immutable khác mutable singleton?
8. Todo khi nào nên tăng architecture?

## Further reading

- Refactoring, Martin Fowler.
- Working Effectively with Legacy Code.
- Go Code Review Comments và package design guidance.
- Clean Architecture/DDD texts, đọc như heuristics có context.

## Quality gate

- [x] 20 anti-pattern/trade-offs
- [x] Subtle Go/dependency/runtime examples
- [x] Justified exceptions
- [x] Todo proportionality
- [x] Production scenario/diagnosis/refactor
- [x] Review exercise/mastery
