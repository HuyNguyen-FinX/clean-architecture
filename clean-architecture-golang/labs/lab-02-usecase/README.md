# Lab 02: Tách TransferMoney Use Case

Thời lượng gợi ý: 60-90 phút. Lab này tiếp nối domain model ở Lab 01 và tập trung vào orchestration boundary.

## Mục tiêu

- tách transport/persistence detail khỏi business workflow;
- thiết kế **TransferCommand** và output vừa đủ;
- đặt Repository và Transactor port ở phía application consumer;
- giữ business invariant trong **Account**;
- unit test use case bằng fake/spy, không cần database.

## Kiến thức cần có

- [Domain Layer](../../03-domain-layer/README.md)
- [Use Case / Application Layer](../../04-usecase-application-layer/README.md)
- context.Context, table-driven test và errors.Is.

## Architecture diagram

~~~mermaid
flowchart LR
    INPUT["TransferCommand"] --> UC["TransferMoney.Execute"]
    UC --> ACCOUNT["Account behavior"]
    UC --> REPO["AccountRepository port"]
    UC --> TX["Transactor port"]
    FAKE["Test fake/spy"] -.implements.-> REPO
    FAKE -.implements.-> TX
~~~

Mũi tên liền là runtime call. Mũi tên chấm là implementation của contract. Application sở hữu contract mà test double implement.

## Problem

[Starter](./starter/) có TransferService giữ map, validate input, mutate balance và điều phối tất cả trong một method. Code chạy, nhưng:

- balance là primitive public;
- không có transaction boundary;
- storage ownership lẫn với use case;
- khó ép lỗi load/save để test failure path;
- API nhận nhiều primitive nên dễ đảo tham số.

Nhiệm vụ của bạn là refactor mà vẫn giữ observable behavior.

## Yêu cầu

1. Tạo domain.Account với private balance và Withdraw/Deposit.
2. Tạo application.TransferCommand.
3. Application định nghĩa interface nhỏ AccountRepository và Transactor.
4. Execute validate command trước khi mở transaction.
5. Trong transaction: load hai account, gọi domain behavior, save cả hai.
6. Domain rejection không được gọi Save.
7. Mọi dependency bắt buộc phải được kiểm tra ở constructor.
8. Test không import SQL driver hoặc adapter production.

## Các bước

### Bước 1: khóa behavior hiện tại

Chạy starter:

~~~bash
cd starter
go test ./...
~~~

Đọc test trước code. Ghi lại behavior nào là business rule, behavior nào chỉ là implementation detail.

### Bước 2: extract domain behavior

Đưa rule số tiền dương và insufficient balance vào Account. Không truyền context.Context vào Withdraw: rule này không phụ thuộc request lifecycle.

### Bước 3: thiết kế input boundary

Dùng command:

~~~go
type TransferCommand struct {
	From   domain.AccountID
	To     domain.AccountID
	Amount int64
}
~~~

Command validation bảo vệ shape của use-case input; domain vẫn tự bảo vệ invariant khi bị gọi từ entry point khác.

### Bước 4: invert dependencies

Định nghĩa port cạnh use case. Không tạo method ngoài nhu cầu hiện tại. Fake repository trong test implement port bằng structural typing.

### Bước 5: đặt transaction boundary

Closure phải bao trọn load, mutate và save. Dùng context mà Transactor truyền vào closure cho mọi repository call.

### Bước 6: test failure paths

Thêm test cho invalid command, account không tồn tại, insufficient balance, save error và transaction error. Assert observable behavior, tránh khóa cứng thứ tự mọi helper call nếu thứ tự đó không phải contract.

## Expected behavior

Với A=1.000 và B=100, transfer 300 cho kết quả A=700, B=400. Transfer 1.200 bị từ chối và không account nào được save. From trùng To, amount bằng 0 hoặc âm bị reject trước transaction.

## Test

~~~bash
cd solution
go test -race ./...
go vet ./...
~~~

Race detector quan trọng vì fake store cũng phải có ownership rõ nếu test chạy song song.

## Questions

1. Tại sao TransferMoney là Application Service chứ không phải Domain Service?
2. Nếu Repository tự mở transaction trong từng Save, atomicity nào bị mất?
3. Tại sao port nằm trong package application là hợp lý ở lab này?
4. Điều gì xảy ra nếu closure bỏ transaction context và dùng context ban đầu?
5. Fake đang kiểm chứng SQL hay kiểm chứng orchestration?

## Challenge

- Thêm Currency và reject transfer khác currency.
- Thêm idempotency port với state started/completed.
- Cho Transactor retry một lỗi giả lập và chứng minh domain mutation không bị áp hai lần lên cùng pointer.
- Tách result query để trả balance mới mà không lộ domain.Account khỏi application.

## Solution explanation

[Solution](./solution/) đặt invariant trong domain.Account, orchestration và port trong application. Test fake giữ account seed, ghi nhận ID đã load/save và Transactor spy gắn marker vào context. Nhờ đó test chứng minh repository call thật sự nằm trong transaction closure mà không phụ thuộc concrete database.

Đọc solution sau khi đã thử ít nhất một thiết kế. So sánh guarantee và dependency direction, không chỉ so tên package.
