# Solution

Solution mong muốn có `TransferMoneyUseCase` phụ thuộc vào `AccountRepository` interface, không phụ thuộc HTTP hoặc database driver.

Flow đúng:

```text
validate command
load sender
load receiver
sender.Withdraw(amount)
receiver.Deposit(amount)
save sender
save receiver
```

Business invariant vẫn nằm trong domain entity.
