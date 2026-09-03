# Lab 07: Testing

## Mục tiêu

Thiết kế test pyramid cho mini banking.

## Yêu cầu

- Domain tests không mock.
- Use case tests dùng fake repository.
- HTTP tests dùng fake use case hoặc in-memory wiring.
- Repository integration test dùng database thật khi thêm PostgreSQL.

## Câu hỏi

- Khi nào mock làm test khó đọc hơn fake?
- Có nên mock domain entity không?
- Test nào nên chạy nhanh trong CI mỗi commit?

## Mastery Check

- [ ] Tôi biết chọn loại test theo layer.
- [ ] Tôi biết tránh mock mọi thứ.
- [ ] Tôi biết integration test adapter thật.
