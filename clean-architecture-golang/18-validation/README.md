# 18 Validation

## Tại sao cần học?

Validation có nhiều lớp. Không phải mọi validation đều là business rule, và không phải mọi business rule đều nên nằm ở request validator.

## Phân loại

- Transport validation: JSON hợp lệ, required field, format cơ bản.
- Application validation: actor có quyền gọi use case, command đầy đủ.
- Domain validation: invariant như balance, state transition, currency.

## Ví dụ

HTTP handler có thể reject body thiếu `amount`. Nhưng rule `Balance không được nhỏ hơn overdraft limit` thuộc domain method `Withdraw`.

## Anti-pattern

- Duplicate domain invariant ở handler.
- Dùng validation tag làm nguồn sự thật duy nhất cho domain.
- Entity phụ thuộc validator framework.

## Mastery Check

- [ ] Tôi biết validation nào thuộc delivery, application, domain.
- [ ] Tôi tránh duplicate rule nghiệp vụ.
- [ ] Tôi biết DTO validation không thay thế domain invariant.
