# Database Adapter: bảo vệ domain mà không coi nhẹ dữ liệu

Database là infrastructure detail về công nghệ, nhưng dữ liệu và consistency thường là tài sản cốt lõi. Clean Architecture không yêu cầu giả vờ PostgreSQL có thể thay bằng file trong một buổi chiều; nó yêu cầu policy không bị driver/schema chi phối ngoài ý muốn.

## Kết quả học tập

- phân biệt domain model, persistence model và transport DTO;
- dùng pgxpool đúng ownership;
- map row qua rehydration và xử lý corrupt state;
- thiết kế schema/constraint bổ trợ invariant;
- quản migration, query, context và pool;
- viết integration test với PostgreSQL thật;
- biết khi nào direct SQL đơn giản hơn Repository.

## 1. Problem

Dùng một public struct cho cả database, JSON và domain có vẻ tiện:

~~~go
type Account struct {
	ID      string
	Balance int64
	Status  *string
}
~~~

Một struct giờ phải thỏa ba model:

- JSON naming/nullability;
- database encoding;
- domain invariant.

API đổi field có thể chạm SQL; nullable DB làm domain chấp nhận nil; mass assignment có thể mutate field không được phép. Reuse không luôn sai, nhưng coupling phải được nhận diện.

## 2. Ba level

### Level 1: trực giác

Database row là cách lưu object, không nhất thiết là object business. Adapter dịch hai phía.

### Level 2: Backend Engineer

Adapter chịu trách nhiệm:

- query parameter hóa;
- scan type/null;
- map driver error;
- deadline/cancel;
- pool và transaction handle;
- migration/constraint/index;
- integration test.

### Level 3: Architecture

Schema và domain là hai model tối ưu cho hai lực khác nhau. Domain tối ưu cho behavior/invariant; schema tối ưu cho durability, query, concurrency và evolution. Mapping là chi phí để hai model tiến hóa có kiểm soát.

## 3. Flow và dependency

~~~mermaid
flowchart LR
    UC["Use Case"] --> PORT["AccountRepository"]
    PG["Postgres Repository"] -.implements.-> PORT
    PG --> DOMAIN["Domain"]
    PG --> PGX["pgx/pgxpool"]
    PGX --> DB[("PostgreSQL")]
~~~

Application không import pgx. Adapter import application/domain. Composition root tạo pool và inject adapter.

Implementation chạy được:

- [repository.go](../examples/mini-banking/internal/account/infrastructure/postgres/repository.go)
- [row.go](../examples/mini-banking/internal/account/infrastructure/postgres/row.go)
- [migration](../examples/mini-banking/internal/account/infrastructure/postgres/migrations/001_accounts.sql)
- [integration test](../examples/mini-banking/internal/account/infrastructure/postgres/integration_test.go)

## 4. pgxpool là resource, không phải Repository

~~~go
pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
if err != nil {
	return err
}
if err := pool.Ping(ctx); err != nil {
	pool.Close()
	return err
}
defer pool.Close()
~~~

New có thể chỉ parse/configure; Ping kiểm connectivity startup. Pool được share và concurrent-safe. Không mở pool mỗi request, không Close trong Repository method.

Production config cần cân nhắc:

- MaxConns theo database capacity và replica count;
- MinConns/health check;
- MaxConnLifetime để rotate connection;
- acquisition timeout qua context;
- pool wait metrics.

Nếu 20 replica mỗi replica MaxConns=50, database có thể nhận 1.000 connection. Tối ưu cục bộ có thể làm hệ thống quá tải toàn cục.

## 5. Persistence model và mapping

~~~go
type accountRow struct {
	id             string
	balance        int64
	currency       string
	overdraftLimit int64
	status         string
}
~~~

Row private trong adapter. Mapping:

~~~go
func (row accountRow) toDomain() (*domain.Account, error) {
	id, err := domain.NewAccountID(row.id)
	if err != nil {
		return nil, fmt.Errorf("account id: %w", err)
	}
	balance, err := domain.NewMoney(row.balance, row.currency)
	if err != nil {
		return nil, fmt.Errorf("balance: %w", err)
	}
	overdraft, err := domain.NewMoney(row.overdraftLimit, row.currency)
	if err != nil {
		return nil, fmt.Errorf("overdraft: %w", err)
	}
	return domain.RehydrateAccount(id, balance, overdraft, domain.AccountStatus(row.status))
}
~~~

