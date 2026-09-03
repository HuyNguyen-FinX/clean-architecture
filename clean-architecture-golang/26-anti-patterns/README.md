# 26 Anti Patterns

## Tại sao cần học?

Biết pattern chưa đủ. Production code thường hỏng vì áp dụng pattern sai context hoặc quá tay.

## Danh sách cần nhận diện

- Anemic Domain Model.
- God Service.
- Interface Everywhere.
- Repository chỉ wrap SQL.
- Domain phụ thuộc framework.
- HTTP status code trong domain.
- PostgreSQL type trong entity.
- Circular dependency.
- Global singleton.
- Excessive abstraction.
- Generic repository cho domain-rich system.

## Over-Engineering

Một Todo CRUD có thể không cần `entity/usecase/repository/adapter/gateway/presenter/controller/factory`. Hãy so sánh complexity hiện tại với complexity abstraction thêm vào.

## Mastery Check

- [ ] Tôi biết phân biệt abstraction có ích và ceremony.
- [ ] Tôi biết phát hiện domain bị framework chi phối.
- [ ] Tôi biết giải thích tại sao generic repository thường yếu với nghiệp vụ giàu rule.
