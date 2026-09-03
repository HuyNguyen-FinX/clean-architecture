# Value Object Và Equality

## Problem: `500000` Có Nghĩa Gì?

Primitive này không mang đủ domain meaning:

```go
func Transfer(fromID, toID string, amount int64) error
```

`amount` có thể là VND, USD cents, reward points hoặc một decimal đã scale. Nó có được âm không? Có được cộng với amount khác currency không? Overflow xử lý thế nào?

Primitive obsession buộc knowledge nằm trong đầu caller hoặc comment. Value Object đưa meaning và rule vào type.

## Ba Level

### Level 1: Trực Giác

Value Object được nhận diện bằng value, không bằng identity.

```text
Money(500.000, VND) == Money(500.000, VND)
```

Hai instance có thể thay thế nhau. Không cần hỏi “đây có phải cùng tờ tiền object không?”.

### Level 2: Backend Engineer

Value Object trong Go thường có:

- Fields unexported.
- Constructor validate/normalize.
- Value receiver cho operations không mutate.
- Explicit equality khi struct không thể/không nên dùng `==` trực tiếp.
- Error khi operation không hợp lệ.

### Level 3: Modeling

Value Object làm Ubiquitous Language executable:

- `Money` giữ amount/currency đi cùng nhau.
- `DateRange` bảo đảm start <= end.
- `Percentage` giới hạn range.
- `AccountID` tách identity types để tránh nhầm primitive.

Nó thu nhỏ state space: code ngoài không còn biểu diễn tùy ý các combination invalid.

## Implementation `Currency`

```go
type Currency string

func NewCurrency(raw string) (Currency, error) {
	normalized := strings.TrimSpace(strings.ToUpper(raw))
	if len(normalized) != 3 {
		return "", ErrInvalidCurrency
	}
	for _, char := range normalized {
		if char < 'A' || char > 'Z' {
			return "", ErrInvalidCurrency
		}
	}
	return Currency(normalized), nil
}
```

Constructor đang validate shape ba chữ cái, không khẳng định code có thật trong ISO 4217 hoặc business hỗ trợ currency đó. Hai policy khác nhau:

- Syntax/normalization ổn định có thể nằm trong Value Object.
- Danh sách currency được sản phẩm hỗ trợ có thể là domain policy/config thay đổi theo market.

Nhồi dynamic config vào constructor sẽ làm domain cần I/O/global state. Một factory/application policy có thể kiểm supported currencies trước khi tạo command/domain value.

## Implementation `Money`

```go
type Money struct {
	amount   int64
	currency Currency
}

func NewMoney(amount int64, currency string) (Money, error) {
	parsed, err := NewCurrency(currency)
	if err != nil {
		return Money{}, err
	}
	return Money{amount: amount, currency: parsed}, nil
}
```

`Money` dùng `int64` minor unit. Ví dụ VND không có phần lẻ trong flow này; USD amount có thể là cents nếu contract quy định rõ.

Không dùng `float64` cho tiền khi exact decimal arithmetic cần được bảo toàn:

```go
// Sai với exact money semantics: binary floating point có rounding.
total := 0.1 + 0.2
```

`int64` cũng không tự giải quyết mọi thứ:

- Currency có số decimal digits khác nhau.
- Scale phải là contract rõ.
- Range hữu hạn và có overflow.
- FX rate/tax có rounding mode.
- Allocation remainder cần policy.

## Equality Theo Value

```go
func (m Money) Equal(other Money) bool {
	return m.amount == other.amount && m.currency == other.currency
}
```

`500.000 VND` không bằng `500.000 USD`. Equality phải gồm mọi attribute cấu thành meaning.

Vì fields hiện tại comparable, package domain có thể dùng `m == other` nội bộ. Method `Equal` công bố semantics ổn định và không expose representation. Nếu sau này Money có field không comparable, caller không bị ép dùng reflection.

## Immutable-style Arithmetic

```go
func (m Money) Add(other Money) (Money, error) {
	if err := m.ensureSameCurrency(other); err != nil {
		return Money{}, err
	}

	amount, err := checkedAdd(m.amount, other.amount)
	if err != nil {
		return Money{}, err
	}
	return Money{amount: amount, currency: m.currency}, nil
}
```

Method không mutate receiver:

```go
original := MustMoney(100_000, "VND")
sum, _ := original.Add(MustMoney(50_000, "VND"))

// original vẫn là 100.000 VND, sum là 150.000 VND.
```

