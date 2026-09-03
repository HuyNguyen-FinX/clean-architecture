# Transaction Management: atomicity, concurrency và failure thực tế

Transaction là nơi architecture gặp vật lý của database. Vẽ layer đúng không cứu được transfer nếu debit commit còn credit fail. Ngược lại, nhét sql.Tx vào domain làm core model dính driver mà vẫn chưa giải quyết retry, idempotency hay network ambiguity.

## Kết quả học tập

- đặt transaction boundary theo atomic business operation;
- implement closure-based Transactor bằng pgx;
- so sánh sáu transaction patterns;
- hiểu isolation, row lock, lost update, deadlock và optimistic version;
- tách local transaction khỏi network workflow;
- thiết kế retry đi cùng idempotency;
- test commit, rollback và concurrent transfer bằng PostgreSQL thật.

## 1. Problem: hai Save không tạo atomicity

Transfer 500.000 VND:

~~~text
A debit thành công
        ↓
Save(A) commit
        ↓
B credit thất bại
        ↓
hệ thống mất tiền
~~~

Atomic flow cần:

~~~sql
BEGIN;
SELECT account A FOR UPDATE;
SELECT account B FOR UPDATE;
UPDATE account A;
UPDATE account B;
INSERT INTO transfers (...);
COMMIT;
~~~

Nếu bất kỳ bước nào fail trước COMMIT, ROLLBACK phải đưa durable state về trước operation.

## 2. Ba level

### Level 1: trực giác

Transaction là hộp “tất cả hoặc không gì cả” cho thay đổi trong cùng transactional resource.

### Level 2: Backend Engineer

Phải quản:

- begin/commit/rollback và commit error;
- transaction handle truyền qua Repository;
- context deadline;
- lock order/isolation;
- retryable SQLSTATE;
- connection pool occupation;
- test rollback thật.

### Level 3: Architecture

Transaction boundary thường gần Application Use Case vì use case biết toàn bộ atomic business operation. Domain biết invariant của object; Repository biết database; chỉ application biết “load hai Account, tạo Transfer, lưu cả ba” là một unit.

~~~mermaid
flowchart TB
    UC["TransferMoney Use Case"]
    subgraph TX["Transaction Boundary"]
      LOAD["load A + B"]
      RULE["withdraw + deposit"]
      SAVE["save A + B + transfer"]
      LOAD --> RULE --> SAVE
    end
    UC --> TX
~~~

## 3. Vì sao Repository tự quản transaction không đủ?

~~~go
func (r *Repository) Save(ctx context.Context, a *Account) error {
	tx, err := r.pool.Begin(ctx)
	// update one row, commit
}
~~~

Repository chỉ thấy một Save. Nó không biết Save thứ hai và insert Transfer phải atomic. Mỗi method commit riêng làm application không thể tạo unit lớn hơn.

Repository-managed transaction vẫn hợp lý cho operation thật sự atomic trong một method, ví dụ single SQL update không tham gia workflow khác.

## 4. Consumer-facing Transactor

~~~go
type Transactor interface {
	WithinTransaction(
		ctx context.Context,
		fn func(txCtx context.Context) error,
	) error
}
~~~

Use case:

~~~go
return uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
	from, err := uc.accounts.FindByID(txCtx, fromID)
	if err != nil {
		return err
	}
	to, err := uc.accounts.FindByID(txCtx, toID)
	if err != nil {
		return err
	}
	if err := from.Withdraw(amount); err != nil {
		return err
	}
	if err := to.Deposit(amount); err != nil {
		return err
	}
	if err := uc.accounts.Save(txCtx, from); err != nil {
		return err
	}
	return uc.accounts.Save(txCtx, to)
})
~~~

Domain không biết transaction. Application biết capability, không biết pgx.Tx.

## 5. pgx implementation

Code chạy được: [transactor.go](../examples/mini-banking/internal/account/infrastructure/postgres/transactor.go).

