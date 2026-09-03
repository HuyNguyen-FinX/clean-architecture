# Refactoring tới Clean Architecture: đổi dependency mà không rewrite

Hệ thống thật hiếm khi bắt đầu sạch. Refactoring kiến trúc là kỹ năng giảm risk theo vertical slice: khóa behavior, lộ rule, tạo boundary tại điểm đau và giữ deployment reversible.

## Kết quả học tập

- refactor God Handler qua Step 0-8;
- dùng characterization test;
- phân biệt move file với invert dependency;
- tạo seam quanh database/event;
- rollout dual path có đo lường;
- tránh big-bang rewrite.

## 1. Step 0: current code

~~~go
func Transfer(w http.ResponseWriter, r *http.Request) {
	var req transferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", 400)
		return
	}
	tx, _ := db.BeginTx(r.Context(), nil)
	var balance int64
	_ = tx.QueryRowContext(r.Context(),
		"SELECT balance FROM accounts WHERE id=$1", req.From).Scan(&balance)
	if balance < req.Amount {
		http.Error(w, "insufficient", 409)
		return
	}
	_, _ = tx.ExecContext(r.Context(), "UPDATE accounts ...")
	_ = kafkaWriter.WriteMessages(r.Context(), kafka.Message{Value: []byte("transferred")})
	_ = tx.Commit()
	w.WriteHeader(201)
}
~~~

Nó có lỗi resource/error/dual-write ngoài vấn đề structure. Không refactor mù mà thay behavior production không biết.

## 2. Ba level

### Level 1

Tách từng responsibility và giữ chương trình chạy sau mỗi bước.

### Level 2

Dùng tests/seams/adapters, commit nhỏ, observability/rollout và rollback plan.

### Level 3

Refactor architecture đổi ownership/source dependency theo change coupling. Mục tiêu là giảm cost/risk, không đạt folder tree lý tưởng.

## 3. Inventory behavior và risk

Trước code:

- current statuses/error bodies;
- transaction semantics thực tế;
- idempotency/retry;
- external side effects;
- callers;
- SLO/traffic;
- known bugs có cần preserve?

Phân biệt characterization “code đang làm” và desired spec. Known bug được test mô tả rồi đổi trong commit behavior riêng.

## 4. Step 1: characterization tests

~~~go
func TestTransferInsufficientDoesNotUpdate(t *testing.T) {
	// HTTP request, fake/real current dependencies
	// assert 409 and no durable balance change
}
~~~

Nếu function hard-coded globals, tạo seam nhỏ: package variables/function parameters hoặc httptest DB boundary. Không cần test từng private line.

Golden master hữu ích cho legacy output lớn, nhưng sanitize nondeterminism và đừng coi bug là spec vĩnh viễn.

## 5. Step 2: identify business rule

Tìm if/math/state transitions:

~~~text
amount > 0
currency match
balance after withdraw >= -overdraft
frozen source cannot withdraw
~~~

Đặt language trước folder. Đây là rules cần sống qua HTTP/gRPC/Kafka.

## 6. Step 3: extract Account behavior

~~~go
func (a *Account) Withdraw(amount Money) error {
	if a.status == Frozen {
		return ErrAccountFrozen
	}
	next, err := a.balance.Sub(amount)
	if err != nil {
		return err
	}
	if next.LessThan(a.overdraftLimit.Negate()) {
		return ErrInsufficientBalance
	}
	a.balance = next
	return nil
}
~~~

Viết domain tests trước khi handler gọi method mới. Giữ adapter mapping tạm thời.

Benefit: rule test nhanh/reuse. Risk: domain model khác legacy null/currency; mapper phải fail rõ.

## 7. Step 4: extract Use Case

~~~go
type TransferMoney struct {
	accounts AccountRepository
	tx       Transactor
}

func (uc *TransferMoney) Execute(ctx context.Context, cmd Command) error {
	return uc.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		// load, domain behavior, save
	})
}
~~~

Handler cũ map request sang command và gọi use case. Chưa cần thay router/framework.

## 8. Step 5: invert Repository dependency

Trước:

~~~text
use case → pgxpool
~~~

Sau:

~~~text
use case → AccountRepository port
postgres adapter → AccountRepository port
~~~

Chỉ chuyển file repository ra folder khác mà interface vẫn trả postgres.Row không đổi semantic dependency.

Interface nhỏ theo consumer. Có thể tạo adapter wrapping legacy DB helper trước, rồi rewrite SQL sau.

## 9. Step 6: move transaction boundary

Legacy mỗi Repository commit riêng. Đưa boundary về use case qua Transactor/UoW và viết integration rollback test:

~~~text
save A success
save B injected failure
reload A/B unchanged
~~~

Không đổi transaction + domain extraction + schema trong một commit nếu có thể chia.

## 10. Step 7: isolate HTTP DTO/error

Handler:

~~~text
strict parse/limits
→ DTO → command
→ use case
→ stable public response
~~~

