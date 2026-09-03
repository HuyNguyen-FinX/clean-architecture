# Case Study 04: Banking Account - Invariant, Atomicity Và Ledger

Banking transfer là ví dụ xuyên suốt curriculum. Case study này nối domain `Money`/`Account` với transaction, locking, idempotency, history và outbox. Code chạy được ở [mini-banking](../../examples/mini-banking/README.md) và capstone [Lab 12](../../labs/lab-12-full-application/README.md).

## Yêu Cầu Và Assumption

- Chuyển tiền nội bộ giữa hai Account cùng currency.
- Balance không thấp hơn `-overdraftLimit`.
- Cùng idempotency key và cùng request phải trả cùng kết quả.
- History chỉ xuất hiện khi debit/credit đã commit.
- Event `MoneyTransferred` cuối cùng phải được publish dù Kafka tạm down.
- PostgreSQL là authoritative store; traffic mục tiêu 5.000 TPS có hot accounts.

Nếu đây là hệ thống kế toán regulated, mutable balance không đủ làm source of truth. Cần double-entry ledger, reconciliation và audit controls. Ví dụ học tập dùng Account balance để làm rõ boundary rồi chỉ ra đường nâng cấp.

## Domain

`Money` là Value Object: amount minor unit + Currency, equality theo value, không cho amount âm ở transfer, kiểm tra overflow/currency mismatch. `Account` là Entity theo `AccountID`; balance thay đổi nhưng identity không đổi.

~~~go
func (a *Account) Withdraw(amount Money) error {
	if a.frozen {
		return ErrAccountFrozen
	}
	next, err := a.balance.Subtract(amount)
	if err != nil {
		return err
	}
	minimum := -a.overdraftLimit.Amount()
	if next.Amount() < minimum {
		return ErrInsufficientBalance
	}
	a.balance = next
	return nil
}
~~~

Domain bảo vệ invariant trên một Account. Atomicity xuyên Account A, Account B, Transfer record và outbox thuộc application transaction boundary. Row lock là persistence mechanism, không phải method domain.

## Application Boundary

~~~go
type UnitOfWork interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context, repos Repositories) error) error
}

type Repositories interface {
	Accounts() AccountRepository
	Transfers() TransferRepository
	Idempotency() IdempotencyRepository
	Outbox() OutboxRepository
}
~~~

Một closure transaction biểu diễn đúng atomic operation. Variant context-based nhỏ gọn hơn nhưng transaction bị giấu trong `context`; explicit scoped repositories rõ capability hơn và khó vô tình dùng non-transactional handle.

~~~mermaid
flowchart LR
    HTTP["HTTP transfer"] --> UC["TransferMoney"]
    KCON["Kafka command consumer"] --> UC
    UC --> MONEY["Money + Account"]
    UC --> UOW["UnitOfWork port"]
    PG["PostgreSQL UoW"] -.implements.-> UOW
    OUT["Outbox relay"] --> KAFKA["Kafka"]
~~~

Compile-time: adapters import application/domain; application không import pgx/Kafka. Runtime: use case gọi concrete PostgreSQL qua port do composition root wire.

## Transaction Flow

~~~text
BEGIN
  claim/check idempotency key + request hash
  SELECT accounts WHERE id IN (A,B) ORDER BY id FOR UPDATE
  A.Withdraw(amount)
  B.Deposit(amount)
  UPDATE A; UPDATE B
  INSERT transfer
  INSERT outbox
  store idempotent response
COMMIT
~~~

Lock theo thứ tự ID ổn định giảm deadlock khi hai transfer A->B và B->A chạy đồng thời. Nó không loại bỏ mọi deadlock; adapter vẫn phải phân loại SQLSTATE và retry toàn transaction với backoff/jitter trong giới hạn.

Repository không tự commit từng `Save`, vì khi B save thất bại thì A phải rollback. Transaction boundary cần component biết toàn use case.

## Idempotency

Khóa phải có scope, ví dụ `(client_id, operation, key)`. Lưu hash của canonical request. Ba trường hợp:

- Chưa có: claim và xử lý.
- Cùng key/cùng hash đã complete: replay response cũ.
- Cùng key/khác hash: conflict, không thực thi.

Nếu commit thành công nhưng HTTP response mất, retry không tạo transfer thứ hai. In-memory mutex không đủ qua restart/nhiều replica; unique constraint và record bền mới là arbitration point.

## Ledger Production

Ở hệ thống tài chính thật, chuyển 500.000 VND tạo transaction header và ít nhất hai entries có tổng bằng zero:

~~~text
Debit  Account A available balance  500000 VND
Credit Account B available balance  500000 VND
~~~

