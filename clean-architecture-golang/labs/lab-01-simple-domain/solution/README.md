# Solution

Solution minh họa một thiết kế idiomatic Go, không phải framework domain bắt buộc.

## Public API

```text
NewMoney / NewPositiveMoney
Money.Add / Sub / Negate / LessThan / Equal

NewAccount
Account.Deposit / Withdraw
Account.Freeze / Activate
Account.Balance / Status
```

## Tại Sao Thiết Kế Này Bảo Vệ Rule Tốt Hơn?

```text
Starter:
caller giữ primitives và sửa Account.Balance trực tiếp
        ↓
rule phụ thuộc discipline của từng caller

Solution:
state private + constructor + behavior methods
        ↓
mọi creation/transition đi qua cùng invariant boundary
```

`Money` giữ amount/currency cùng nhau. `Add`/`Sub` trả value mới, không mutate operands. `Account` là Entity nên methods chuyển state dùng pointer receiver.

`Withdraw` tính candidate balance, validate rồi mới assign. Khi error, Account giữ nguyên state.

`NewAccount` không chỉ điền fields: nó reject empty ID, invalid Money, currency mismatch, negative overdraft và initial balance dưới limit.

## Chạy

```bash
go test -race ./... -v
```

## Giới Hạn Có Chủ Đích

- Money dùng `int64` nhưng solution lab chưa implement checked overflow; full mini-banking có.
- Currency chỉ validate shape, không validate danh sách market hỗ trợ.
- Không có repository, transaction hoặc concurrent persistence.
- Không có Domain Event.

Các giới hạn này giữ lab tập trung vào Value Object, Entity và Invariant. Production guarantee sẽ được thêm theo từng chapter.
