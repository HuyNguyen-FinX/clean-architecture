# Domain-Driven Design: modeling và strategic boundaries

DDD giải quyết domain complexity và shared language; Clean Architecture bảo vệ policy khỏi details. Chúng bổ sung nhưng không đồng nghĩa. Có Clean Architecture cho CRUD, và có DDD model trong modular monolith không dùng vòng tròn layers.

## Kết quả học tập

- phân biệt strategic và tactical DDD;
- xây Ubiquitous Language cùng domain experts;
- xác định Bounded Context/Context Map;
- dùng Entity, Value Object, Aggregate, Service, Event, Repository đúng lực;
- tránh shared model/anemic/aggregate lớn;
- chọn khi DDD không hoàn vốn.

## 1. Problem: một từ, nhiều nghĩa

“Account” trong:

- Core Banking: balance, overdraft, status;
- IAM: login identity;
- CRM: customer organization;
- Accounting: ledger account.

Ép một global Account model tạo hàng chục optional fields và team coupling. Bounded Context cho phép cùng từ có model khác theo language/context.

## 2. Ba level

### Level 1

Model code bằng ngôn ngữ nghiệp vụ; object bảo vệ rule.

### Level 2

Entity/Value Object/Aggregate/Repository/Domain Service tổ chức behavior và consistency.

### Level 3

Strategic DDD chọn core domain, bounded contexts, ownership và relationship. Đây thường có leverage lớn hơn việc tạo nhiều tactical class.

## 3. Ubiquitous Language

Language phải xuất hiện trong conversation, examples, code, test và event contract:

~~~text
Transfer is initiated
Funds are reserved
Transfer is posted
Transfer is reversed
~~~

“Update account record” quá kỹ thuật nếu business phân biệt reserve/post. Glossary không đủ; code behavior cần phản ánh state transitions.

Language có scope context. Đừng thống nhất giả tạo toàn enterprise.

## 4. Knowledge crunching

Model phát triển qua workshop, examples và contradictions:

- Khi nào frozen Account vẫn nhận deposit?
- Overdraft áp theo currency nào?
- Transfer success lúc balance đổi hay beneficiary nhận?
- Reversal có xóa transaction cũ không?

Viết Example Mapping/Gherkin/table cases để lộ rule. Domain expert và engineer cùng sửa language/model.

## 5. Subdomains

- Core: lợi thế/capability chính, đầu tư model sâu.
- Supporting: cần riêng nhưng không khác biệt, có thể đơn giản.
- Generic: auth/email/observability, mua/dùng commodity.

Không DDD tactical hóa mọi thứ như nhau. Tập trung nơi complexity/business value cao.

## 6. Bounded Context

Mini-banking có thể bắt đầu một Account context. Khi lớn:

~~~mermaid
flowchart LR
    ACCOUNT["Account Context"] -->|MoneyTransferred v1| NOTIFY["Notification Context"]
    ACCOUNT -->|Ledger entries| LEDGER["Ledger Context"]
    RISK["Risk Context"] -->|Decision| ACCOUNT
~~~

Mỗi context sở hữu model/data contract. Notification không đọc private accounts table.

Bounded Context không bắt buộc microservice. Packages/modules và schema ownership trong monolith có thể enforce trước.

## 7. Context Map

Relationships:

- Partnership: phối hợp hai team;
- Customer/Supplier;
- Conformist: downstream chấp nhận upstream model;
- Anti-Corruption Layer: translate/protect;
- Shared Kernel: share nhỏ, governance chặt;
- Open Host Service/Published Language;
- Separate Ways.

Tên pattern ít quan trọng hơn việc công bố dependency/team power/evolution.

## 8. Entity

Identity bền qua thay đổi:

~~~go
type Account struct {
	id      AccountID
	balance Money
	status  AccountStatus
}
~~~

Account hôm nay/ngày mai cùng identity dù balance đổi. Equality không chỉ compare fields.

## 9. Value Object

Money equality theo amount+currency, immutable concept:

~~~go
next, err := balance.Add(amount)
~~~

Primitive 500000 không nói currency/unit. Constructor ngăn invalid currency/amount class.

## 10. Aggregate

Aggregate là consistency boundary, không phải graph object tùy ý. Root kiểm mutation.

Account A, Account B và Transfer không nên luôn gom một Aggregate khổng lồ:

- mỗi Account có lifecycle/traffic riêng;
- load graph lớn;
- hot lock/contention;
- transfer cần application transaction across aggregates;
- cross-aggregate reference bằng identity.

Aggregate nhỏ theo invariants cần immediate consistency. Eventual rule không cần nhét vào cùng Aggregate.

## 11. Invariant và transaction

Trong Account:

~~~text
balance >= -overdraftLimit
~~~

Transfer use case cần atomic two-account write, nhưng mỗi Account vẫn bảo vệ invariant. Database lock/version ngăn concurrent stale decision. Domain model không tự giải quyết isolation.

## 12. Factory/Rehydration

Creation có business defaults/events; rehydration tái tạo persisted state nhưng vẫn validate.

