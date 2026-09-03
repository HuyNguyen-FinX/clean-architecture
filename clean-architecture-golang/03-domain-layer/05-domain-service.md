# Domain Service

## Problem: Rule Không Thuộc Tự Nhiên Về Một Entity

Không phải mọi business rule đều là method của một Entity. Ví dụ tính transfer fee có thể phụ thuộc loại sender, receiver, channel và policy, nhưng không object nào tự nhiên sở hữu toàn bộ rule.

Ép vào `Account`:

```go
func (a *Account) CalculateTransferFee(
	receiver *Account,
	channel Channel,
) Money
```

có thể làm Account biết quá nhiều về workflow/other aggregate.

Domain Service là stateless domain operation khi behavior:

- Là business rule thật.
- Cần nhiều domain objects/values.
- Không thuộc tự nhiên về một Entity/Value Object cụ thể.
- Có thể chạy không cần infrastructure I/O.

## Ba Level

### Level 1: Trực Giác

Nếu rule là “động từ nghiệp vụ” nhưng không object nào là chủ ngữ tự nhiên, dùng Domain Service/function.

### Level 2: Backend Engineer

Trong Go, Domain Service có thể chỉ là function:

```go
func CalculateTransferFee(
	from CustomerTier,
	to BankCode,
	amount Money,
) (Money, error)
```

Hoặc struct chứa immutable policy/config domain:

```go
type TransferFeePolicy struct {
	internalRate BasisPoints
	externalRate BasisPoints
}
```

Không cần interface/base class chỉ để gọi nó là service.

### Level 3: Modeling

Domain Service phải giữ Ubiquitous Language và pure decision. Nếu nó load repository, mở transaction, gọi Kafka và map HTTP error, nó đã trở thành Application Service/God Service.

## Domain Service Vs Application Service

| Domain Service | Application Service / Use Case |
|---|---|
| Business calculation/decision | Orchestration actor goal |
| Nhận domain values/entities | Nhận command/context |
| Không I/O | Gọi repository/gateway ports |
| Không transaction lifecycle | Thường quyết định transaction boundary |
| Không biết idempotency storage/HTTP | Phối hợp idempotency/workflow |

Transfer money:

```text
Application Service:
  load sender/receiver
  begin transaction
  call sender.Withdraw
  call receiver.Deposit
  save

Domain Service (nếu có):
  calculate fee theo tier/channel
  decide transfer eligibility dựa trên domain facts đã có
```

## Wrong: `TransferService` Chứa Mọi Thứ

```go
type TransferService struct {
	db        *sql.DB
	publisher *kafka.Writer
	logger    *zap.Logger
}

func (s *TransferService) Transfer(ctx context.Context, req HTTPRequest) error
```

Tên “Service” không nói layer. Responsibilities ở đây gồm application orchestration, persistence, integration, transport DTO và observability. Đây không phải Domain Service.

Chuỗi hậu quả:

```text
rule tính fee nằm cạnh SQL/Kafka
        ↓
test rule cần technical doubles
        ↓
rule khó reuse ở quote/batch flow
        ↓
technology change chạm domain calculation
```

## Correct: Pure Policy + Orchestration

```go
type TransferFeePolicy struct {
	internal BasisPoints
	external BasisPoints
}

func (p TransferFeePolicy) Fee(
	amount Money,
	sameBank bool,
) (Money, error) {
	rate := p.external
	if sameBank {
		rate = p.internal
	}
	return rate.Apply(amount)
}
```

Application:

```go
fee, err := uc.fees.Fee(amount, sender.Bank() == receiver.Bank())
if err != nil {
	return err
}
if err := sender.WithdrawTotal(amount, fee); err != nil {
	return err
}
```

Domain policy không load Account. Application cung cấp facts/objects đã load.

## External Data Trong Domain Rule

Eligibility cần credit score từ provider. Có hai bước:

```text
Application gọi CreditScoringGateway
        ↓
adapter map provider response thành RiskScore domain value
        ↓
Domain policy quyết định Eligible/Rejected
```

Không đưa SDK/provider response vào Domain Service. External score acquisition là I/O; decision từ normalized score có thể là domain.

Một interface `RateProvider` trong domain có thể hợp lý nếu abstraction là domain concept, nhưng nếu implementation luôn remote và cần context/retry, thường application gọi port rồi truyền `Rate` vào pure domain operation rõ hơn.

## Domain Service Có Được Mutate Entity?

Có thể, nhưng cần ownership rõ. Service có thể gọi public behavior trên Entities; tránh sửa private state bằng backdoor trong cùng package.

```go
func ApplySettlement(account *Account, amount Money) error {
	return account.Deposit(amount)
}
```

Function trên không thêm rule và chỉ delegate, nên không có giá trị. Đặt behavior trực tiếp ở Entity.

Chỉ dùng service khi nó thực sự phối hợp một domain rule không có owner tự nhiên.

## `context.Context` Có Thuộc Domain Service?

Pure calculation không cần context. Nếu service nhận context vì gọi repository/network, rất có thể nó là Application Service hoặc domain abstraction đang trộn I/O.

Computation dài có thể cần cancellation, nhưng hãy dùng có chủ đích và test semantics. Không dùng context để lấy logger/repository/config.

## Testing

Domain Service test giống pure function test:

- Table-driven cases theo policy matrix.
- Boundary rates/amounts.
- Currency/rounding errors.
- Không mock database/Kafka.
- Property tests khi calculation có invariant phù hợp.

Nếu test cần mock năm dependency, service có lẽ không còn là domain service thuần.

## Khi Nào Không Nên Dùng?

- Behavior thuộc rõ một Entity/Value Object.
- Function chỉ delegate.
- “Service” được tạo để tránh viết method.
- Domain đơn giản/procedural flow rõ hơn.
- Operation chủ yếu là I/O orchestration, nên thuộc application.

Domain Service không phải thùng chứa mọi logic không biết đặt đâu.

## Mastery Questions

1. Điều gì làm một service là Domain Service thay vì tên package?
2. Vì sao `TransferMoneyUseCase` là Application Service?
3. Fee calculation nên nằm ở Account hay service? Dữ kiện nào quyết định?
4. External credit score được lấy và dùng qua boundary nào?
5. Domain Service có nên nhận repository không?
6. Khi nào một function đơn giản tốt hơn struct/interface service?

