# 25 Refactoring

## Tại sao cần học?

Nhiều hệ thống không bắt đầu sạch. Kỹ năng quan trọng là refactor từng bước từ spaghetti hoặc layered architecture sang boundary rõ hơn mà không rewrite toàn bộ.

## Flow Refactor

```text
Step 1: Identify business rules
Step 2: Extract domain behavior
Step 3: Extract repository port
Step 4: Create use case
Step 5: Move infrastructure outward
Step 6: Add tests around changed boundary
```

## Nguyên tắc

Refactor theo lát cắt use case. Không cần gom toàn bộ project vào Clean Architecture trong một commit lớn.

## Anti-pattern

- Rewrite toàn bộ rồi mất behavior cũ.
- Chỉ di chuyển file, không đổi dependency.
- Tạo abstraction trước khi hiểu business rule.
- Không khóa behavior bằng test.

## Mastery Check

- [ ] Tôi biết refactor theo từng use case.
- [ ] Tôi biết xác định rule trước khi tạo folder.
- [ ] Tôi biết dùng test để giữ behavior khi tách layer.
