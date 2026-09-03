# Solution

Solution chia thành hai package:

~~~text
domain/       Account, identity và invariant
application/  TransferMoney, command, repository/transactor ports
~~~

Test trong application dùng fake/spy và kiểm tra transaction context. Không có infrastructure package vì mục tiêu của lab là consumer boundary; Lab 03 sẽ thêm adapter.

~~~bash
go test -race ./...
go vet ./...
~~~
