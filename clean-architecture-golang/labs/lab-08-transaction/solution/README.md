# Solution

Solution mong muốn dùng một boundary rõ:

```go
err := tx.WithinTransaction(ctx, func(txCtx context.Context) error {
	// load for update
	// withdraw/deposit
	// save both
	return nil
})
```

PostgreSQL adapter quyết định cách bind `txCtx` với transaction cụ thể.