Domain error bỏ HTTP status. Có thể giữ old response compatibility bằng mapper.

## 11. Step 8: replace Kafka dual write bằng outbox

Không chỉ tạo EventPublisher port rồi publish sau commit; dual-write vẫn còn. Trong DB transaction thêm Transfer + outbox. Worker publish/mark, consumer idempotent.

Rollout:

1. migration outbox additive;
2. write outbox disabled/shadow;
3. deploy worker dry-run/metrics;
4. enable publish;
5. compare old/new events;
6. remove direct publish.

## 12. Dependency progression

~~~mermaid
flowchart LR
    H0["God Handler"] --> DB0["DB"]
    H0 --> K0["Kafka"]
    H1["HTTP Adapter"] --> UC["Use Case"]
    UC --> D["Domain"]
    UC --> PORT["Ports"]
    PG["Postgres Adapter"] -.implements.-> PORT
    OUT["Outbox Worker"] --> K1["Kafka"]
~~~

Refactor hoàn tất khi source dependency/ownership đổi, không chỉ khi số file tăng.

## 13. Strangler inside a monolith

Route/use case mới chạy cạnh legacy:

- feature flag theo tenant;
- shadow compare read-only result;
- canary percentage;
- old path fallback chỉ nếu preserve guarantee;
- metrics business invariant.

Không dual-write hai sources mà không reconciliation.

## 14. Database evolution

Nếu domain model cần new fields:

- expand schema;
- mapper hỗ trợ old/new;
- backfill chunk;
- activate invariant;
- contract/production metrics;
- contract later.

Code rollback phải đọc schema mới. Data migration thường không rollback đơn giản.

## 15. Test progression

1. characterization HTTP/E2E;
2. domain tests;
3. use-case fake;
4. Repository contract;
5. Postgres integration transaction;
6. HTTP contract;
7. outbox/consumer;
8. production canary invariants.

Khi lower tests đủ, giữ vài characterization high-level, không nhất thiết giữ mọi legacy snapshot.

## 16. Observability trước rollout

Đo old/new:

- outcome/status;
- balance/ledger reconciliation;
- latency/DB locks;
- idempotency duplicates;
- outbox lag;
- errors by category.

“No exception” không chứng minh money correct.

## 17. Common traps

### Abstract trước khi hiểu

Generic Repository/ServiceFactory làm legacy khó hơn.

### Big-bang package move

Merge conflicts, behavior diff lẫn movement. Tách mechanical move khỏi semantic change.

### Preserve bug vô thời hạn

Characterization là safety net, không đạo luật. Mark test/issue và change intentionally.

### New architecture gọi legacy God Service

Wrapper chỉ đổi tên. Có thể là migration seam tạm, phải có exit criterion.

### Rewrite

Mất edge cases/operational knowledge và kéo dài thời gian không deliver. Rewrite chỉ cân nhắc khi component nhỏ, spec/test rõ, migration/rollback khả thi.

## 18. Production scenario

Refactor transfer ở 5.000 TPS:

- canary 1% theo stable key;
- both paths write same DB? tránh duplicate;
- compare outcome via shadow domain calculation, không shadow side effect;
- monitor lock/latency/reconciliation;
- kill switch;
- expand outbox capacity trước enable;
- rollback app vẫn hiểu schema.

## 19. Debug regression

1. classify mapping/domain/transaction/event difference;
2. replay same command/idempotency key in isolated env;
3. compare old/new timeline;
4. inspect DB writes/locks/outbox;
5. bisect small commits;
6. revert feature flag/path, không destructive data rollback;
7. add regression at lowest boundary.

## 20. Khi nào chấp nhận architecture cũ?

Stable low-risk code ít thay đổi, coverage yếu và refactor không mở business value có thể để nguyên. Bọc integration boundary, document risk, tập trung hotspots. Cleanliness không tự hoàn vốn.

## 21. Lab

Làm [Lab 11: Refactoring](../labs/lab-11-refactoring/README.md): executable God Handler → domain/usecase/adapters theo commits/steps.

## 22. Mastery questions

1. Move file khác inversion thế nào?
2. Characterization test xử lý known bug ra sao?
3. Vì sao event port chưa giải dual write?
4. Step nào cho value sớm nhất?
5. Rollback app/schema khác data rollback?
6. Shadow traffic không được thực hiện side effect gì?
7. Legacy wrapper khi nào chấp nhận?
8. Metric nào chứng minh transfer refactor đúng?

## Further reading

- Martin Fowler, Refactoring và Strangler Fig.
- Michael Feathers, Working Effectively with Legacy Code.
- Branch by Abstraction và expand-contract migration literature.
- [Code Review Exercise](../code-review-exercises/01-god-handler/problem.md).

## Quality gate

- [x] Step 0-8 code/dependency progression
- [x] Characterization/seams/testing
- [x] Transaction/outbox/database rollout
- [x] Strangler/canary/observability/rollback
- [x] Traps, production, debug, trade-off
- [x] Lab/mastery/references
