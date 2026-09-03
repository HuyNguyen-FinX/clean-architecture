# 05 Repository Pattern

## Tại sao cần học?

Repository là một trong những pattern bị hiểu sai nhiều nhất. Repository không đơn giản là DAO CRUD. Nó là abstraction giúp application lấy và lưu aggregate theo ngôn ngữ domain.

## Repository Không Phải DAO

DAO thường bám sát table:

```text
InsertUser
UpdateUser
FindUserRow
```

Repository nên bám sát aggregate/use case:

```text
FindAccountByID
SaveAccount
FindPendingTransfers
```

## Dependency

Application định nghĩa repository port khi nó là consumer. PostgreSQL adapter implement port. Domain không biết adapter.

## Go Implementation

```go
type AccountRepository interface {
	FindByID(ctx context.Context, id domain.AccountID) (*domain.Account, error)
	Save(ctx context.Context, account *domain.Account) error
}
```

Không tạo generic repository chỉ vì Go có generics. Domain-rich system thường cần method theo intent.

## Anti-pattern

- Repository chỉ wrap SQL mà không che persistence detail.
- Interface quá lớn vì phục vụ mọi use case.
- Repository trả DB model ra application.
- Application phụ thuộc `*sql.DB` hoặc `*pgxpool.Pool` trực tiếp.

## Mastery Check

- [ ] Tôi phân biệt Repository, DAO, Gateway, Data Mapper.
- [ ] Tôi biết repository interface nên nhỏ theo use case.
- [ ] Tôi biết khi nào repository là ceremony thừa cho CRUD rất nhỏ.
