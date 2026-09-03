# Starter

Starter là baseline **compile và test được nhưng chưa bảo vệ domain**. Test `TestPublicBalanceCanBreakInvariant` cố ý chứng minh caller có thể tạo invalid state.

Chạy:

```bash
go test ./... -v
```

Sau đó:

1. Viết thêm behavior tests từ README của lab.
2. Refactor public primitives thành `Money` và private Account fields.
3. Thay free function `Withdraw` bằng Account behavior.
4. Xóa/thay test public mutation khi compiler không còn cho truy cập field.
5. Giữ mọi test hợp lệ pass.

Không import code từ `solution`. Hãy thiết kế public API trước, rồi dùng test/compiler dẫn đường.