Rehydrate không bỏ invariant. Dữ liệu DB có thể sai vì migration cũ, manual operation, bug hoặc replica từ version khác. Trả lỗi corrupt-data và alert tốt hơn tạo Aggregate bất hợp lệ.

## 6. Schema bổ trợ domain invariant

~~~sql
CREATE TABLE accounts (
    id TEXT PRIMARY KEY,
    balance_minor BIGINT NOT NULL,
    currency CHAR(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    overdraft_limit_minor BIGINT NOT NULL CHECK (overdraft_limit_minor >= 0),
    status TEXT NOT NULL CHECK (status IN ('active', 'frozen')),
    CHECK (balance_minor >= -overdraft_limit_minor)
);
~~~

Invariant ở domain vẫn cần vì:

- fail sớm trước I/O;
- code dùng memory/other adapter vẫn đúng;
- error có business meaning;
- state transition test nhanh.

Constraint DB vẫn cần vì:

- nhiều writer hoặc script có thể bypass domain;
- concurrency/check xảy ra tại durable boundary;
- bug application không được persist corrupt state.

Defense in depth không phải duplication vô nghĩa khi hai boundary bảo vệ trước failure khác nhau.

## 7. Query và scan

~~~go
row := pool.QueryRow(ctx,
	"SELECT id, balance_minor, currency, overdraft_limit_minor, status "+
		"FROM accounts WHERE id = $1",
	id,
)
err := row.Scan(&stored.id, &stored.balance, &stored.currency,
	&stored.overdraftLimit, &stored.status)
~~~

Luôn:

- parameter hóa, không nối input vào SQL;
- liệt kê column thay vì SELECT *;
- scan theo thứ tự rõ;
- wrap operation và identifier an toàn;
- giữ errors.Is hoạt động;
- không log full query args chứa PII/secret.

Với list query, defer rows.Close và kiểm rows.Err. pgxpool giải phóng connection khi rows được close/đọc hết.

## 8. Null không tự là domain optional

sql.NullString hoặc pointer là persistence representation. Map explicit:

~~~go
var closedAt *time.Time
if row.closedAt.Valid {
	value := row.closedAt.Time.UTC()
	closedAt = &value
}
~~~

Nếu NULL nghĩa “unknown”, “not set” và “not applicable” khác nhau trong business, một pointer có thể quá nghèo. Domain Value Object/enum cần diễn đạt semantics.

## 9. Error mapping

| Driver/database error | Adapter semantics |
|---|---|
| pgx.ErrNoRows | ErrAccountNotFound |
| unique violation | typed conflict nếu consumer cần |
| check violation | corrupt input/concurrent invariant |
| serialization failure | retryable transaction error |
| context canceled/deadline | giữ errors.Is |
| connection failure | wrapped infrastructure error |

Không map HTTP status ở đây. Không map mọi error thành not found. Đừng string-match message nếu driver cung cấp SQLSTATE.

## 10. Migration là code production

Migration cần:

- version, review và CI từ DB rỗng;
- forward/backward compatibility khi rolling deploy;
- lock/time estimate cho table lớn;
- backup/rollback/recovery strategy;
- owner rõ, không chạy đồng thời ngẫu nhiên từ mọi replica.

Mini-banking cho phép AUTO_MIGRATE=1 để học. Production thường chạy migration bằng job/command riêng trước rollout.

### Expand and contract

Đổi column không nên deploy một nhát:

1. add nullable/new column;
2. deploy code dual-read/write nếu cần;
3. backfill có chunk;
4. enforce constraint;
5. chuyển read;
6. xóa old column ở release sau.

Schema migration là distributed systems problem khi nhiều app version cùng chạy.

## 11. Index theo query, không theo trực giác

Primary key hỗ trợ FindByID. History query có thể cần:

~~~sql
CREATE INDEX transfers_account_created_idx
ON transfers (account_id, created_at DESC, id DESC);
~~~

Dùng EXPLAIN (ANALYZE, BUFFERS) trên dữ liệu gần production. Index tăng write cost/storage và lock/migration risk. Không thêm index cho mọi column.

## 12. N+1 và Aggregate loading

Repository loop query child:

~~~text
1 query orders
N query items
~~~

Có thể sửa bằng join/batch query. Nhưng join tạo duplicate parent rows và mapping phức tạp. Đo query count/latency, chọn query theo access pattern. Read model projection thường tránh rehydrate Aggregate cho màn hình list.

## 13. Transaction-aware query

Repository mini-banking chọn pgx.Tx trong context khi có, pool khi không:

~~~go
if tx, ok := transactionFromContext(ctx); ok {
	row = tx.QueryRow(ctx, query+" FOR UPDATE", id)
} else {
	row = pool.QueryRow(ctx, query, id)
}
~~~

Điểm mạnh: application port không biết pgx. Điểm yếu: transaction dependency ẩn. Chương 11 phân tích pattern khác.

## 14. DTO vs Domain vs Row

Không bắt buộc ba struct cho mọi endpoint.

Reuse hợp lý:

- simple read projection;
- internal CRUD;
- same lifecycle/shape;
- không có private invariant.

Tách hợp lý:

- request cho phép field khác persistence;
- domain behavior/private state;
- version/null/storage metadata;
- backward compatibility;
- security boundary.

Đếm mapper lines không đủ để quyết định. Hãy đo change coupling và failure impact.

## 15. Integration test

Mock SQL chỉ test code gọi mock theo expectation; không chứng minh PostgreSQL hiểu query. Test thật cần:

- migration;
- round-trip mapping;
- constraints;
- not-found;
- commit/rollback;
- lock/concurrent transfer;
- context timeout;
- SQLSTATE mapping.

Mini-banking chạy suite khi có:

~~~bash
TEST_DATABASE_URL=postgres://... go test -race ./internal/account/infrastructure/postgres
~~~

Không có env, suite skip để unit tests vẫn nhanh. CI production nên có job bắt buộc cấp PostgreSQL; skip ở job đó phải được xem là failure bằng script/policy.

## 16. Production failure

### Pool cạn

Phân biệt connection acquisition time với query time. Tìm transaction dài, rows không close, MaxConns quá thấp hoặc database saturated.

### Query timeout

Context từ HTTP đi qua application tới adapter. Timeout trả về không bảo đảm server-side work đã dừng tức thì trong mọi dependency; xác minh driver cancellation và quan sát database.

### Replica lag

Ghi primary rồi đọc replica có thể không thấy dữ liệu. Read-after-write consistency là product guarantee, không chỉ routing detail.

### Commit thành công, response mất

Database đúng nhưng client không biết. Retry có thể ghi lần hai; transaction không thay thế idempotency.

## 17. Debug checklist

1. trace request/use case/query;
2. xem pool acquired/idle/wait;
3. inspect pg_stat_activity và lock waits;
4. EXPLAIN query với parameters đại diện;
5. kiểm schema version;
6. kiểm context deadline;
7. kiểm errors.Is/SQLSTATE sau wrapping;
8. tái hiện trong integration test.

## 18. Khi nào direct database access phù hợp?

Reporting service, migration tool, tiny CRUD hoặc performance-specific path có thể gọi pgx trực tiếp ở application/delivery package nếu coupling được chấp nhận. Không tạo Repository chỉ forward QueryRow.

Nếu business logic nhỏ nhưng SQL là core complexity, query-centric architecture có thể chân thật hơn Aggregate façade.

## 19. Bài thực hành

Làm [Lab 04: PostgreSQL](../labs/lab-04-postgresql/README.md):

1. map account row;
2. thêm constraint;
3. map no rows;
4. test migration/round trip;
5. test corrupt row;
6. quan sát pool và query deadline.

## 20. Mastery questions

1. Vì sao database là detail nhưng data không “không quan trọng”?
2. Khi nào ba model là đáng giá?
3. Tại sao domain và DB cùng kiểm invariant?
4. Mock SQL không chứng minh điều gì?
5. Pool size nhân theo replica tạo failure nào?
6. Vì sao auto-migrate từ mọi replica nguy hiểm?
7. Replica lag thuộc guarantee nào?
8. Context transaction ẩn tạo rủi ro gì?

## Further reading

- [pgxpool package](https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool).
- [PostgreSQL constraints](https://www.postgresql.org/docs/current/ddl-constraints.html).
- [PostgreSQL indexes](https://www.postgresql.org/docs/current/indexes.html).
- Go context documentation.

## Quality gate

- [x] Models, mapping và dependency analysis
- [x] Executable pgx adapter/migration/test
- [x] Pool, query, errors, null, index, migrations
- [x] Wrong/correct trade-off và direct access case
- [x] Production failures/debugging
- [x] Exercise, mastery và references
