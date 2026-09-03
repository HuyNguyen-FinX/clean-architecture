# Solution

Solution mong muốn:

```go
repo := postgres.NewAccountRepository(pool)
uc := application.NewTransferMoneyUseCase(repo, tx)
handler := httpadapter.NewHandler(uc)
```

`main.go` là nơi lắp ráp. Core package vẫn không import adapter cụ thể.
