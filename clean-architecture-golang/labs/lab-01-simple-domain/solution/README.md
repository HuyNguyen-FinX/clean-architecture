# Solution

Solution giữ fields của `Account` private và buộc mọi thay đổi balance đi qua method `Withdraw` hoặc `Deposit`.

Điểm chính:

- `Money` là Value Object, chịu trách nhiệm kiểm tra currency khi cộng trừ.
- `Account` là Entity vì có identity.
- Business invariant nằm trong domain method.
- Test domain không cần mock hoặc database.

So sánh với starter:

```text
Starter:
  caller có thể sửa account.Balance trực tiếp

Solution:
  caller chỉ đọc Balance()
  thay đổi balance phải đi qua Account.Withdraw hoặc Account.Deposit
```

Đây là bước nhỏ nhưng rất quan trọng: boundary đầu tiên không phải giữa HTTP và use case, mà nằm ngay bên trong domain object.
