# Dependency Injection trong Go: lắp Object Graph có chủ đích

Dependency Injection (DI) không phải framework và cũng không tự tạo Clean Architecture. DI là kỹ thuật đưa dependency từ bên ngoài vào object cần nó. Giá trị kiến trúc xuất hiện khi contract và dependency direction đã được thiết kế đúng.

## Kết quả học tập

- phân biệt DI, Dependency Inversion Principle (DIP) và Service Locator;
- lắp object graph thủ công bằng constructor injection;
- quản lý config, startup failure, ownership và shutdown;
- biết cách inject optional/cross-cutting dependency mà không tạo interface rỗng;
- đánh giá Wire/Fx theo quy mô graph;
- test wiring và phát hiện dependency cycle.

## 1. Problem

Một use case tự tìm dependency:

~~~go
func NewTransferMoney() *TransferMoney {
	url := os.Getenv("DATABASE_URL")
	pool, _ := pgxpool.New(context.Background(), url)
	return &TransferMoney{repo: postgres.NewRepository(pool)}
}
~~~

Hệ quả:

~~~text
use case tự đọc environment
        ↓
use case chọn PostgreSQL và quản connection lifecycle
        ↓
unit test cần environment/database
        ↓
không thể thay adapter tại composition boundary
        ↓
startup error xuất hiện muộn giữa request
~~~

Constructor injection làm dependency bắt buộc hiển thị:

~~~go
func NewTransferMoney(
	accounts AccountRepository,
	tx Transactor,
) *TransferMoney
~~~

## 2. Ba level

### Level 1: trực giác

Object không tự đi mua công cụ. Nơi lắp hệ thống đưa đúng công cụ cho nó. TransferMoney cần Repository và Transactor; nó không cần biết chúng được tạo bằng memory hay PostgreSQL.

### Level 2: Backend Engineer

DI giúp:

- thay fake trong unit test;
- validate config và connect dependency khi startup;
- đóng pool/server theo thứ tự;
- tránh singleton mutable;
- làm constructor trở thành executable documentation.

### Level 3: Architecture

DIP quyết định source dependency quay vào policy. DI là cơ chế runtime nối concrete adapter vào inward-facing port.

~~~mermaid
flowchart LR
    MAIN["cmd/api composition root"] --> PG["Postgres adapter"]
    MAIN --> UC["TransferMoney"]
    MAIN --> HTTP["HTTP Handler"]
    PG -.implements.-> PORT["Application port"]
    UC --> PORT
    HTTP --> UC
~~~

main import mọi phía là hợp lệ vì nó là điểm nối ngoài cùng, không chứa business policy.

## 3. Manual DI trước

Object graph tối thiểu:

~~~go
func main() {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	accounts := postgres.NewRepository(pool)
	tx := postgres.NewTransactor(pool)
	transfer := application.NewTransferMoneyUseCase(accounts, tx)
	handler := httpadapter.NewHandler(transfer)

	server := &http.Server{Addr: cfg.HTTPAddress, Handler: handler.Routes()}
	log.Fatal(server.ListenAndServe())
}
~~~

Full implementation có lựa chọn memory/PostgreSQL ở [main.go](../examples/mini-banking/cmd/api/main.go).

Thứ tự lắp graph kể một câu chuyện:

~~~text
config/resource → adapters → use cases → delivery → process lifecycle
~~~

Manual DI là mặc định tốt cho service nhỏ/vừa vì compiler kiểm type, control flow rõ và không cần container runtime.

## 4. Constructor contract

Dependency bắt buộc không nên âm thầm có default:

~~~go
func NewTransferMoneyUseCase(repo AccountRepository, tx Transactor) *TransferMoneyUseCase {
	if repo == nil {
		panic("application: nil account repository")
	}
	if tx == nil {
		panic("application: nil transactor")
	}
	return &TransferMoneyUseCase{accounts: repo, transactor: tx}
}
~~~

Panic ở composition time có thể chấp nhận cho programmer error. Với config/input runtime, trả error thường phù hợp hơn:

~~~go
func NewClient(cfg Config) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, errors.New("base URL is required")
	}
	return &Client{baseURL: cfg.BaseURL}, nil
}
~~~

Không tự thay nil Transactor bằng Noop. Caller có thể tưởng đang có atomicity. Nếu cung cấp Noop, đặt tên rõ và inject chủ đích như memory.NoopTransactor.

## 5. Config cũng là input boundary

Đọc env một lần ở outer layer:

~~~go
type Config struct {
	HTTPAddress    string
	DatabaseURL    string
	ShutdownTimeout time.Duration
}

func LoadConfig(getenv func(string) string) (Config, error) {
	// parse, default và validate cross-field constraints
}
~~~

Application service không gọi os.Getenv. Lợi ích:

- config invalid làm startup fail-fast;
- test parser bằng map function;
- secret không vô tình đi vào domain/log;
- ownership của default rõ.

Config struct khổng lồ truyền vào mọi constructor lại là Service Locator trá hình. Mỗi component chỉ nhận subset nó cần.

## 6. Lifecycle và ownership

Ai tạo resource thì thường chịu trách nhiệm đóng nó. pgxpool được tạo ở main nên main gọi Close. Repository dùng pool nhưng không đóng pool, vì nhiều adapter có thể share.

Startup production nên có phases:

1. load/validate config;
2. build logger/telemetry;
3. connect và ping critical dependencies;
4. build graph;
5. start listeners/workers;
6. nhận signal;
7. ngừng nhận work mới;
8. drain với timeout;
9. close resources.

Sai:

~~~go
func (r *Repository) Save(...) error {
	defer r.pool.Close()
	// pool chết sau request đầu
}
~~~

## 7. Interface và DI

Không tạo interface chỉ vì constructor injection:

~~~go
type ConfigProvider interface {
	Config() Config
}
~~~

Nếu consumer chỉ cần immutable Config value, truyền value. Nếu component concrete ổn định và test không cần thay, nhận pointer concrete cũng idiomatic.

Interface đáng có khi:

- consumer cần một capability nhỏ;
- có nhiều implementation thực hoặc test double có ý nghĩa;
- cần đảo dependency khỏi outer detail;
- contract mô tả behavior/guarantee, không chỉ mirror methods.

## 8. Cross-cutting dependencies

Logger có thể được inject vào handler/use case, nhưng domain Entity thường không cần logger. Instrument ở boundary:

~~~text
HTTP middleware: request duration/status
Application decorator: use-case result
Repository adapter: query latency/error
Kafka adapter: publish/consume
~~~

Một decorator giữ core constructor gọn:

~~~go
type instrumentedTransfer struct {
	next Transfer
	metrics Metrics
}

func (d instrumentedTransfer) Execute(ctx context.Context, cmd Command) error {
	start := time.Now()
	err := d.next.Execute(ctx, cmd)
	d.metrics.ObserveTransfer(time.Since(start), err)
	return err
}
~~~

Đừng inject Logger, Metrics, Tracer, Clock vào mọi Entity để chứng minh “observability”.

## 9. Multiple implementations và selection

Mini-banking chọn adapter ở composition root:

~~~go
if cfg.DatabaseURL == "" {
	repo = memory.NewRepository(seed...)
	tx = memory.NoopTransactor{}
} else {
	repo = postgres.NewRepository(pool)
	tx = postgres.NewTransactor(pool)
}
~~~

Selection là deployment policy, không phải business rule. Tuy vậy, hai adapter phải nói thật guarantee: memory Noop không rollback, PostgreSQL transactor có atomic commit/rollback.

Feature flag cũng nên chọn strategy lúc wiring nếu flag ít đổi. Nếu flag đổi theo request/user, application policy có thể cần một port rõ hơn.

## 10. Wire, Fx và code generation

| Cách | Ưu điểm | Chi phí |
|---|---|---|
| Manual | explicit, không magic, dễ debug | nhiều dòng khi graph lớn |
| Wire | generate code compile-time | thêm generation/tooling |
| Fx | lifecycle/module runtime mạnh | reflection/container concepts |
| Service Locator | gọi tiện ở mọi nơi | hidden dependency, runtime failure |

Wire/Fx đáng cân nhắc khi hàng chục module có graph/lifecycle lặp lại. Chúng không sửa interface sai hoặc package cycle. Bắt đầu manual giúp hiểu graph trước khi tự động hóa.

## 11. Testing wiring

Unit test constructor kiểm dependency invariant. Một smoke test composition có thể build graph bằng test config và gọi route:

~~~go
func TestBuildApplication(t *testing.T) {
	app, cleanup, err := build(testConfig(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	// exercise readiness or one route
}
~~~

Compiler bắt nhiều wiring error, nhưng không bắt:

- hai adapter cùng type bị nối nhầm;
- config nhầm topic/table;
- cleanup thiếu;
- router quên register;
- readiness báo healthy trước dependency.

## 12. Failure scenario

PostgreSQL down lúc startup:

- nếu DB là dependency bắt buộc, process fail-fast để orchestrator restart;
- nếu có degraded mode thực sự được thiết kế, readiness phải phản ánh capability;
- không fallback sang in-memory âm thầm, vì mất durability mà API vẫn báo thành công.

Kafka down có thể vẫn cho HTTP ready nếu outbox cho phép ghi durable. Quyết định readiness nằm ở production policy, được wiring/lifecycle thể hiện.

## 13. Debug graph

Khi request dùng nhầm adapter:

1. log component type/config đã redact lúc startup;
2. đọc composition root từ resource tới handler;
3. kiểm build tag/environment selection;
4. kiểm constructor nhận interface cùng shape nhưng khác semantics;
5. thêm smoke test assert concrete capability hoặc health endpoint.

Không debug bằng cách cho use case type-assert sang PostgresRepository. Việc đó phá boundary.

## 14. Khi nào không cần DI framework?

Phần lớn Go service có graph dưới vài chục constructor dùng manual DI tốt. Todo API ba component không cần container, provider set, module registry và reflection lifecycle. Ceremony lớn hơn variability mà nó xử lý.

## 15. Bài thực hành

Làm [Lab 06: Dependency Injection](../labs/lab-06-dependency-injection/README.md):

1. build graph memory và PostgreSQL từ config;
2. fail-fast khi config sai;
3. test cleanup ownership;
4. thêm graceful shutdown;
5. giải thích dependency direction từng import ở main.

## 16. Mastery questions

1. DI khác DIP thế nào?
2. Vì sao main được import cả delivery và infrastructure?
3. Tại sao fallback Noop âm thầm nguy hiểm?
4. Ai đóng shared connection pool?
5. Interface nào không cần tồn tại chỉ để inject?
6. Config struct toàn cục trở thành Service Locator ra sao?
7. Khi nào Fx/Wire hoàn vốn?
8. Startup test bổ sung gì ngoài compiler?

## Further reading

- Go blog và Effective Go về constructors, package design và errors.
- Mark Seemann, *Dependency Injection Principles, Practices, and Patterns*.
- Google Wire và Uber Fx documentation, đọc trade-off trước khi chọn.
- [Go module layout](https://go.dev/doc/modules/layout).

## Quality gate

- [x] Problem, mental model và ba level
- [x] Manual Go implementation
- [x] DIP/DI, compile-time/runtime analysis
- [x] Config, lifecycle, errors và testing
- [x] Framework trade-off và over-engineering
- [x] Production failure/debug scenario
- [x] Exercise, mastery và executable link