~~~go
func (t *Transactor) WithinTransaction(
	ctx context.Context,
	fn func(context.Context) error,
) (err error) {
	tx, err := t.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			err = errors.Join(err, rollbackErr)
		}
	}()

	txCtx := context.WithValue(ctx, privateTxKey{}, tx)
	if err := fn(txCtx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
~~~

Rollback được defer cho mọi exit path, kể cả panic unwind. Sau commit thành công, Rollback trả ErrTxClosed và được bỏ qua. Commit có thể fail nên không được bỏ error.

Production code còn phải quyết định panic policy, rollback timeout riêng và nested transaction semantics.

## 6. Context-based transaction: tiện nhưng ẩn

Repository chọn tx từ context:

~~~go
if tx, ok := transactionFromContext(ctx); ok {
	return tx.Exec(ctx, query, args...)
}
return pool.Exec(ctx, query, args...)
~~~

Ưu điểm:

- port sạch khỏi pgx;
- nhiều Repository cùng transaction;
- call sites không thêm parameter driver;
- hợp closure style của Go.

Nhược điểm:

- dependency tx không hiện trong signature;
- context.Background làm query thoát transaction;
- wrong key/value chỉ fail runtime;
- nested call policy khó thấy;
- context vốn dành cho request-scoped cancellation/value, không phải generic bag.

Mitigation: private typed key, architecture tests, transaction integration tests, không tạo background context trong call chain, reject nested transaction rõ.

## 7. Sáu patterns và trade-off

| Pattern | Clarity | Coupling | Testability | Leakage | Phù hợp |
|---|---|---|---|---|---|
| Repository-managed tx | local rõ | thấp | dễ | che multi-step | single repo operation |
| Use case import sql/pgx Tx | flow rõ | application dính driver | khó hơn | cao | pragmatic small service |
| Transaction Manager closure | boundary rõ | port nhỏ | tốt | context có thể ẩn | nhiều Go service |
| Unit of Work | changed set rõ | abstraction lớn | tốt | custom model | complex aggregate graph |
| Explicit tx-aware repos | type-safe | nhiều interfaces | khá | signature ceremony | high-assurance flow |
| Callback nhận repositories | rất explicit | factory/UoW coupling | tốt | ít driver leak | multi-repo transaction |

Ví dụ explicit callback:

~~~go
type UnitOfWork interface {
	Within(ctx context.Context, fn func(Repositories) error) error
}

type Repositories interface {
	Accounts() AccountRepository
	Transfers() TransferRepository
}
~~~

Không có pattern thắng mọi tiêu chí. Chọn guarantee dễ hiểu nhất cho team và test được.

## 8. ACID không có nghĩa “khn không race”

Atomicity bảo toàn group write. Isolation quyết định concurrent transaction quan sát nhau thế nào. Read Committed vẫn có thể lost update nếu flow read-compute-write không lock/version.

Scenario:

~~~text
balance = 1.000.000
Tx A read 1.000.000
Tx B read 1.000.000
Tx A withdraw 800.000 → save 200.000
Tx B withdraw 800.000 → save 200.000
~~~

Hai withdrawal được chấp nhận, chỉ một effect hiện trong balance.

## 9. Pessimistic locking

~~~sql
SELECT id, balance_minor, currency, overdraft_limit_minor, status
FROM accounts
WHERE id = $1
FOR UPDATE;
~~~

Row lock giữ đến transaction end. Transaction khác muốn update/lock row phải chờ.

Ưu điểm:

- mental model đơn giản cho hot invariant;
- conflict xử lý trước mutation;
- phù hợp operation ngắn.

Chi phí:

- lock wait và pool occupancy;
- deadlock;
- throughput thấp với hot key;
- timeout tạo rollback work.

Mini-banking chỉ thêm FOR UPDATE khi Repository chạy trong transaction.

## 10. Deadlock và lock order

Hai transfer ngược chiều:

~~~text
Tx1 lock A → chờ B
Tx2 lock B → chờ A
~~~

Database phát hiện deadlock và abort một transaction. Giảm xác suất bằng cách lock resource theo thứ tự ổn định:

~~~go
firstID, secondID := fromID, toID
if string(firstID) > string(secondID) {
	firstID, secondID = secondID, firstID
}
first := repo.FindByID(txCtx, firstID)
second := repo.FindByID(txCtx, secondID)
~~~

Sau load, map lại sender/receiver theo identity. Stable order không xóa mọi deadlock trong hệ thống lớn, nhưng loại cycle phổ biến.

## 11. Optimistic concurrency

Schema:

~~~sql
ALTER TABLE accounts ADD COLUMN version BIGINT NOT NULL DEFAULT 0;
~~~

Update:

~~~sql
UPDATE accounts
SET balance_minor = $1, version = version + 1
WHERE id = $2 AND version = $3;
~~~

Zero rows affected nghĩa version đổi hoặc row mất. Adapter trả ErrConcurrentModification. Application có thể reload và retry toàn use case.

Optimistic phù hợp conflict hiếm. Với hot account, retry storm có thể tệ hơn lock. Version không thay idempotency: retry request sau commit vẫn là operation mới nếu không có key.

## 12. Isolation levels

### Read Committed

Mỗi statement thấy committed snapshot mới. Default phổ biến. Cần lock/version cho read-modify-write.

### Repeatable Read

Snapshot ổn định hơn; PostgreSQL có thể abort serialization anomaly trong write conflict. Caller phải xử lý retry.

### Serializable

Kết quả tương đương một serial order, nhưng transaction có thể fail với serialization error. “Serializable” không đồng nghĩa không bao giờ lỗi; retry là phần contract.

Chọn isolation theo anomaly phải ngăn, không theo cảm giác “cao nhất an toàn nhất”. Isolation cao tăng abort/contention cost.

## 13. Retry transaction

Chỉ retry error được phân loại retryable, ví dụ serialization/deadlock. Policy:

~~~go
for attempt := 0; attempt < maxAttempts; attempt++ {
	err := tx.WithinTransaction(ctx, operation)
	if !isRetryable(err) {
		return err
	}
	if err := sleepWithJitter(ctx, attempt); err != nil {
		return err
	}
}
return ErrRetryExhausted
~~~

Điều kiện:

- closure phải tái tạo Aggregate từ DB mỗi attempt;
- không reuse object đã mutate;
- không gửi email/Kafka/network side effect trong closure;
- có max attempts, jitter, metrics;
- context budget còn đủ.

## 14. Transaction không giải quyết idempotency

Commit thành công nhưng HTTP response mất. Client retry cùng request:

~~~text
attempt 1: commit, response lost
attempt 2: commit lần nữa
~~~

Cả hai transaction đều atomic, nhưng business effect bị nhân đôi. Idempotency key cần được persist atomically với effect:

~~~sql
BEGIN;
INSERT INTO idempotency_keys(key, status)
VALUES ($1, 'started')
ON CONFLICT DO NOTHING;
-- nếu key mới: thực hiện transfer
UPDATE idempotency_keys SET status='completed', response=$2 WHERE key=$1;
COMMIT;
~~~

Phải định nghĩa key scope, request hash, retention, concurrent duplicate và replay response.

## 15. Network call trong DB transaction

Nguy hiểm:

~~~text
BEGIN
UPDATE account
call external payment API
COMMIT
~~~

Vấn đề:

- giữ row lock/connection trong network latency;
- timeout không biết remote đã xử lý chưa;
- DB rollback không rollback remote;
- retry có thể charge hai lần;
- contention lan rộng.

Thay bằng local atomic state + asynchronous coordination:

~~~sql
BEGIN;
UPDATE accounts ...;
INSERT INTO transfers ...;
INSERT INTO outbox ...;
COMMIT;
~~~

Worker publish outbox; consumer idempotent; reconciliation sửa ambiguous state. Saga/compensation dùng khi workflow qua nhiều transactional boundary. Không gọi đó là một ACID transaction phân tán.

## 16. Outbox failure matrix

| Failure | State | Recovery |
|---|---|---|
| DB rollback | account/transfer/outbox đều không đổi | request retry |
| Commit OK, worker down | outbox pending | worker restart |
| Publish OK, mark fail | broker có duplicate | consumer idempotent |
| Kafka down | outbox tăng | retry/backoff/alert |
| Poison payload | không publish được | quarantine/DLQ + repair |

Outbox cho at-least-once delivery, không tự tạo exactly-once end-to-end.

## 17. Context và cancellation

Flow:

~~~text
HTTP request context
→ use case
→ WithinTransaction
→ Repository
→ pgx
~~~

Không dùng context.Background giữa flow. Domain method Account.Withdraw(ctx, amount) thường sai vì rule không phụ thuộc deadline.

Khi request context hết hạn đúng lúc commit, kết quả có thể ambiguous tùy thời điểm/driver/network. Idempotency và read-after-retry cần xử lý, không chỉ trả 504.

## 18. Testing strategy

### Unit test use case

Spy Transactor gắn marker context, fake Repository ghi nhận call. Test orchestration, invalid input trước tx và error propagation. Nó không chứng minh rollback.

### Integration test PostgreSQL

Mini-banking test:

- transfer commit cập nhật hai account;
- injected second-save failure rollback first update;
- opposite transfer dùng stable lock order;
- row mapping reject corrupt invariant.

Chạy:

~~~bash
TEST_DATABASE_URL=postgres://... go test -race ./internal/account/infrastructure/postgres
~~~

### Concurrency test

Dùng goroutine/barrier để tạo overlap, deadline để tránh treo CI, assert final state và accepted/rejected count. Không chỉ chạy cùng test nhiều lần rồi hy vọng race xuất hiện.

## 19. Production scenario 5.000 TPS

Hot merchant account:

- lock serializes writes;
- queue vượt deadline;
- retry tăng load;
- pool wait tăng;
- transfer success nhưng response timeout;
- Kafka delayed.

Các hướng thiết kế:

- ledger append-only thay vì mutate một balance row;
- partition workload theo account;
- reservation/available balance;
- asynchronous credit;
- idempotency durable;
- load shedding;
- reconciliation.

Đây là system design trade-off. Clean boundary giúp thay strategy, không tự làm hot key biến mất.

## 20. Failure investigation

### Balance lệch

1. audit transfer/idempotency records;
2. kiểm transaction bao trùm mọi write;
3. xem code path dùng pool thay tx;
4. tìm retries/duplicate requests;
5. kiểm manual writers;
6. chạy reconciliation từ ledger.

### Deadlock tăng

1. lấy PostgreSQL deadlock log;
2. dựng wait graph;
3. so lock order;
4. rút ngắn transaction;
5. kiểm query plan/index;
6. retry bounded.

### Pool exhausted

Đo acquire latency, active tx age, lock wait, network call trong closure và unclosed rows. Tăng MaxConns có thể chỉ đẩy bottleneck vào DB.

## 21. Khi nào không cần abstraction transaction?

Một single atomic SQL statement có thể gọi trực tiếp. CRUD insert một row không cần UnitOfWork framework. Nếu app chỉ PostgreSQL và nhỏ, application import database/sql đôi khi là trade-off chấp nhận được.

Đừng tạo Transactor nếu mọi use case chỉ gọi một Repository method vốn atomic và không có dấu hiệu mở rộng. Nhưng khi guarantee là money movement nhiều write, abstraction phải đủ rõ/test được.

## 22. Bài thực hành

Làm [Lab 08: Transaction](../labs/lab-08-transaction/README.md):

1. tái hiện partial write;
2. implement closure Transactor;
3. test rollback;
4. thêm stable lock order;
5. so sánh context tx với explicit repositories;
6. thiết kế idempotency/outbox extension.

## 23. Mastery questions

1. Vì sao transaction boundary không trùng Repository boundary?
2. Context-based tx che dependency nào?
3. Read Committed vẫn lost update ra sao?
4. Stable lock order giảm deadlock thế nào?
5. Retry transaction vì sao phải reload Aggregate?
6. Transaction và idempotency giải hai failure khác nhau nào?
7. Network call trong DB transaction tạo ambiguous state gì?
8. Outbox có exactly-once không?
9. Serializable yêu cầu caller xử lý gì?
10. Khi nào optimistic lock tệ hơn pessimistic lock?

## Further reading

- [PostgreSQL Transaction Isolation](https://www.postgresql.org/docs/current/transaction-iso.html).
- [PostgreSQL Explicit Locking](https://www.postgresql.org/docs/current/explicit-locking.html).
- [PostgreSQL Deadlocks](https://www.postgresql.org/docs/current/explicit-locking.html#LOCKING-DEADLOCKS).
- pgx Tx và pgxpool documentation.
- Martin Fowler về Unit of Work và Saga.

## Quality gate

- [x] Problem, mental model và ba level
- [x] Production-style Go Transactor
- [x] Sáu pattern và trade-off
- [x] Locking, isolation, deadlock, optimistic concurrency
- [x] Retry, idempotency, network call, outbox
- [x] Runtime/compile-time dependency
- [x] Tests, production failure và debugging
- [x] Khi không dùng, exercise, mastery, references
