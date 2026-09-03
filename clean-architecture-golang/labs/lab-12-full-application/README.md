# Lab 12: Full Application

## Mục tiêu

Ghép các phần thành mini banking production-style theo từng phase.

## Scope

- Customer.
- Account.
- Deposit.
- Withdrawal.
- Money Transfer.
- Transaction History.
- REST API.
- PostgreSQL.
- Redis.
- Kafka/outbox.
- Observability.

## Câu hỏi

- Bounded context nào cần tách?
- Transaction boundary của transfer nằm đâu?
- Event nào cần publish?
- Phần nào là over-engineering nếu traffic/domain còn nhỏ?

## Mastery Check

- [ ] Tôi thiết kế được module lớn theo boundary.
- [ ] Tôi biết thêm adapter mà không sửa domain.
- [ ] Tôi biết production concern thuộc layer nào.
