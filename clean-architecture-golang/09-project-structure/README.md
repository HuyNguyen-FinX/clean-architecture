# 09 Project Structure

## Tại sao cần học?

Go project structure nên phản ánh boundary và ownership, không copy máy móc từ template. `cmd`, `internal`, `pkg` là công cụ, không phải kiến trúc.

## Package By Layer vs Package By Feature

Package by layer:

```text
handler/
service/
repository/
model/
```

Dễ bắt đầu nhưng domain lớn dễ bị trộn theo technical role.

Package by feature/domain:

```text
internal/account/domain
internal/account/application
internal/account/infrastructure
internal/account/delivery
```

Dễ giữ ownership theo bounded context hơn.

## `internal` và `pkg`

`internal` dùng cho code không muốn package ngoài module import. `pkg` chỉ nên dùng khi bạn thật sự muốn expose reusable library. Không tạo `pkg` chỉ vì template có.

## Anti-pattern

- Folder sâu nhưng package dependency sai.
- Shared package thành thùng rác.
- Đưa mọi helper vào `pkg/utils`.
- Chia layer toàn cục làm module domain phụ thuộc lẫn nhau không rõ ownership.

## Mastery Check

- [ ] Tôi biết structure phải phục vụ dependency boundary.
- [ ] Tôi biết khi nào dùng `internal`.
- [ ] Tôi tránh `pkg` khi chưa có public API rõ ràng.
