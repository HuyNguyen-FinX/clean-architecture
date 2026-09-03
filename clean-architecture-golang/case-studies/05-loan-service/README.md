# Case Study 05: Loan Service - Bounded Context, Decision Và Audit

Loan lifecycle kéo dài từ vài phút đến nhiều ngày, đi qua scoring provider, manual review, document verification và core banking. Vì vậy đây là bài toán modeling ngôn ngữ và workflow trước khi là bài toán folder.

## Scope Và Ubiquitous Language

Các thuật ngữ phải được thống nhất với business:

- `Application`: hồ sơ xin vay, không phải application layer.
- `Applicant`: người/tổ chức nộp hồ sơ.
- `Offer`: điều khoản được đề nghị, có thời hạn.
- `Decision`: approved, rejected hoặc manual review kèm reason/evidence.
- `Facility/Loan`: khoản vay tồn tại sau khi giải ngân.
- `Disbursement`: operation chuyển tiền, có thể pending/unknown.

Một model duy nhất tên `Loan` từ lúc draft đến collection thường trộn các lifecycle và team ownership khác nhau.

## Bounded Context Đề Xuất

~~~mermaid
flowchart LR
    ORIG["Origination\napplication + offer"] --> RISK["Risk decision"]
    ORIG --> DOC["Document verification"]
    ORIG --> DISB["Disbursement"]
    DISB --> CORE["Core lending/accounting"]
    CORE --> COLL["Servicing + collection"]
~~~

Ban đầu các context có thể là module trong modular monolith. Boundary ngôn ngữ/data vẫn có ích mà chưa phải trả chi phí network. Chỉ tách service khi ownership, scaling, release cadence hoặc isolation thực sự khác.

## Aggregate Application

Application Root bảo vệ lifecycle cục bộ:

~~~text
Draft -> Submitted -> UnderReview -> Approved -> OfferAccepted
                    |             |
                    v             v
                 Rejected      Expired
                                      \
                                       -> DisbursementPending -> Disbursed
~~~

Invariant ví dụ:

- Chỉ owner được sửa Draft.
- Submit cần mandatory facts/document references.
- Decision gắn với policy version và evidence snapshot.
- Approver không được duyệt vượt authority limit.
- Offer accepted phải còn hiệu lực và không đổi terms.
- Disbursement chỉ bắt đầu một lần cho accepted offer.

~~~go
func (a *Application) ApplyDecision(d Decision, decidedAt time.Time) error {
	if a.status != UnderReview {
		return ErrInvalidTransition
	}
	if d.PolicyVersion == "" || len(d.Reasons) == 0 {
		return ErrDecisionWithoutEvidence
	}
	a.decision = d
	if d.Outcome == Approved {
		a.status = Approved
	} else if d.Outcome == Rejected {
		a.status = Rejected
	} else {
		a.status = ManualReview
	}
	return nil
}
~~~

## Risk Score Không Đồng Nghĩa Decision

Provider có thể trả số `742`, grade `B`, hoặc reason codes. Adapter map dữ liệu provider thành result nội bộ có provenance; Policy domain dùng facts + score + version để tạo Decision. Nếu provider đã là owner của quyết định regulatory, local domain không giả vờ tái hiện logic đó; nó lưu decision và evidence được trả về.

`RiskScore` có thể là Value Object nếu có range, scale, version và comparison semantics ổn định. Một `int` không mang đủ meaning khi model/provider đổi scale.

## Ports Và Compile-Time Dependency

~~~go
type RiskAssessor interface {
	Assess(ctx context.Context, facts RiskFacts, operationID string) (Assessment, error)
}

type DisbursementGateway interface {
	Submit(ctx context.Context, instruction Instruction) (Submission, error)
	Lookup(ctx context.Context, operationID string) (DisbursementStatus, error)
}
~~~

Application sở hữu ports vì use cases tiêu thụ capability. Provider adapters import application contract và SDK. Domain policy nhận `Assessment`/facts thuần, không nhận provider JSON hay `context.Context`.

## Workflow Và Transaction Boundary

Submit Application là local transaction: validate transition, snapshot facts, lưu aggregate, ghi outbox `ApplicationSubmitted`. Risk worker xử lý sau. Không giữ DB transaction trong lúc scoring API chạy.

