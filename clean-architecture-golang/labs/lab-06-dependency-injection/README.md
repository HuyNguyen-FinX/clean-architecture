# Lab 06: Dependency Injection

## Mục tiêu

Lắp object graph bằng constructor injection trong `cmd/api/main.go`.

## Yêu cầu

- Tạo config tối thiểu.
- Tạo repository adapter.
- Tạo use case.
- Tạo handler.
- Start HTTP server.

## Câu hỏi

- Vì sao `main.go` được phép biết cả infrastructure và delivery?
- Service có nên tự đọc env không?
- Khi nào Wire/Fx đáng dùng?

## Mastery Check

- [ ] Tôi biết composition root là gì.
- [ ] Tôi biết tránh global singleton.
- [ ] Tôi biết constructor injection đủ cho nhiều service Go.