~~~go
func OpenAccount(...) (*Account, []DomainEvent, error)
func RehydrateAccount(...) (*Account, error)
~~~

Không phát “AccountOpened” khi load từ DB.

## 13. Domain Service

Rule không thuộc tự nhiên một Entity/VO:

~~~go
type ExchangeService struct {
	rates ExchangeRatePolicy
}

func (s ExchangeService) Convert(m Money, to Currency) (Money, error)
~~~

Domain Service không orchestration DB/HTTP. Application Service load, transaction, call service, save/publish.

## 14. Domain Event

Fact quá khứ có identity/time:

~~~go
type MoneyTransferred struct {
	TransferID TransferID
	From       AccountID
	To         AccountID
	Amount     Money
}
~~~

Integration mapper quyết định public payload. Event collection/dispatch timing phải rõ: trước/after commit, outbox, handler failure.

## 15. Repository

Repository cung cấp Aggregate theo domain/application needs, không global CRUD. Transaction có thể bao nhiều repositories. Query side có projection riêng.

## 16. Specification/Policy

Rule composable/persistent query có thể thành Specification, nhưng closure đơn giản thường idiomatic Go hơn:

~~~go
type EligibilityPolicy interface {
	Evaluate(CustomerSnapshot) Decision
}
~~~

Không tạo class cho mỗi if. Abstraction cần language/reuse/variation thật.

## 17. Anemic vs over-rich

Anemic:

~~~go
account.Balance -= amount
~~~

Rule nằm service, dễ bypass.

Over-rich:

~~~text
Account loads DB, calls Kafka, sends email
~~~

Entity giữ behavior/invariant local. Application orchestration và ports giữ I/O.

## 18. Data ownership

Context sở hữu write model/schema. Cross-context:

- published API/event;
- replicated read model;
- anti-corruption adapter.

Direct cross-schema read có thể pragmatic trong monolith/reporting, nhưng là coupling cần owner/version/test, không vô hình.

## 19. Eventual consistency

Notification sau MoneyTransferred có thể delayed/duplicate. Product UI phải thể hiện pending/processed; reconciliation và idempotency là phần model vận hành.

Không dùng eventual consistency cho invariant tiền mà business yêu cầu immediate nếu chưa đổi product semantics.

## 20. Modeling transfer lifecycle

~~~text
Requested → Posted
          ↘ Rejected
Posted → Reversed
~~~

Không sửa/xóa lịch sử để “rollback” business. Reversal là operation/event mới. Technical DB rollback khác business compensation.

## 21. Testing DDD model

- examples thành domain tests;
- invariant boundary/property tests;
- state transition table;
- repository contract;
- context contract/event fixtures;
- process manager/saga failure scenarios.

Mock ít trong domain. Test language: “frozen source cannot withdraw” thay “service returns code 3”.

## 22. Production scenario

Risk service chậm trong transfer:

- nếu decision bắt buộc trước posting, workflow PendingRisk;
- không giữ Account DB transaction qua network;
- command/event correlation;
- timeout → Pending/manual/retry;
- risk response duplicate/out-of-order;
- state machine reject invalid transition.

DDD làm rõ states/language; Clean boundary đặt gateway/event adapters.

## 23. Debug model

1. reconstruct timeline từ commands/events/ledger;
2. xác định Bounded Context owner;
3. kiểm transition/invariant;
4. phân biệt technical rollback/business reversal;
5. kiểm integration mapping/schema version;
6. bổ sung example làm executable test.

## 24. Khi nào không dùng tactical DDD sâu?

CRUD admin, configuration catalogue, short-lived tool, domain ít rule: transaction script + clear packages có thể tốt. DDD workshop vẫn giúp language, nhưng Aggregate/Event/Repository proliferation không hoàn vốn.

## 25. Bài tập

1. Vẽ contexts Account/Ledger/Risk/Notification.
2. Model Transfer state machine và reversal.
3. Chọn Aggregate boundaries, giải thích contention.
4. Thiết kế ACL cho provider payment.
5. Tìm global Account model bị overloaded.

## 26. Mastery questions

1. DDD khác Clean Architecture?
2. Bounded Context có phải microservice?
3. Vì sao Aggregate lớn gây contention?
4. Domain Event khác integration event?
5. Reversal khác rollback?
6. Shared Kernel rủi ro gì?
7. Domain Service khác Application Service?
8. Khi nào transaction script tốt hơn rich model?

## Further reading

- Eric Evans, Domain-Driven Design.
- Vaughn Vernon, Implementing Domain-Driven Design.
- Domain-Driven Design Distilled.
- Context Mapping và EventStorming literature.
- [Domain Layer series](../03-domain-layer/README.md).

## Quality gate

- [x] Strategic + tactical DDD
- [x] Language/subdomain/context map
- [x] Entity/VO/Aggregate/Service/Event/Repository
- [x] Data ownership/eventual consistency
- [x] Transfer production model/debug/tests
- [x] Non-dogmatic trade-off/exercises/mastery
