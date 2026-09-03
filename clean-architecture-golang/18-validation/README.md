# Validation: syntax, workflow và invariant không phải một thứ

Validation bị đặt sai chỗ khi team gom mọi rule vào request tags hoặc lặp cùng rule ở handler, service và Entity. Cần phân loại theo trust boundary và ownership.

## Kết quả học tập

- tách transport, application và domain validation;
- dùng constructor/Value Object chống invalid state;
- xử lý cross-field, temporal và external validation;
- quản duplication có chủ đích;
- test boundary thay vì chỉ test validator library.

## 1. Ba level

### Level 1: trực giác

Mỗi cửa kiểm một loại vé. JSON parser kiểm gói tin đọc được; use case kiểm người gọi/quy trình; domain kiểm business state luôn đúng.

### Level 2: Backend Engineer

~~~text
HTTP: body, type, required, format
Application: command completeness, actor, idempotency
Domain: amount, currency, overdraft, transition
Database: durable constraint
~~~

### Level 3: Architecture

Validation là enforcement của invariants tại nhiều trust boundary. Duplicate check có thể đúng nếu bảo vệ failure khác nhau; nguồn sự thật business vẫn phải ở model sở hữu rule.

## 2. Transport validation

Ví dụ:

- malformed/trailing JSON;
- unknown field theo compatibility policy;
- Content-Type;
- body size;
- string length/UUID syntax;
- required header.

Handler có thể trả field violations nhanh. Nhưng transport validator không bảo vệ Kafka/CLI/internal caller.

## 3. Application validation

Command:

~~~go
if cmd.FromAccountID == cmd.ToAccountID {
	return ErrSameAccountTransfer
}
if !actor.CanTransfer(cmd.FromAccountID) {
	return ErrForbidden
}
~~~

Application kiểm workflow context: actor, tenant, idempotency key, feature availability, command relationships trước expensive I/O khi có thể.

## 4. Domain invariant

~~~go
func (a *Account) Withdraw(amount Money) error {
	if !amount.IsPositive() {
		return ErrInvalidAmount
	}
	next, err := a.balance.Sub(amount)
	if err != nil {
		return err
	}
	if next < -a.overdraftLimit {
		return ErrInsufficientBalance
	}
	a.balance = next
	return nil
}
~~~

Private field và behavior duy trì invariant qua mọi entry point. Constructor/Rehydrate từ chối initial invalid state.

## 5. Value Object chống primitive obsession

~~~go
amount, err := domain.NewPositiveMoney(cmd.Amount, cmd.Currency)
~~~

Sau constructor, use case không lặp amount > 0/currency format. Type làm invalid state khó biểu diễn.

Không biến mọi string thành Value Object nếu không có semantics/validation/reuse; type proliferation cũng có cost.

## 6. Cross-field validation

from khác to, start trước end, currency của hai Money giống nhau. Đặt ở object/use case sở hữu relation. Field tag đơn lẻ không diễn đạt tốt.

## 7. Temporal/external rule

“Account tồn tại” cần Repository; “daily limit chưa vượt” cần state/history. Đây không phải validation thuần DTO. Application load data rồi Domain Service/Aggregate policy quyết định.

Đừng gọi database từ custom JSON validator; I/O ẩn làm timeout/transaction khó quản.

## 8. Database constraint

CHECK balance >= -overdraft bảo vệ durable state trước writer khác. Nó bổ trợ, không thay domain behavior. Unique constraint là nơi cuối cùng xử lý race cho uniqueness; pre-check chỉ cải thiện UX.

## 9. Normalize hay reject?

Currency có thể TrimSpace/Uppercase ở constructor. Email normalization phức tạp hơn. Quyết định phải explicit:

- normalize trước equality/idempotency hash;
- lưu canonical và có thể raw nếu audit cần;
- không silently thay semantic input;
- test Unicode/locale nếu domain liên quan.

## 10. Error shape

Transport có thể trả:

~~~json
{
  "code": "invalid_request",
  "violations": [
    {"field": "amount", "reason": "must_be_positive"}
  ]
}
~~~

Không buộc domain ValidationError phải chứa JSON field name. Adapter map domain/application semantics sang external fields.

## 11. Duplication có chủ đích

Amount > 0 có thể check ở handler và Money constructor:

- handler cho feedback sớm;
- domain là source of truth.

Rủi ro drift nếu threshold business đổi. Rule phức tạp chỉ nên gọi một policy/domain method; delivery không copy.

## 12. Testing

- decoder tests cho malformed/limits;
- command tests cho cross-field/actor;
- domain table tests cho boundary values;
- property/fuzz tests cho parser/Value Object;
- DB integration tests cho constraints/races;
- cross-entry test để HTTP/Kafka cùng outcome.

Test boundary: zero, negative, max int overflow, currency mismatch, frozen state, exact overdraft limit.

## 13. Production failure

Hai request cùng pre-check “username chưa tồn tại”, rồi insert. Chỉ unique constraint giải quyết race. Adapter map SQLSTATE unique violation thành conflict. Validation query không tạo atomicity.

## 14. Debug

Khi HTTP accept nhưng Kafka reject:

1. so raw DTO/message;
2. so normalization;
3. so command constructor;
4. tìm duplicated rule;
5. xác định source of truth;
6. thêm shared application/domain test, không shared transport type.

## 15. Khi nào validation đơn giản đủ?

CRUD nội bộ có thể dùng decoder + vài if. Không cần validator framework. Framework hữu ích cho nhiều wire fields/i18n, nhưng domain invariant vẫn ở domain.

## 16. Bài tập

1. Phân loại 20 rule của transfer theo boundary.
2. Fuzz NewCurrency.
3. Tái hiện uniqueness race.
4. Thiết kế violation response không leak internal field.
5. Xóa duplicated overdraft rule khỏi handler.

## 17. Mastery questions

1. Required amount khác insufficient balance thế nào?
2. Vì sao pre-check uniqueness không đủ?
3. Khi nào duplicate validation đúng?
4. context có nên đi vào Money constructor không?
5. Rehydrate vì sao vẫn validate?
6. Validator gọi DB tạo hidden cost gì?

## Further reading

- Go fuzzing documentation.
- OWASP Input Validation Cheat Sheet.
- DDD literature về Value Object và invariant.
- PostgreSQL constraint documentation.

## Quality gate

- [x] Ba validation boundaries + database
- [x] Go domain/command examples
- [x] Cross-field/external/normalization
- [x] Race/failure/debug/tests
- [x] Trade-off, exercises, mastery