Go không có keyword immutable. Immutable-style đến từ:

- Fields private.
- Không trả pointer cho mutable internal data.
- Value receiver.
- Operation trả value mới.
- Không chứa slice/map/pointer có aliasing, hoặc phải copy defensively.

## Currency Mismatch

Wrong:

```go
balance := account.BalanceAmount + deposit.Amount
```

Nếu currency tách thành field khác, caller có thể quên check.

Value Object giữ pair:

```go
next, err := balance.Add(deposit)
if errors.Is(err, ErrCurrencyMismatch) {
	// business operation không hợp lệ
}
```

Rule được enforce bất kể caller là deposit, refund hay reconciliation.

## Integer Overflow

Go integer arithmetic có thể wrap theo two's-complement result; business không được để balance âm/dương bất ngờ vì overflow.

```go
func checkedAdd(left, right int64) (int64, error) {
	if right > 0 && left > math.MaxInt64-right {
		return 0, ErrMoneyOverflow
	}
	if right < 0 && left < math.MinInt64-right {
		return 0, ErrMoneyOverflow
	}
	return left + right, nil
}
```

Overflow có thể hiếm với VND, nhưng Value Object là đúng nơi tập trung guarantee. Production policy có thể dùng decimal library, arbitrary precision hoặc database numeric; mapping/range contract phải nhất quán.

## Zero Value Trong Go

Go khuyến khích useful zero value khi phù hợp. Với `Money`, zero-value struct là:

```go
Money{amount: 0, currency: ""}
```

Nó không có currency nên invalid. Có ba hướng:

1. Chấp nhận zero value invalid và validate ở constructors/boundaries, như mini-banking.
2. Thiết kế type sao cho zero value có semantics hợp lệ.
3. Dùng pointer/option để biểu diễn absence, nhưng thêm nil handling.

Không ép mọi domain type có useful zero value nếu điều đó tạo meaning giả. Quan trọng là API và tests nói rõ.

## Value Object Chứa Slice/Map

Đoạn này không immutable dù field private:

```go
type Tags struct {
	values []string
}

func NewTags(values []string) Tags {
	return Tags{values: values} // caller vẫn giữ backing array
}
```

Correct defensive copy:

```go
func NewTags(values []string) Tags {
	copyOfValues := append([]string(nil), values...)
	return Tags{values: copyOfValues}
}

func (t Tags) Values() []string {
	return append([]string(nil), t.values...)
}
```

Immutability là ownership của data, không chỉ là viết value receiver.

## Mapping Qua Boundary

HTTP DTO gửi primitive:

```json
{"amount": 500000, "currency": "VND"}
```

Application/domain boundary parse:

```go
amount, err := domain.NewPositiveMoney(cmd.Amount, cmd.Currency)
```

Database adapter scan primitive rồi rehydrate `Money`. Kafka event có schema riêng. Không serialize private field trực tiếp bằng cách thêm JSON tags vào domain chỉ để tránh mapper.

## Khi Nào Không Nên Tạo Value Object?

- Primitive không có rule/meaning riêng.
- Type chỉ được dùng một chỗ và wrapper làm API khó đọc hơn.
- Domain là CRUD nhỏ, coupling đã được chấp nhận.
- Team chưa biết semantics và type sớm sẽ khóa assumption sai.

Một `FirstName`, `LastName`, `Street`, `City` type riêng cho mọi string có thể gây noise nếu không có behavior. Tạo Value Object khi nó gom concept, validation hoặc operation có giá trị.

## Testing Strategy

Test Value Object theo property và edge:

- Normalize input.
- Reject invalid construction.
- Equality/reflexivity và inequality đúng.
- Operation không mutate operands.
- Boundary values và overflow.
- Currency mismatch.
- Zero-value behavior.

Full tests: [`money_test.go`](../examples/mini-banking/internal/account/domain/money_test.go).

## Mastery Questions

1. Vì sao `500000` không đủ domain meaning?
2. Value equality khác Entity identity equality thế nào?
3. Fields private có đủ để một Value Object chứa slice trở thành immutable không?
4. Vì sao `int64` tốt hơn `float64` cho flow này nhưng chưa phải money solution hoàn chỉnh?
5. Supported currency list nên nằm trong `NewCurrency` hay policy khác? Dữ kiện nào quyết định?
6. Zero value invalid có trái idiomatic Go không? Bạn quản lý nó thế nào?
7. Khi nào wrapper type chỉ tạo ceremony?
