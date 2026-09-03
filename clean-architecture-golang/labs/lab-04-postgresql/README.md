# Lab 04: PostgreSQL Adapter và Row Mapping

Thời lượng: 90-150 phút. PostgreSQL integration test chạy khi có TEST_DATABASE_URL; mapping test luôn chạy.

## Mục tiêu

- thay storage detail mà application port không đổi;
- map persistence row qua domain constructor;
- thêm schema constraint;
- map not-found và giữ context error;
- test migration/query bằng PostgreSQL thật.

## Kiến thức cần

- [Repository](../../05-repository-pattern/README.md)
- [Database](../../10-database/README.md)
- pgx v5, PostgreSQL và Docker/local database.

## Diagram

~~~mermaid
flowchart LR
    APP["GetAccount"] --> PORT["AccountRepository"]
    PG["postgres.Repository"] -.implements.-> PORT
    PG --> DOMAIN["Account"]
    PG --> POOL["pgxpool"]
    POOL --> DB[("PostgreSQL")]
~~~

## Problem

Starter dùng một AccountRecord public cho cả application và persistence. Trạng thái balance âm có thể đi thẳng lên core mà không qua invariant. Hãy tách row private trong adapter và chỉ trả domain.Account hợp lệ.

## Yêu cầu

1. Domain không import pgx.
2. Application không import pgx hoặc package postgres.
3. Adapter implement consumer-owned port.
4. Query parameter hóa và liệt kê columns.
5. pgx.ErrNoRows map thành application.ErrAccountNotFound.
6. Row map qua domain.RehydrateAccount.
7. Migration có PK, currency/status và balance constraints.
8. Integration test kiểm round-trip, not-found và DB constraint.

## Các bước

1. Chạy starter và đọc coupling trong record.go.
2. Tạo domain Account với private fields.
3. Định nghĩa AccountRepository ở application.
4. Tạo accountRow private và mapper.
5. Implement FindByID/Save bằng pgxpool.
6. Viết migration.
7. Chạy unit mapping tests.
8. Cấp PostgreSQL URL và chạy integration tests.

## Expected behavior

- Row hợp lệ rehydrate thành Aggregate.
- Row vi phạm invariant trả domain error đã wrap.
- ID không tồn tại giữ errors.Is với ErrAccountNotFound.
- Save rồi Find trả đúng snapshot.
- DB từ chối balance thấp hơn overdraft.

## Test

~~~bash
cd solution
go test -race ./...
TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable \
  go test -race ./...
go vet ./...
~~~

## Questions

1. Vì sao mapping error không nên bị đổi thành not-found?
2. Domain và CHECK constraint có trùng trách nhiệm không?
3. Mock pgx không chứng minh query nào?
4. Ai sở hữu/đóng pool?
5. Khi nào reuse row/domain struct là hợp lý?

## Challenge

- Thêm optimistic version.
- Phân loại PostgreSQL SQLSTATE thay vì string matching.
- Dùng Testcontainers tạo database tự động.
- Thêm EXPLAIN test/benchmark cho query history.

## Solution explanation

Solution giữ Account và port độc lập driver; postgres.Repository chứa accountRow private. Integration suite được skip nếu thiếu env để bài vẫn chạy ở máy không có Docker, nhưng CI cho adapter production phải cấp database và cấm skip. Mini-banking có implementation phong phú hơn tại [PostgreSQL adapter](../../examples/mini-banking/internal/account/infrastructure/postgres/).
