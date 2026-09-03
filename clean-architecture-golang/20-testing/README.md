# Testing Clean Architecture: test đúng guarantee ở đúng boundary

Boundary tốt cho phép test rẻ, nhưng nhiều test không tự chứng minh architecture tốt. Một mock-heavy suite có thể xanh dù SQL sai, transaction không rollback và domain invariant bị bypass.

## Kết quả học tập

- chọn unit/contract/integration/E2E theo risk;
- viết domain/use-case/HTTP/Repository tests;
- phân biệt fake, stub, spy, mock bằng code;
- test transaction, concurrency, fuzz và architecture;
- tránh brittle/flaky test;
- thiết kế CI test lanes.

## 1. Test portfolio theo question

| Question | Test rẻ nhất đáng tin |
|---|---|
| Withdraw giữ invariant? | Domain unit |
| Use case save đúng outcome? | Application + fake |
| Adapter cùng semantics? | Contract suite |
| SQL/schema/lock chạy đúng? | PostgreSQL integration |
| HTTP map đúng? | httptest |
| Object graph hoạt động? | Composition smoke |
| User workflow qua process? | E2E |
| Import direction đúng? | Architecture fitness |

Không dùng pyramid như quota phần trăm. Blast radius và uncertainty quyết định coverage.

## 2. Domain tests không mock

~~~go
func TestAccountWithdrawAtOverdraftBoundary(t *testing.T) {
	account := mustAccount(t, 100, 50)
	amount := domain.MustMoney(150, "VND")

	if err := account.Withdraw(amount); err != nil {
		t.Fatal(err)
	}
	if got := account.Balance().Amount(); got != -50 {
		t.Fatalf("balance=%d", got)
	}
}
~~~

Matrix:

- happy;
- insufficient;
- exact overdraft;
- zero/negative;
- currency mismatch;
- frozen;
- overflow;
- constructor/rehydration invalid.

Test observable behavior, không truy cập private field bằng reflection.

## 3. Table-driven boundary tests

~~~go
tests := []struct {
	name    string
	amount  int64
	wantErr error
}{
	{"zero", 0, domain.ErrInvalidAmount},
	{"negative", -1, domain.ErrInvalidAmount},
	{"too much", 151, domain.ErrInsufficientBalance},
}
~~~

Tên case diễn đạt rule. Không nhét toàn test suite vào một table khổng lồ khó debug.

## 4. Use-case test

Fake Repository + spy Transactor:

~~~go
type fakeAccounts struct {
	accounts map[domain.AccountID]*domain.Account
	saved    []domain.AccountID
	err      error
}
~~~

Test:

- invalid command trước transaction;
- load IDs;
- domain behavior;
- no Save khi reject;
- save both inside tx;
- error propagation;
- dependency constructor.

Code chạy: [transfer_money_test.go](../examples/mini-banking/internal/account/application/transfer_money_test.go).

Fake không chứng minh atomic rollback; PostgreSQL integration làm việc đó.

## 5. Stub

Stub trả canned response, không quan tâm interaction:

~~~go
type rateStub struct{ rate decimal.Decimal }

func (s rateStub) Current(context.Context, CurrencyPair) (decimal.Decimal, error) {
	return s.rate, nil
}
~~~

Dùng khi test cần input gián tiếp ổn định.

## 6. Fake

Fake có working implementation đơn giản:

~~~go
type fakeRepository struct {
	mu       sync.Mutex
	accounts map[AccountID]*Account
}
~~~

Fake dễ đọc cho state-based tests, nhưng phải nói thật semantics. Pointer alias/Noop transaction có thể làm fake khác production.

## 7. Spy

Spy ghi interaction để assert:

~~~go
type publisherSpy struct {
	events []Event
}

func (s *publisherSpy) Publish(_ context.Context, event Event) error {
	s.events = append(s.events, event)
	return nil
}
~~~

Dùng khi interaction chính là output, ví dụ event được yêu cầu publish.

## 8. Mock

Mock đặt expectation trước:

~~~go
repo.EXPECT().FindByID(ctx, sourceID).Return(source, nil)
~~~

Hữu ích với protocol phức tạp/rare branches, nhưng sequence-heavy expectations dễ khóa implementation thay vì behavior.

Smell:

~~~text
EXPECT repo load
EXPECT logger info
EXPECT metric increment
EXPECT publisher
EXPECT clock twice
~~~

Refactor harmless làm test vỡ. Ưu tiên state/outcome; chỉ assert interaction có ý nghĩa contract.

## 9. Repository contract test

~~~go
func AccountRepositoryContract(t *testing.T, factory Factory) {
	t.Run("round trip", ...)
	t.Run("not found semantics", ...)
	t.Run("detached ownership", ...)
	t.Run("canceled context", ...)
}
~~~

Chạy cho memory và Postgres. Contract không chứa assumption riêng adapter như exact SQL.

Executable ở [Lab 03](../labs/lab-03-repository/solution/repositorytest/contract.go).

## 10. PostgreSQL integration

Real database test:

- apply migrations;
- isolate/truncate data;
- round-trip;
- constraints;
- rollback;
- row lock/concurrent transfer;
- SQLSTATE;
- timeout.

Mock SQL driver chỉ chứng minh code gọi một string kỳ vọng, không parser/planner/schema.

Testcontainers tăng portability nhưng Docker startup cost. CI service DB nhanh hơn nhưng isolation/cleanup cần chặt. Có thể chia lane.

## 11. HTTP test

httptest với fake use case kiểm:

- method/content type/body limit;
- malformed/unknown/trailing;
- command/context mapping;
- status/error schema;
- no internal leak;
- headers.

