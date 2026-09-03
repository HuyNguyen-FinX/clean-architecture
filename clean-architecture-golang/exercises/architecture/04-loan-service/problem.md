# Bài 4: Thiết Kế Loan Service

## Requirement

Loan Service:

- Submit loan application.
- Check eligibility.
- Score risk.
- Approve/reject loan.
- Disburse loan.

## Nhiệm vụ

1. Xác định Bounded Context.
2. Thiết kế domain model.
3. Thiết kế external gateway tới credit scoring/core banking.
4. Phân tích rule nào thuộc domain, rule nào thuộc provider.
5. Thiết kế audit trail.

## Câu hỏi

- Risk score có phải Value Object không?
- Approval rule nằm trong domain hay application?
- Disbursement failure xử lý thế nào?

## Bối Cảnh Bổ Sung

Loan application sống nhiều ngày, có document verification, automated scoring, manual approver và core banking disbursement. Policy/model thay đổi theo version; sáu tháng sau auditor phải tái hiện được decision cũ.

Bạn có thể chọn modular monolith hoặc services. Bounded Context không mặc định đồng nghĩa deploy riêng.

## Failure Injection

- Scoring provider trả reason code chưa biết.
- Hai approver quyết định cùng lúc với authority khác nhau.
- Offer hết hạn đúng lúc applicant accept.
- Core banking timeout sau khi có thể đã chuyển tiền.
- Decision commit nhưng Kafka down.

## Deliverables

1. Context map và ubiquitous language cho Application/Decision/Loan.
2. Aggregate lifecycle và authority invariants.
3. Policy/provider ownership decision kèm anti-corruption mapping.
4. Workflow transaction/outbox/reconciliation boundaries.
5. Audit evidence schema, PII access/redaction/retention notes.
6. Tests gồm historical decision replay và manual path.
7. SLO cho pending/manual/disbursement unknown.

## Self-review

- `RiskScore` có scale/source/version semantics hay chỉ là `int`?
- Technical log có đang bị dùng thay business audit không?
- Disbursement timeout có bị đánh dấu failed quá sớm không?
- Model có trộn origination với servicing/collection không?