Entries append-only; balance có thể là projection/cache được cập nhật atomic hoặc rebuild từ ledger. Invariant mạnh hơn là tổng debit = tổng credit theo currency. Correction dùng reversal entry, không sửa lịch sử.

Đây là thay đổi domain model, không chỉ đổi bảng. Repository boundary nên phản ánh posting intent thay vì generic CRUD entries.

## Concurrency Và Isolation

Hai withdraw 800.000 từ balance 1.000.000 có thể cùng đọc giá trị cũ. `SELECT FOR UPDATE` serializes writers trên Account row; optimistic version cũng dùng được khi conflict hiếm. Serializable isolation có thể phát hiện anomaly nhưng yêu cầu retry transaction.

Hot account tạo queue ở database. Các lựa chọn gồm partition business workload, per-account sequencing, limits, hoặc ledger architecture tối ưu posting. Không dùng process mutex vì nhiều replica không chia sẻ memory.

## Failure Matrix

| Failure | Kết quả |
|---|---|
| A debit rồi B update lỗi | toàn transaction rollback |
| Deadlock SQLSTATE `40P01` | retry toàn transaction, cùng idempotency key |
| Client cancel trước COMMIT | query bị cancel; xác minh outcome nếu connection lỗi đúng lúc commit |
| COMMIT thành công, response mất | replay từ idempotency record |
| Kafka down | transfer vẫn bền, outbox backlog tăng |
| Publish xong, worker chết trước mark | publish lại; consumer deduplicate |
| Một account frozen | domain từ chối trước save |
| Currency mismatch | domain error; không mở side effect ngoài transaction |

PostgreSQL có commit-unknown case khi connection rơi đúng lúc commit. Adapter/application không nên retry với key mới; query idempotency/transfer ID để xác định.

## API Và Error Mapping

HTTP adapter giới hạn body, reject unknown fields, parse amount/currency rồi tạo command. `ErrInsufficientBalance`, `ErrAccountFrozen`, `ErrCurrencyMismatch`, conflict và unavailable được map thành contract public ổn định. Internal SQL text không được trả ra ngoài.

`context.Context` đi HTTP -> application -> repository để deadline/cancellation dừng I/O. `Account.Withdraw(ctx, ...)` là sai boundary vì invariant không phụ thuộc lifecycle request.

## Testing Strategy

- Domain matrix: exact overdraft boundary, insufficient, zero/negative, currency, overflow, frozen.
- Use-case fake/UoW: rollback snapshot, save failure, idempotent replay, hash conflict, outbox intent.
- Repository contract test cho memory và PostgreSQL adapters để semantics giống nhau.
- PostgreSQL integration: schema constraints, commit/rollback, opposite-direction transfer, deadlock/retry.
- HTTP test: strict JSON, safe error, request ID.
- Outbox test: crash trước publish, sau publish/trước mark, poison event.
- Property test: tổng tiền hai Account không đổi sau mọi transfer hợp lệ.

## Observability Và Operations

Metrics: transfer outcome/latency, DB lock wait, transaction retry, idempotent replay, outbox oldest age. Trace nối request -> transaction -> outbox publish bằng IDs trong baggage/log, nhưng account ID không làm metric label. Audit ghi actor, command ID, transfer ID, entries và timestamp đáng tin cậy.

Runbook cần query transfer theo idempotency key, kiểm tra ledger imbalance, replay outbox an toàn và khóa operation khi nghi duplicate. Dashboard HTTP 2xx không đủ chứng minh tiền đúng.

## Trade-off Và Giới Hạn

- Memory adapter là công cụ học/test, không có durability hoặc distributed concurrency.
- Row locking dễ hiểu nhưng throughput hot key hữu hạn.
- Optimistic locking tốt khi conflict thấp; conflict cao gây retry storm.
- Outbox cho atomic DB+intent, không cho exactly-once end-to-end.
- Clean Architecture làm policy độc lập; nó không tự tạo ACID, chống fraud hay đáp ứng regulation.

## Câu Hỏi Mastery

1. Vì sao hai `Save` thành công riêng lẻ vẫn chưa tạo transfer đúng?
2. Lock order giải quyết gì và không giải quyết gì?
3. Tại sao idempotency và transaction là hai cơ chế khác nhau?
4. Mutable balance khác double-entry ledger về khả năng audit ra sao?
5. Boundary nào biết `409`, boundary nào biết SQLSTATE `40P01`?

## Bài Thực Hành

Chạy test race và PostgreSQL integration của mini-banking. Sau đó thiết kế migration từ mutable Transfer history sang double-entry ledger mà API transfer không đổi; ghi rõ dual-write sẽ tránh bằng cách nào.