Không cần Postgres để test 409 mapping.

## 12. Kafka/Redis/external adapter tests

Kafka:

- duplicate/inbox;
- retry/DLQ classifier;
- offset/broker integration;
- old schema fixtures.

Redis:

- hit/miss/error;
- TTL fake clock;
- invalidation/stampede;
- integration cho Lua/expiry.

External HTTP:

- httptest.Server;
- timeout/retry/idempotency header;
- body close/limit;
- ambiguous status.

## 13. Transaction test

Use-case fake verifies closure use. Chỉ DB integration verifies rollback:

~~~text
save source success
inject receiver save failure
transaction returns error
reload source → unchanged
~~~

Mini-banking integration suite làm đúng scenario và opposite-transfer concurrency.

## 14. Concurrency và race

~~~bash
go test -race ./...
~~~

Race detector tìm unsynchronized memory access trong execution đã chạy, không chứng minh algorithm tránh lost update ở DB.

Deterministic concurrency test dùng barrier/channel để ép hai operations overlap. Có timeout tránh treo CI. Assert invariant/final effects, không dựa sleep.

## 15. Fuzz/property tests

Value Object parser/arithmetic:

~~~go
func FuzzCurrency(f *testing.F) {
	f.Add("VND")
	f.Fuzz(func(t *testing.T, raw string) {
		currency, err := domain.NewCurrency(raw)
		if err == nil && len(currency.String()) != 3 {
			t.Fatal("constructor returned invalid currency")
		}
	})
}
~~~

Properties: Add/Sub round trip khi không overflow, constructor never returns invalid state, decoder never panic.

## 16. Architecture fitness test

Parse imports chặn:

~~~text
domain → third-party/outer
application → delivery/infrastructure
~~~

Executable [dependency_test.go](../examples/mini-banking/internal/architecture/dependency_test.go). Fitness test bổ sung review; nó không đọc semantic coupling qua shared DTO.

## 17. E2E và smoke

E2E ít nhưng giá trị:

- migrations + process start;
- POST transfer;
- query history;
- duplicate idempotency;
- outbox/Kafka eventual effect;
- graceful shutdown.

E2E fail chậm và khó localize, nên không thay unit/integration.

## 18. Test data

Builder helper giữ defaults hợp lệ:

~~~go
func validAccount(t *testing.T, options ...AccountOption) *domain.Account
~~~

Không dùng magic fixture toàn cục mutable. Mỗi test sở hữu data. Clock/ID generator inject để deterministic.

## 19. Flakiness

Nguồn:

- sleep/time;
- shared DB rows;
- unordered map/goroutine;
- port collision;
- timezone;
- external network;
- leaked resources.

Fix root cause: fake clock/barrier/unique schema/ephemeral port/t.Cleanup. Không chỉ retry CI test.

## 20. Test error semantics

errors.Is/As thay string. Snapshot public JSON có thể hợp nếu schema contract; snapshot full logs/internal errors thường brittle.

## 21. Coverage

Coverage chỉ nói line executed. 100% có thể bỏ invariant combinations/concurrency. Dùng coverage để tìm vùng mù, không làm quality target duy nhất. Mutation testing có thể cho biết assertions có bắt behavior change không.

## 22. CI lanes

~~~text
PR fast: format, vet, unit, race selected, architecture
PR integration: PostgreSQL/Redis/Kafka contract
main/nightly: full race, fuzz budget, E2E, load/chaos
~~~

Không để integration skip âm thầm trong job bắt buộc. Script verify env/test actually ran.

## 23. Production scenario

Bug lost update pass toàn unit tests vì fake serial:

- race detector không thấy data race;
- mock SQL đúng expectation;
- chỉ concurrent PostgreSQL integration bắt final balance;
- production metric/reconciliation phát hiện residual.

Test portfolio phải bao phủ semantics của resource thật.

## 24. Debug test failure

1. xác định boundary/guarantee test hỏi;
2. reproduce deterministic seed;
3. xem fake có khác production;
4. kiểm shared state/time/order;
5. thu test logs/artifacts;
6. giảm về smallest failing layer;
7. không tăng sleep.

## 25. Khi nào không mock?

Không mock Value Object/Entity, standard library thuần, data struct. Interface chỉ để mock là smell nếu concrete dependency nhanh/deterministic. Mock ở volatile/I/O boundary.

## 26. Lab

Làm [Lab 07: Testing](../labs/lab-07-testing/README.md): domain matrix, use-case fake/spy, HTTP test, Repository contract và PostgreSQL integration plan.

## 27. Mastery questions

1. Fake khác stub/spy/mock ở behavior nào?
2. Vì sao mock SQL không đủ?
3. Race detector không bắt lost update DB vì sao?
4. Contract test nên/không nên chứa assumption gì?
5. E2E không thay unit test vì sao?
6. 100% coverage chưa chứng minh gì?
7. Test double overuse khóa implementation thế nào?
8. Integration skip trong CI cần quản ra sao?

## Further reading

- Go testing, fuzzing, race detector docs.
- Testcontainers for Go documentation.
- Martin Fowler, Test Double.
- Growing Object-Oriented Software, Guided by Tests, đọc và điều chỉnh cho Go idioms.

## Quality gate

- [x] Mỗi test level có code/question
- [x] Fake/stub/spy/mock code và trade-off
- [x] Domain/usecase/repo/HTTP/Kafka/Redis tests
- [x] Transaction/concurrency/fuzz/architecture/E2E
- [x] CI/flakiness/coverage/debug
- [x] Lab, mastery, references
