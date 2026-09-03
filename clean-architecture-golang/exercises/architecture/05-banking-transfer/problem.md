# Bài 5: Thiết Kế Banking Transfer

## Requirement

A chuyển 500.000 VND cho B.

Phải xử lý:

- Balance invariant.
- Transaction.
- Row locking.
- Idempotency.
- Retry.
- Double spending.
- Transaction history.

## Nhiệm vụ

1. Thiết kế `Account`, `Money`, `Transfer`.
2. Thiết kế `TransferMoneyUseCase`.
3. Thiết kế repository port.
4. Đặt transaction boundary.
5. Phân tích Kafka event nếu cần publish `MoneyTransferred`.

## Câu hỏi

- Locking thuộc domain hay repository adapter?
- Idempotency thuộc layer nào?
- Clean Architecture giúp gì và không giúp gì?

## Constraints

PostgreSQL là source of truth, nhiều replicas, 5.000 TPS và có hot Account. HTTP client/Kafka producer đều có thể retry. Event `MoneyTransferred.v1` phải cuối cùng được publish, nhưng broker không nằm trong DB transaction.

Bạn phải công bố đang dùng mutable balance hay double-entry ledger. Nếu chọn balance cho V1, mô tả đường migration và giới hạn audit.

## Failure Injection

- A debit thành công trong memory nhưng B save lỗi.
- Hai withdraw 800.000 cùng đọc balance 1.000.000.
- A->B và B->A lock ngược thứ tự.
- COMMIT thành công nhưng response mất.
- Publish thành công nhưng outbox chưa mark; event bị duplicate.

## Deliverables

1. Domain code signatures và invariant table.
2. Runtime sequence cùng compile-time import graph.
3. So sánh ít nhất hai transaction patterns rồi chọn một.
4. SQL lock/version/idempotency/outbox schema.
5. Error/retry classification, bao gồm commit unknown.
6. Test matrix: property, race, database concurrency và crash windows.
7. Metrics/runbook cho lock wait, duplicate, outbox age, imbalance.

## Self-review

- Hai repository `Save` có thật sự cùng transaction không?
- Process mutex có bị hiểu nhầm là distributed lock không?
- Retry transaction có tạo transfer thứ hai không?
- HTTP/Kafka contract có dùng cùng domain use case mà không rò DTO không?
