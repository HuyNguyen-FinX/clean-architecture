# Lab 08: Transaction

## Mục tiêu

Đặt transaction boundary cho transfer money.

## Yêu cầu

- Tạo `Transactor` hoặc Unit of Work.
- Đảm bảo load và save nằm trong cùng transaction.
- Phân tích row locking.
- Viết test cho rollback khi save receiver lỗi.

## Câu hỏi

- Transaction nên bắt đầu ở repository hay use case?
- `context.Context` có nên giấu transaction không?
- Network call có nên nằm trong DB transaction không?

## Mastery Check

- [ ] Tôi biết transaction boundary theo use case.
- [ ] Tôi biết trade-off của `WithinTx`.
- [ ] Tôi biết atomicity local khác distributed transaction.
