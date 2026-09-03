# Lab 02: Use Case

## Mục tiêu

Tách workflow `TransferMoney` ra khỏi domain entity và delivery layer. Lab này tập trung vào orchestration: load account, gọi domain behavior, save kết quả và trả error phù hợp.

## Yêu cầu

- Tạo `TransferMoneyUseCase`.
- Tạo command input không phụ thuộc HTTP.
- Tạo repository port nhỏ theo nhu cầu use case.
- Viết test use case bằng fake repository.

## Câu hỏi

- Logic nào thuộc `Account.Withdraw`, logic nào thuộc `TransferMoneyUseCase`?
- Use case có nên nhận `http.Request` không?
- Repository port nên nằm ở package nào trong bài này?

## Mastery Check

- [ ] Tôi test được use case không cần HTTP.
- [ ] Tôi test được use case không cần PostgreSQL.
- [ ] Tôi phân biệt được orchestration và domain invariant.
