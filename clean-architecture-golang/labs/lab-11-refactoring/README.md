# Lab 11: Refactoring

## Mục tiêu

Refactor một handler xấu chứa HTTP, SQL, validation, business rule và Kafka thành boundary rõ hơn.

## Yêu cầu

- Liệt kê vấn đề.
- Xác định business logic.
- Tách domain.
- Tách use case.
- Tách repository/gateway port.
- Giữ behavior bằng test.

## Câu hỏi

- Refactor step nào nên làm trước?
- Làm sao tránh rewrite toàn bộ?
- Test nào cần viết trước khi di chuyển logic?

## Mastery Check

- [ ] Tôi biết refactor theo lát cắt use case.
- [ ] Tôi biết boundary thực sự nằm ở dependency, không chỉ folder.
- [ ] Tôi biết giữ behavior cũ khi tách layer.
