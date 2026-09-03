# Lab 05: HTTP

## Mục tiêu

Tạo REST adapter cho use case mà không đưa HTTP concept vào application/domain.

## Yêu cầu

- Tạo request DTO.
- Map DTO sang use case command.
- Map domain/application error sang HTTP status.
- Test bằng `httptest`.

## Câu hỏi

- Handler có nên gọi repository trực tiếp không?
- Domain error có nên chứa status code không?
- Request DTO khác command ở điểm nào?

## Mastery Check

- [ ] Tôi biết handler chỉ xử lý transport.
- [ ] Tôi biết test endpoint không cần DB thật khi use case được fake.
- [ ] Tôi biết map error ở delivery layer.
