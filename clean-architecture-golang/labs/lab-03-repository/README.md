# Lab 03: Repository

## Mục tiêu

Thiết kế repository port theo ngôn ngữ use case/domain, sau đó viết in-memory adapter để test.

## Yêu cầu

- Tạo `AccountRepository` interface với method vừa đủ.
- Implement `memory.Repository`.
- Repository trả domain entity, không trả DTO hoặc DB row.
- Viết test chứng minh use case không biết implementation.

## Câu hỏi

- Repository khác DAO ở điểm nào?
- Vì sao không dùng generic CRUD repository ở lab này?
- Interface đặt ở application hay infrastructure?

## Mastery Check

- [ ] Tôi biết adapter implement port mà không cần keyword `implements`.
- [ ] Tôi biết fake repository giúp test use case.
- [ ] Tôi biết giữ repository interface nhỏ.