Decision flow:

1. Inbox claim event theo message ID.
2. Load Application và current version.
3. Nếu chưa có assessment, gọi provider ngoài DB transaction bằng stable operation ID.
4. Trong transaction ngắn, re-load/lock Application, kiểm tra state, apply Decision.
5. Save Decision, audit record và outbox atomically.

Giữa bước 3 và 4 process có thể crash; operation ID cho phép reuse/lookup assessment. Facts hash ngăn dùng assessment của input khác.

## Disbursement Và Ambiguous Outcome

Core banking có thể accept transfer rồi connection timeout. Không mark `Failed` chỉ từ timeout. State `DisbursementUnknown` kích hoạt reconciliation; cùng instruction ID được lookup hoặc submit idempotently.

Compensation không đơn giản là "undo loan": tiền có thể đã vào tài khoản. Business cần reversal/manual process được audit. Saga chỉ phối hợp state; nó không xóa thực tế đã xảy ra.

## Persistence Và Audit

- Application current state dùng optimistic version.
- Decision lưu policy/model version, facts hash, reasons, actor/source và timestamp.
- Documents lưu reference/hash/verification result, không nhét binary vào aggregate.
- Workflow operation lưu stable ID, request hash, attempts và external reference.
- Outbox/inbox xử lý event delivery.
- Audit append-only, access-controlled, có retention; log kỹ thuật không thay audit business.

PII phải được phân loại, mã hóa, redact khỏi logs/traces và có deletion/retention policy phù hợp. Việc domain độc lập framework không tự giải quyết compliance.

## Failure Matrix

| Failure | State trung thực | Recovery |
|---|---|---|
| Document thiếu | Draft/Submitted bị từ chối transition | trả requirement cụ thể |
| Scoring unavailable | UnderReview | retry có budget hoặc manual queue |
| Model trả unknown reason code | không tự approve | quarantine/contract alert |
| Hai reviewer quyết định | version conflict; một thắng | reload, audit attempt |
| Offer hết hạn cùng lúc accept | transaction kiểm tra clock/version | reject deterministic |
| Disbursement timeout | Unknown | lookup/reconciliation |
| Kafka down sau decision | decision commit + outbox pending | relay retry |

## Testing Strategy

- Domain transition table cho mọi status và authority boundary.
- Policy tests bằng fixed facts/assessment/policy version; golden cases phải được domain experts duyệt.
- Use-case tests cho stale version, provider unknown, duplicate command và audit/outbox atomicity.
- Provider contract tests với sandbox/recorded sanitized fixtures, unknown enum, timeout.
- PostgreSQL integration cho unique operation, optimistic conflict, rollback.
- Replay historical decision set khi đổi policy/model để phát hiện drift; không tự động thay decision cũ.
- Security tests cho authorization và PII redaction.

## Observability Và Human Operations

Metrics: age theo workflow state, approval/rejection theo policy version (không label PII), manual-review backlog, provider latency/error, disbursement unknown age. Trace correlation qua application/operation ID. Alert business như "approved chưa disburse quá SLA" quan trọng hơn chỉ CPU/5xx.

Operator UI phải dùng use case có authorization và reason, không cho sửa DB. Manual override tạo Decision mới hoặc explicit override record; không xóa evidence cũ.

## Trade-off

State machine và audit model tăng ceremony nhưng tương xứng với lifecycle/risk. Với lending product rất đơn giản do một upstream platform quyết định toàn bộ, service có thể chỉ là integration workflow; không cần giả tạo rich domain. Ngược lại, nhét tất cả policy vào một `LoanService` làm rule khó version/test/audit.

## Câu Hỏi Mastery

1. Khi nào `RiskScore` là Value Object, khi nào chỉ là provider data?
2. Policy version phải được lưu ở đâu để tái hiện Decision?
3. Vì sao disbursement timeout không được chuyển sang rejected?
4. Audit log khác structured application log ở trách nhiệm nào?
5. Bounded Context có bắt buộc là microservice không?

## Bài Thực Hành

Viết decision table cho ba loan products và hai approval authority levels. Sau đó thiết kế schema/evidence để sáu tháng sau giải thích được vì sao một hồ sơ được approve mà không cần chạy lại model hiện tại.
